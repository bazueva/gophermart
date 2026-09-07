package order

import (
	"context"
	"errors"
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/service/order/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderProcessor_AddOrderIDToQueue(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to queue", func(t *testing.T) {
		t.Parallel()

		mockLogger := interfacesMocks.NewMockLogger(t)

		ch := make(chan string, 1)

		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обработку",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:             mockLogger,
			ordersProcessingCh: ch,
		}

		op.AddOrderIDToQueue("123456789")

		assert.Equal(t, "123456789", <-ch)
	})

	t.Run("error - queue is full", func(t *testing.T) {
		t.Parallel()

		mockLogger := interfacesMocks.NewMockLogger(t)

		ch := make(chan string, 1)
		ch <- "existing-order"

		mockLogger.EXPECT().
			Warn(
				"Очередь обработки переполнена, заказ обработается позже из БД",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:             mockLogger,
			ordersProcessingCh: ch,
		}

		op.AddOrderIDToQueue("123456789")

		assert.Equal(t, "existing-order", <-ch)
	})
}

func TestOrderProcessor_AddOrderToQueue(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to queue", func(t *testing.T) {
		t.Parallel()

		mockLogger := interfacesMocks.NewMockLogger(t)

		ch := make(chan entities.Order, 1)

		order := entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusNew,
			BonusSum: 100,
		}

		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обновление",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:            mockLogger,
			ordersProcessedCh: ch,
		}

		op.AddOrderToQueue(order)

		assert.Equal(t, order, <-ch)
	})

	t.Run("error - queue is full", func(t *testing.T) {
		t.Parallel()

		mockLogger := interfacesMocks.NewMockLogger(t)

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

		mockLogger.EXPECT().
			Warn(
				"Очередь обработки переполнена",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:            mockLogger,
			ordersProcessedCh: ch,
		}

		op.AddOrderToQueue(order)

		assert.Equal(t, existingOrder, <-ch)
	})
}

func TestOrderProcessor_checkOrderBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - order added to processed queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		order := &entities.Order{
			OrderID:  "123456789",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: 100,
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)

		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обновление данных",
				mock.Anything,
				mock.Anything,
			).
			Once()

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		assert.Equal(t, *order, <-processedCh)
	})

	t.Run("error - order not found in bonus", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		domainErr := entities.NewNoContentError(nil, "")

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(nil, domainErr)

		mockLogger.EXPECT().
			Info(
				"Заказ не найден в bonus, заказу присвоен статус INVALID",
				mock.Anything,
			).
			Once()
		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обновление данных",
				mock.Anything,
			).
			Once()

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		order := <-processedCh

		assert.Equal(t, "123456789", order.OrderID)
		assert.NotNil(t, order.NextCheckAt)
	})

	t.Run("error - bonus repository failed", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		domainErr := entities.NewInternalServerError(
			errors.New("bonus service error"),
			"",
		)

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(nil, domainErr)

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
		}

		op.checkOrderBonus(t.Context(), "123456789")

		assert.Empty(t, processedCh)
	})

	t.Run("error - context canceled while sending to queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		order := &entities.Order{
			OrderID: "123456789",
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)

		processedCh := make(chan entities.Order)

		op := &OrderProcessor{
			bonusRepository:   mockBonusRepository,
			ordersProcessedCh: processedCh,
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
		mockLogger := interfacesMocks.NewMockLogger(t)

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

		mockLogger.EXPECT().
			Info(
				"Заказ успешно обновлен",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:          mockLogger,
			orderRepository: mockOrderRepository,
		}

		op.updateStatusOrder(t.Context(), order)
	})

	t.Run("error - empty order id", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		order := entities.Order{
			OrderID: "",
			Status:  entities.OrdersStatusProcessed,
		}

		mockLogger.EXPECT().
			Info(
				"updateStatusOrder пустой orderID",
				mock.Anything,
			).
			Once()

		op := &OrderProcessor{
			logger:          mockLogger,
			orderRepository: mockOrderRepository,
		}

		op.updateStatusOrder(t.Context(), order)
	})

	t.Run("error - repository failed", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

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
	})

	t.Run("error - retriable error, order added back to queue", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

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

		mockLogger.EXPECT().
			Error(
				"Ошибка обновления заказа, заказ повторно отправлен в очередь",
				mock.Anything,
				mock.Anything,
			).
			Once()

		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обновление",
				mock.Anything,
			).
			Once()

		processedCh := make(chan entities.Order, 1)

		op := &OrderProcessor{
			logger:            mockLogger,
			orderRepository:   mockOrderRepository,
			ordersProcessedCh: processedCh,
		}

		op.updateStatusOrder(t.Context(), order)

		assert.Equal(t, order, <-processedCh)
	})
}

func TestOrderProcessor_orderCheckStatus(t *testing.T) {
	t.Parallel()

	t.Run("success - orders processed from queue", func(t *testing.T) {
		t.Parallel()

		mockBonusRepository := mocks.NewMockBonusRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		order := &entities.Order{
			OrderID: "123456789",
		}

		mockBonusRepository.EXPECT().
			GetOrder(mock.Anything, "123456789").
			Return(order, nil)

		mockLogger.EXPECT().
			Info(
				"Заказ отправлен в очередь на обновление данных",
				mock.Anything,
				mock.Anything,
			).
			Once()

		processingCh := make(chan string, 1)
		processedCh := make(chan entities.Order, 1)

		processingCh <- "123456789"
		close(processingCh)

		op := &OrderProcessor{
			logger:             mockLogger,
			bonusRepository:    mockBonusRepository,
			ordersProcessingCh: processingCh,
			ordersProcessedCh:  processedCh,
		}

		op.orderCheckStatus(t.Context())

		assert.Equal(t, *order, <-processedCh)
	})
}

func TestOrderProcessor_orderUpdateBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - order updated from queue", func(t *testing.T) {
		t.Parallel()

		mockOrderRepository := mocks.NewMockOrderRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

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

		mockLogger.EXPECT().
			Info(
				"Заказ успешно обновлен",
				mock.Anything,
			).
			Once()

		processedCh := make(chan entities.Order, 1)
		processedCh <- order
		close(processedCh)

		op := &OrderProcessor{
			logger:            mockLogger,
			orderRepository:   mockOrderRepository,
			ordersProcessedCh: processedCh,
		}

		op.orderUpdateBonus(t.Context())
	})
}
