package order

import (
	"context"
	"strings"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces"
	dbPkg "github.com/bazueva/gofermart/internal/repository/db"
	"go.uber.org/zap"
)

type Repository interface {
	FindByOrderID(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError)
	CreateOrder(ctx context.Context, orderID string, userID int32, status entities.OrderStatus) *entities.DomainError
	FindByUserID(ctx context.Context, filter entities.OrderFilter, limit int64, offset int64) ([]entities.Order, *entities.DomainError)
	CountOrdersByUserID(ctx context.Context, orderFilter entities.OrderFilter) (int32, *entities.DomainError)
	UserBalance(ctx context.Context, id int32) (float64, *entities.DomainError)
	CreateOrderWithWithdraw(ctx context.Context, userID int32, orderID string, bonusSum float64) *entities.DomainError
	BeginTransaction(ctx context.Context) (interfaces.Tx, error)
	UserBalanceWithWithdrawn(ctx context.Context, id int32) (entities.Balance, *entities.DomainError)
}

type OrderQueue interface {
	AddOrderIDToQueue(orderID string)
}

// Order Сервис заказов.
type Order struct {
	repository Repository
	logger     interfaces.Logger
	orderQueue OrderQueue
}

// UserBalance возвращает текущий баланс пользователя и общую сумму списанных бонусов.
func (o *Order) UserBalance(ctx context.Context, userID int32) (entities.Balance, *entities.DomainError) {
	return o.repository.UserBalanceWithWithdrawn(ctx, userID)
}

