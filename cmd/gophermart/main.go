package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
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
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := initConfig()

	initLogger(&cfg)
	defer syncLogger(cfg.logger)

	db := initDatabase(cfg)
	defer closeDatabase(db, cfg.logger)

	if cfg.DatabaseDSN != "" {
		if err := dbpkg.RunMigrations(db); err != nil {
			cfg.logger.Fatal("Migration failed:", zap.Error(err))
		}
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// Запускаем pprof сервер с graceful shutdown
	runPprofServer(ctx, cfg.logger, g)

	wrappedDB := dbPkg.NewSQLDBWrapper(db)
	components := initComponents(cfg, wrappedDB)

	// Запускаем фоновые процессоры
	components.OrderProcessor.Start(ctx, g)
	components.OrderProcessor.StartDatabasePoller(ctx, g)

	startServer(ctx, cfg, components, g)

	if err := g.Wait(); err != nil {
		cfg.logger.Error("Ошибка завершения фоновых процессов", zap.Error(err))
	}

	cfg.logger.Info("Программа завершена")
}

func initDatabase(cfg config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		cfg.logger.Fatal("Ошибка инициализации базы данных", zap.Error(err))
	}

	return db
}

func closeDatabase(db *sql.DB, logger *zap.Logger) {
	if err := db.Close(); err != nil {
		logger.Error("Ошибка закрытия базы данных", zap.Error(err))
	}
}

func syncLogger(logger *zap.Logger) {
	if err := logger.Sync(); err != nil {
		log.Printf("failed to sync logger: %v", err)
	}
}

func initLogger(cfg *config) {
	var err error

	cfg.logger, err = zap.NewProduction(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatal("init logger", err)
	}
}

func initConfig() config {
	cfg, err := readConfig()
	if err != nil {
		log.Fatal("init config", err)
	}

	if cfg.SecretKey == "" {
		log.Fatal("missing secret key")
	}

	return cfg
}

func runPprofServer(ctx context.Context, logger *zap.Logger, g *errgroup.Group) {
	pprofServer := &http.Server{
		Addr:         "localhost:6060",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		logger.Info("pprof server started",
			zap.String("url", "http://localhost:6060/debug/pprof/"),
		)
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("Остановка pprof сервера...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := pprofServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		logger.Info("pprof сервер остановлен")

		return nil
	})
}

func startServer(ctx context.Context, cfg config, components *AppComponents, g *errgroup.Group) {
	router := setupRouter(components, cfg.logger)

	server := &http.Server{
		Addr:    cfg.ServerAddr.String(),
		Handler: router,
	}

	g.Go(func() error {
		cfg.logger.Info("Сервер запущен", zap.String("addr", cfg.ServerAddr.String()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		cfg.logger.Info("Остановка сервера...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		cfg.logger.Info("Сервер остановлен")

		return nil
	})
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
		cfg.logger.Fatal("Ошибка инициализации репозитория бонусов", zap.Error(err))
	}

	userRepo := user.NewRepository(db, cfg.logger)
	orderRepo := order.NewRepository(db)

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
