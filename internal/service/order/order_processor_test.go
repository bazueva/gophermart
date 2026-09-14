package order

import (
	"context"
	"errors"
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/service/order/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestOrderProcessor_AddOrderIDToQueue(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to queue", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zap.InfoLevel)
		mockLogger := zap.New(core)

		ch := make(chan string, 1)

		op := &OrderProcessor{
			logger:             mockLogger,
			ordersProcessingCh: ch,
		}

		op.AddOrderIDToQueue("123456789")

		assert.Equal(t, "123456789", <-ch)
		assert.Equal(t, "Заказ отправлен в очередь на обработку", logs.All()[0].Message)
	})

	t.Run("error - queue is full", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		ch := make(chan string, 1)
		ch <- "existing-order"

		op := &OrderProcessor{
			logger:             mockLogger,
			ordersProcessingCh: ch,
		}

		op.AddOrderIDToQueue("123456789")

		assert.Equal(t, "existing-order", <-ch)
		assert.Equal(t, "Очередь обработки переполнена, заказ обработается позже из БД", logs.All()[0].Message)
	})
}

func TestOrderProcessor_AddOrderToQueue(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to queue", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		ch := make(chan entities.Order, 1)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusNew,
			BonusSum: 100,
		}

		op := &OrderProcessor{
			logger:            mockLogger,
			ordersProcessedCh: ch,
		}

		op.AddOrderToQueue(order)

		assert.Equal(t, order, <-ch)
		assert.Equal(t, "Заказ отправлен в очередь на обновление", logs.All()[0].Message)
	})

	t.Run("error - queue is full", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		ch := make(chan entities.Order, 1)

		existingOrder := entities.Order{
			OrderID: "existing-order",
		}

		ch <- existingOrder

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusNew,
			BonusSum: 100,
		}

		op := &OrderProcessor{
			logger:            mockLogger,
			ordersProcessedCh: ch,
		}

		op.AddOrderToQueue(order)

		assert.Equal(t, existingOrder, <-ch)
		assert.Equal(t, "Очередь обработки переполнена", logs.All()[0].Message)
	})
}

func TestOrderProcessor_checkOrderBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to processed queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)
		waiter := mocks.NewMockWaiting(t)

		order := &entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)

		waiter.EXPECT().Wait(mock.Anything).Return(nil)

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
			workerPause:       waiter,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		assert.Equal(t, *order, <-processedCh)
		assert.Equal(t, "Заказ отправлен в очередь на обновление данных", logs.All()[0].Message)
	})

	t.Run("error - order not found in bonus", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)
		waiter := mocks.NewMockWaiting(t)

		domainErr := entities.NewNoContentError(nil, "")

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(nil, domainErr)

		waiter.EXPECT().Wait(mock.Anything).Return(nil)

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
			workerPause:       waiter,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		order := <-processedCh

		assert.Equal(t, "123456789", order.OrderID)
		assert.NotNil(t, order.NextCheckAt)
		assert.Equal(t, "Заказ не найден в bonus, заказу присвоен статус INVALID", logs.All()[0].Message)
		assert.Equal(t, "Заказ отправлен в очередь на обновление данных", logs.All()[1].Message)
	})

	t.Run("error - bonus repository failed", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		mockLogger := zap.NewNop()
		waiter := mocks.NewMockWaiting(t)

		domainErr := entities.NewInternalServerError(
			errors.New("bonus service error"),
			"",
		)

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(nil, domainErr)
		waiter.EXPECT().Wait(mock.Anything).Return(nil)

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
			workerPause:       waiter,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		assert.Empty(t, processedCh)
	})

	t.Run("error - context canceled while sending to queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		waiter := mocks.NewMockWaiting(t)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		order := &entities.Order{
			OrderID: "123456789",
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)
		waiter.EXPECT().Wait(mock.Anything).Return(nil)

		processedCh := make(chan entities.Order)

		op := &OrderProcessor{
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
			workerPause:       waiter,
		}

		op.checkOrderBonus(ctx, "123456789")

		assert.Empty(t, processedCh)
	})
}

