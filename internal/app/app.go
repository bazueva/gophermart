package app

import (
	"context"
	"errors"

	contextPkg "github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, user forms.UserForm) (string, *entities.DomainError)
	Login(ctx context.Context, user forms.LoginForm) (string, *entities.DomainError)
	CheckJWTToken(token string) (int32, *entities.DomainError)
}

type OrderService interface {
	CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError
	OrdersListUser(ctx context.Context, userID int32, pagination *pagination.Pagination) ([]entities.Order, *entities.DomainError)
	BalanceWithdraw(ctx context.Context, userID int32, withdraw entities.BalanceWithdraw) *entities.DomainError
	OrdersWithdrawalsListUser(ctx context.Context, userID int32, newPagination *pagination.Pagination) ([]entities.Order, *entities.DomainError)
	UserBalance(ctx context.Context, id int32) (entities.Balance, *entities.DomainError)
}

// App Приложение, управляющее сервисами пользователей и заказов.
type App struct {
	userService  UserService
	orderService OrderService
	logger       interfaces.Logger
}

// UserBalance возвращает баланс пользователя.
func (a *App) UserBalance(ctx context.Context) (entities.Balance, *entities.DomainError) {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return entities.Balance{}, err
	}

	return a.orderService.UserBalance(ctx, userID)
}

// UserWithdrawals возвращает список заказов пользователя со списанием бонусов.
func (a *App) UserWithdrawals(ctx context.Context, page int32, perPage int32) ([]entities.Order, *entities.DomainError) {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return nil, err
	}

	return a.orderService.OrdersWithdrawalsListUser(ctx, userID, pagination.NewPagination(int64(page), int64(perPage)))
}

// BalanceWithDraw списание бонусов пользователя.
func (a *App) BalanceWithDraw(ctx context.Context, request models.BalanceWithdrawRequest) *entities.DomainError {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return err
	}

	return a.orderService.BalanceWithdraw(ctx, userID, entities.BalanceWithdraw{
		Order: request.Order,
		Sum:   request.Sum,
	})
}

// UserOrdersList список заказов пользователя.
func (a *App) UserOrdersList(ctx context.Context, page int32, perPage int32) ([]entities.Order, *entities.DomainError) {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return nil, err
	}

	return a.orderService.OrdersListUser(ctx, userID, pagination.NewPagination(int64(page), int64(perPage)))
}

// userIDFromContext возвращает userID из контекста.
func (a *App) userIDFromContext(ctx context.Context, errorIfEmpty bool) (int32, *entities.DomainError) {
	userID, ok := contextPkg.UserIDFromContext(ctx)
	if !ok {
		err := errors.New("ctx without userID")
		a.logger.Error("ctx without userID", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	if errorIfEmpty && userID == 0 {
		return 0, entities.NewUnauthorizedError(nil, "")
	}

	return userID, nil
}

// CreateOrder создает заказ.
func (a *App) CreateOrder(ctx context.Context, orderID string) *entities.DomainError {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return err
	}

	return a.orderService.CreateOrder(ctx, orderID, userID)
}

// CheckJWTToken проверяет JWT токен.
func (a *App) CheckJWTToken(token string) (int32, *entities.DomainError) {
	return a.userService.CheckJWTToken(token)
}

// Login авторизация пользователя.
func (a *App) Login(ctx context.Context, request models.LoginRequest) (string, *entities.DomainError) {
	return a.userService.Login(ctx, forms.LoginForm{
		Login:    request.Login,
		Password: request.Password,
	})
}

// Register регистрация пользователя.
func (a *App) Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError) {
	return a.userService.Register(ctx, forms.UserForm{
		Login:    request.Login,
		Password: request.Password,
	})
}

// NewApp создает новый экземпляр приложения.
func NewApp(
	userService UserService,
	orderService OrderService,
	logger interfaces.Logger,
) *App {
	return &App{
		userService:  userService,
		orderService: orderService,
		logger:       logger,
	}
}
