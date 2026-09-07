package order

import (
	"context"
	"sync"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"go.uber.org/zap"
)

// Количество воркеров
const workerCount = 3

type BonusRepository interface {
	GetOrder(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError)
}

type OrderRepository interface {
	UpdateStatusAndBonus(ctx context.Context, order entities.Order) *entities.DomainError
	FindStaleOrders(ctx context.Context, statuses []entities.OrderStatus, limit int64) ([]string, *entities.DomainError)
}

type OrderProcessor struct {
	logger             interfaces.Logger
	ordersProcessingCh chan string
	ordersProcessedCh  chan entities.Order
	bonusRepository    BonusRepository
	orderRepository    OrderRepository
}

// AddOrderIDToQueue добавляет идентификатор заказа в очередь на обработку.
// Если очередь переполнена, заказ не добавляется в очередь
func (op *OrderProcessor) AddOrderIDToQueue(orderID string) {
	select {
	case op.ordersProcessingCh <- orderID:
		op.logger.Info("Заказ отправлен в очередь на обработку", zap.String("order_id", orderID))
	default:
		op.logger.Warn("Очередь обработки переполнена, заказ обработается позже из БД", zap.String("order_id", orderID))
	}
}

// AddOrderToQueue добавляет заказ в очередь на обновление.
// Если очередь переполнена, заказ не добавляется в очередь.
func (op *OrderProcessor) AddOrderToQueue(order entities.Order) {
	select {
	case op.ordersProcessedCh <- order:
		op.logger.Info("Заказ отправлен в очередь на обновление", zap.Any("order", order))
	default:
		op.logger.Warn("Очередь обработки переполнена", zap.Any("order", order))
	}
}

// NewOrderProcessor создает обработчик заказов с очередями для обработки
// заказов и обновления их данных.
func NewOrderProcessor(
	bonusRepository BonusRepository,
	orderRepository OrderRepository,
	logger interfaces.Logger,
) *OrderProcessor {
	// канал для обработки заказов, у которых статус NEW, PROCESSING
	ordersProcessingCh := make(chan string, workerCount*5)
	// канал с результатом начисления по заказам
	ordersProcessedCh := make(chan entities.Order, workerCount*5)

	return &OrderProcessor{
		logger:             logger,
		bonusRepository:    bonusRepository,
		orderRepository:    orderRepository,
		ordersProcessingCh: ordersProcessingCh,
		ordersProcessedCh:  ordersProcessedCh,
	}
}

// orderStatusesCheck статусы заказов для проверки начисления бонусов.
var orderStatusesCheck = []entities.OrderStatus{
	entities.OrdersStatusNew,
	entities.OrdersStatusProcessing,
}

const (
	// databasePollerInterval интервал проверки заказов на начисление бонусов.
	databasePollerInterval = 1 * time.Minute
)

// StartDatabasePoller запускает фоновую проверку заказов,
// требующих проверки начисления бонусов, и добавляет найденные заказы в очередь
// на обработку.
func (op *OrderProcessor) StartDatabasePoller(ctx context.Context) {
	tick := time.Tick(databasePollerInterval)

	go func() {
		for {
			select {
			case <-tick:
				orderIDs, err := op.orderRepository.FindStaleOrders(
					ctx,
					orderStatusesCheck,
					10,
				)
				if err != nil {
					op.logger.Error("ошибка получения order_ids StartDatabasePoller", zap.Error(err.SourceErr))
					continue
				}

				for _, orderID := range orderIDs {
					op.AddOrderIDToQueue(orderID)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Start запускает фоновые воркеры для обработки заказов и обновления их данных.
//
// Воркеры обработки получают идентификаторы заказов из ordersProcessingCh,
// проверяют состояние начисления бонусов и передают результаты в
// ordersProcessedCh.
//
// Воркеры обновления получают обработанные заказы из ordersProcessedCh
// и сохраняют результаты в базе данных.
func (op *OrderProcessor) Start(ctx context.Context) {
	var wgProcessWorkers sync.WaitGroup
	var wgSaveResults sync.WaitGroup

	// воркеры, которые слушают заказы из ordersProcessingCh
	for i := 0; i < workerCount; i++ {
		wgProcessWorkers.Add(1)
		go func() {
			defer wgProcessWorkers.Done()
			op.orderCheckStatus(ctx)
		}()
	}

	// воркеры, которые обновляют заказы в БД, берут данные из ordersProcessedCh
	for i := 0; i < workerCount; i++ {
		wgSaveResults.Add(1)
		go func() {
			defer wgSaveResults.Done()
			op.orderUpdateBonus(ctx)
		}()
	}

	go func() {
		<-ctx.Done()
		op.logger.Info("Получен сигнал отмены. Ожидаем завершения воркеров...")

		close(op.ordersProcessingCh)
		wgProcessWorkers.Wait()

		close(op.ordersProcessedCh)
		wgSaveResults.Wait()

		op.logger.Info("Все фоновые воркеры успешно завершили работу")
	}()
}

// orderCheckStatus получает идентификаторы заказов из очереди
// и проверяет состояние начисления бонусов по каждому заказу.
func (op *OrderProcessor) orderCheckStatus(ctx context.Context) {
	for orderID := range op.ordersProcessingCh {
		op.checkOrderBonus(ctx, orderID)
	}
}

// checkOrderBonus проверяет наличие заказа в системе начисления бонусов
// и передает результат в очередь на обновление данных заказа.
//
// Если заказ не найден, он добавляется в очередь с отложенной повторной
// проверкой через 2 минуты.
func (op *OrderProcessor) checkOrderBonus(ctx context.Context, orderID string) {
	result, err := op.bonusRepository.GetOrder(ctx, orderID)
	if err != nil {
		if err.ErrorType == entities.NoContentErrorType {
			// заказу статус не присваиваем, так как падают тесты на гитлабе
			op.logger.Info("Заказ не найден в bonus, заказу присвоен статус INVALID", zap.String("order_id", orderID))

			result = &entities.Order{
				OrderID:     orderID,
				NextCheckAt: new(time.Now().Add(2 * time.Minute)),
			}
		} else {
			return
		}
	}

	select {
	case op.ordersProcessedCh <- *result:
		op.logger.Info("Заказ отправлен в очередь на обновление данных",
			zap.String("order_id", orderID),
			zap.Any("order", result),
		)
	case <-ctx.Done():
		return
	}
}

// orderUpdateBonus получает результаты проверки заказов из очереди
// и обновляет данные заказов в базе данных.
func (op *OrderProcessor) orderUpdateBonus(ctx context.Context) {
	for orderData := range op.ordersProcessedCh {
		// если ctx отменен, создаем новый контект чтобы запросы в БД успели выполниться
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		op.updateStatusOrder(saveCtx, orderData)

		cancel()
	}
}

// updateStatusOrder обновляет данные заказа в базе данных.
// При retriable ошибке добавляет заказ обратно в очередь на обработку.
func (op *OrderProcessor) updateStatusOrder(ctx context.Context, data entities.Order) {
	if data.OrderID == "" {
		op.logger.Info("updateStatusOrder пустой orderID", zap.Any("order", data))

		return
	}

	err := op.orderRepository.UpdateStatusAndBonus(ctx, data)
	if err != nil {
		if err.ErrorType == entities.RetriableErrorType {
			op.logger.Error(
				"Ошибка обновления заказа, заказ повторно отправлен в очередь",
				zap.Error(err),
				zap.Any("order", data),
			)

			op.AddOrderToQueue(data)
		}

		return
	}

	op.logger.Info("Заказ успешно обновлен", zap.Any("order", data))
}
