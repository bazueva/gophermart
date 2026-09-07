package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	dbpkg "github.com/bazueva/gofermart/db"
	"github.com/bazueva/gofermart/internal/app"
	handlerPkg "github.com/bazueva/gofermart/internal/handler"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/middleware"
	"github.com/bazueva/gofermart/internal/repository/bonus"
	dbPkg "github.com/bazueva/gofermart/internal/repository/db"
	"github.com/bazueva/gofermart/internal/repository/db/order"
	"github.com/bazueva/gofermart/internal/repository/db/user"
	orderService "github.com/bazueva/gofermart/internal/service/order"
	userService "github.com/bazueva/gofermart/internal/service/user"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg := initConfig()

	initLogger(&cfg)
	defer syncLogger(cfg.logger)

	db := initDatabase(cfg)
	defer closeDatabase(db)

	if cfg.DatabaseDSN != "" {
		if err := dbpkg.RunMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
	}

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(ctxWithCancel, cancel, cfg.logger)

	// Запускаем pprof сервер с graceful shutdown
	go runPprofServer(ctxWithCancel, cfg.logger)

	wrappedDB := dbPkg.NewSQLDBWrapper(db)
	components := initComponents(cfg, wrappedDB)

	// Запускаем фоновые процессоры
	components.OrderProcessor.Start(ctxWithCancel)
	components.OrderProcessor.StartDatabasePoller(ctxWithCancel)

	startServer(ctxWithCancel, cfg, components)

	// ждем 6 секунд, чтобы дать воркерам orderProcessor (у которых таймаут 5с)
	// гарантированно завершить Graceful Shutdown перед тем, как проверять утечки памяти.
	time.Sleep(6 * time.Second)

	<-ctxWithCancel.Done()
	cfg.logger.Info("Программа завершена")
}

func setupSignalHandler(ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Получен Ctrl+C, останавливаемся...")

		cancel()
	}()
}

func initDatabase(cfg config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	return db
}

func closeDatabase(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}
}

func syncLogger(logger *zap.Logger) {
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()
}

func initLogger(cfg *config) {
	var err error

	cfg.logger, err = zap.NewProduction(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}
}

func initConfig() config {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	return cfg
}

func runPprofServer(ctx context.Context, logger *zap.Logger) {
	pprofServer := &http.Server{
		Addr:         "localhost:6060",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("pprof server started",
			zap.String("url", "http://localhost:6060/debug/pprof/"),
		)
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("pprof server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Остановка pprof сервера...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Ошибка остановки pprof сервера", zap.Error(err))
	} else {
		logger.Info("pprof сервер остановлен")
	}
}

func startServer(ctx context.Context, cfg config, components *AppComponents) {
	router := setupRouter(components, cfg.logger)

	server := &http.Server{
		Addr:    cfg.ServerAddr.String(),
		Handler: router,
	}

	go func() {
		cfg.logger.Info("Сервер запущен", zap.String("addr", cfg.ServerAddr.String()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cfg.logger.Error("Ошибка сервера", zap.Error(err))
		}
	}()

	<-ctx.Done()
	cfg.logger.Info("Остановка сервера...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		cfg.logger.Error("Ошибка остановки сервера", zap.Error(err))
	}
}

func setupRouter(components *AppComponents, logger *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.ServerLogger(logger))
	router.Use(middleware.JSONMiddleware)

	router.Post("/api/user/register", components.Handler.RegisterUser)
	router.Post("/api/user/login", components.Handler.LoginUser)

	router.Group(func(r chi.Router) {
		r.Use(middleware.Authorization(components.App, logger))

		r.Post("/api/user/orders", components.Handler.CreateOrder)
		r.Get("/api/user/orders", components.Handler.UserOrdersList)
		r.Post("/api/user/balance/withdraw", components.Handler.BalanceWithdraw)
		r.Get("/api/user/withdrawals", components.Handler.UserWithdrawals)
		r.Get("/api/user/balance", components.Handler.UserBalance)
	})

	return router
}

type AppComponents struct {
	UserService    *userService.UserService
	OrderService   *orderService.Order
	OrderProcessor *orderService.OrderProcessor
	App            *app.App
	Handler        *handlerPkg.Handler
}

// initComponents инициализирует все компоненты приложения
func initComponents(cfg config, db interfaces.DB) *AppComponents {
	// Репозитории
	bonusRepo, err := bonus.NewRepository(
		cfg.AccrualSystemAddress,
		cfg.logger,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to init bonus repository: %v", err))
	}
	userRepo := user.NewRepository(db, cfg.logger)
	orderRepo := order.NewRepository(db, cfg.logger)

	// Воркеры
	orderProcessor := orderService.NewOrderProcessor(bonusRepo, orderRepo, cfg.logger)

	// Сервисы
	userService := userService.NewUserService(userRepo, cfg.logger, cfg.SecretKey)
	orderService := orderService.NewOrder(orderRepo, orderProcessor, cfg.logger)

	// Приложение и хендлер
	application := app.NewApp(userService, orderService, cfg.logger)
	handler := handlerPkg.NewHandler(cfg.logger, application)

	return &AppComponents{
		UserService:    userService,
		OrderService:   orderService,
		OrderProcessor: orderProcessor,
		App:            application,
		Handler:        handler,
	}
}