func TestOrderProcessor_updateStatusOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - order updated", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		mockOrderRepository.EXPECT().
			UpdateStatusAndBonus(
				mock.Anything,
				order,
			).
			Return(nil)

		op := &OrderProcessor{
			logger:          mockLogger,
			orderRepository: mockOrderRepository,
		}

		op.updateStatusOrder(t.Context(), order)
		assert.Equal(t, "Заказ успешно обновлен", logs.All()[0].Message)
	})

	t.Run("error - empty order id", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		order := entities.Order{
			OrderID: "",
			Status:  entities.OrdersStatusProcessed,
		}

		op := &OrderProcessor{
			logger:          mockLogger,
			orderRepository: mockOrderRepository,
		}

		op.updateStatusOrder(t.Context(), order)
		assert.Equal(t, "updateStatusOrder пустой orderID", logs.All()[0].Message)
	})

	t.Run("error - repository failed", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"",
		)

		mockOrderRepository.EXPECT().
			UpdateStatusAndBonus(
				mock.Anything,
				order,
			).
			Return(domainErr)

		op := &OrderProcessor{
			logger:          mockLogger,
			orderRepository: mockOrderRepository,
		}

		op.updateStatusOrder(t.Context(), order)
		assert.Equal(t, "Ошибка обновления заказа", logs.All()[0].Message)
	})

	t.Run("error - retriable error, order added back to queue", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		domainErr := entities.NewRetriableError(
			errors.New("temporary database error"),
			"",
		)

		mockOrderRepository.EXPECT().
			UpdateStatusAndBonus(
				mock.Anything,
				order,
			).
			Return(domainErr)

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			orderRepository:   mockOrderRepository,
			ordersProcessedCh: processedCh,
		}

		op.updateStatusOrder(t.Context(), order)

		assert.Equal(t, order, <-processedCh)
		assert.Equal(t, "Ошибка обновления заказа, заказ повторно отправлен в очередь", logs.All()[0].Message)
		assert.Equal(t, "Заказ отправлен в очередь на обновление", logs.All()[1].Message)
	})
}

func TestOrderProcessor_orderCheckStatus(t *testing.T) {
	t.Parallel()

	t.Run("success - orders processed from queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)
		waiter := mocks.NewMockWaiting(t)

		order := &entities.Order{
			OrderID: "123456789",
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)

		waiter.EXPECT().Wait(mock.Anything).Return(nil)

		processingCh := make(chan string, 1)
		processedCh := make(chan entities.Order, 1)

		processingCh <- "123456789"
		close(processingCh)

		op := &OrderProcessor{
			logger:             mockLogger,
			bonusRepository:    mockBonusRepository,
			ordersProcessingCh: processingCh,
			ordersProcessedCh:  processedCh,
			workerPause:        waiter,
		}

		op.orderCheckStatus(t.Context())

		assert.Equal(t, *order, <-processedCh)
		assert.Equal(t, "Заказ отправлен в очередь на обновление данных", logs.All()[0].Message)
	})
}

func TestOrderProcessor_orderUpdateBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - order updated from queue", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		core, logs := observer.New(zap.DebugLevel)
		mockLogger := zap.New(core)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		mockOrderRepository.EXPECT().
			UpdateStatusAndBonus(
				mock.Anything,
				order,
			).
			Return(nil)

		processedCh := make(chan entities.Order, 1)
		processedCh <- order
		close(processedCh)

		op := &OrderProcessor{
			logger:            mockLogger,
			orderRepository:   mockOrderRepository,
			ordersProcessedCh: processedCh,
		}

		op.orderUpdateBonus(t.Context())
		assert.Equal(t, "Заказ успешно обновлен", logs.All()[0].Message)
	})
}