// OrdersWithdrawalsListUser возвращает список заказов пользователя,
// по которым были выполнены списания бонусов, с учетом пагинации.
func (o *Order) OrdersWithdrawalsListUser(
	ctx context.Context,
	userID int32,
	pagination *pagination.Pagination,
) ([]entities.Order, *entities.DomainError) {
	ordersFilter := entities.OrderFilter{UserID: userID, OrderType: new(entities.OrderFilterWriteOffBalanceType)}

	countOrders, err := o.repository.CountOrdersByUserID(ctx, ordersFilter)
	if err != nil {
		return nil, err
	}

	if countOrders == 0 {
		return nil, nil
	}

	pagination.SetTotalCount(int64(countOrders))

	orders, err := o.repository.FindByUserID(
		ctx,
		ordersFilter,
		pagination.GetPerPage(), pagination.GetOffset(),
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

const (
	// Тип блокировки при создании заказа списания бонусов.
	lockTypeCreateOrderWithdraw = 100
)

// BalanceWithdraw выполняет списание бонусов с баланса пользователя
// и создает заказ со списанной суммой.
// Операция выполняется в рамках транзакции с блокировкой на уровне БД
func (o *Order) BalanceWithdraw(ctx context.Context, userID int32, withdraw entities.BalanceWithdraw) *entities.DomainError {
	withdraw.Order = strings.Trim(withdraw.Order, " ")

	errDomain := o.validateOrderID(ctx, withdraw.Order, userID)
	if errDomain != nil {
		if errDomain.ErrorType != entities.InternalServerErrorType {
			return entities.NewBadRequestError(nil, "неверный номер заказа")
		}

		return errDomain
	}

	tx, err := o.repository.BeginTransaction(ctx)
	if err != nil {
		o.logger.Error("ошибка начала транзакции", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}
	defer tx.Rollback()

	ctx = dbPkg.WithTx(ctx, tx)

	if errDomain = o.tryAdvisoryLock(ctx, tx, lockTypeCreateOrderWithdraw, int64(userID)); errDomain != nil {
		return errDomain
	}

	userBalance, errDomain := o.repository.UserBalance(ctx, userID)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.validateWithdraw(userID, userBalance, withdraw)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.repository.CreateOrderWithWithdraw(ctx, userID, withdraw.Order, withdraw.Sum)
	if errDomain != nil {
		return errDomain
	}

	err = tx.Commit()
	if err != nil {
		o.logger.Error("ошибка Commit", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	return nil
}

// tryAdvisoryLock устанавливает advisory-блокировку для указанного ключа.
// Если блокировка уже установлена другой операцией, возвращает ошибку
// о невозможности выполнить операцию параллельно.
func (o *Order) tryAdvisoryLock(
	ctx context.Context,
	tx interfaces.Tx,
	lockType int64,
	key int64,
) *entities.DomainError {
	ok, err := dbPkg.TryLock(
		ctx,
		tx,
		lockType,
		key,
	)
	if err != nil {
		o.logger.Error(
			"ошибка выполнения запроса блокировки",
			zap.Error(err),
			zap.Int64("key", key),
			zap.Int64("lockType", lockType),
		)

		return entities.NewInternalServerError(err, "")
	}

	if !ok {
		o.logger.Warn(
			"запрос отклонен: операция уже выполняется параллельно",
			zap.Int64("key", key),
			zap.Int64("lockType", lockType),
		)

		return entities.NewTooManyRequestError(
			nil,
			"операция уже выполняется, попробуйте позже",
		)
	}

	return nil
}

// OrdersListUser возвращает список заказов пользователя.
func (o *Order) OrdersListUser(
	ctx context.Context,
	userID int32,
	pagination *pagination.Pagination,
) ([]entities.Order, *entities.DomainError) {
	ordersFilter := entities.OrderFilter{UserID: userID, OrderType: new(entities.OrderFilterAddBalanceType)}

	countOrders, err := o.repository.CountOrdersByUserID(ctx, ordersFilter)
	if err != nil {
		return nil, err
	}

	if countOrders == 0 {
		return nil, nil
	}

	pagination.SetTotalCount(int64(countOrders))

	orders, err := o.repository.FindByUserID(
		ctx,
		ordersFilter,
		pagination.GetPerPage(), pagination.GetOffset(),
	)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// CreateOrder регистрирует новый заказ пользователя и отправляет его в очередь
// на дальнейшую обработку.
func (o *Order) CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError {
	orderID = strings.Trim(orderID, " ")

	errDomain := o.validateOrderID(ctx, orderID, userID)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.repository.CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew)
	if errDomain != nil {
		return errDomain
	}

	o.orderQueue.AddOrderIDToQueue(orderID)

	o.logger.Info("Заказ отправлен в очередь на обработку", zap.String("order_id", orderID))

	return nil
}

// validateOrderID валидирует номер заказа.
func (o *Order) validateOrderID(ctx context.Context, id string, userID int32) *entities.DomainError {
	if !helpers.ValidateLuhn(id) {
		return entities.NewUnprocessableEntity(nil, "неверный формат номера заказа")
	}

	ord, err := o.repository.FindByOrderID(ctx, id)
	if err != nil {
		return err
	}

	if ord != nil {
		if ord.UserID == userID {
			return entities.NewOkError(nil, "номер заказа уже был загружен этим пользователем")
		}

		return entities.NewConflictError(nil, "номер заказа уже был загружен другим пользователем")
	}

	return nil
}

// validateWithdraw валидирует сумму для списания бонусов.
func (o *Order) validateWithdraw(userID int32, userBalance float64, withdraw entities.BalanceWithdraw) *entities.DomainError {
	if userBalance <= 0 {
		if userBalance < 0 {
			o.logger.Error("У пользователя отрицательный баланс",
				zap.Int32("user_id", userID),
				zap.Float64("balance", userBalance),
			)
		}

		return entities.NewPaymentRequiredError(nil, "на счету недостаточно средств")
	}

	if withdraw.Sum <= float64(0) {
		return entities.NewBadRequestError(nil, "сумма для списания должна быть больше 0")
	}

	if withdraw.Sum > userBalance {
		return entities.NewPaymentRequiredError(nil, "на счету недостаточно средств")
	}

	return nil
}

// NewOrder создает новый экземпляр сервиса заказов.
func NewOrder(
	repository Repository,
	orderQueue OrderQueue,
	logger interfaces.Logger,
) *Order {
	return &Order{
		repository: repository,
		logger:     logger,
		orderQueue: orderQueue,
	}
}
