package order

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/helpers"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/service/order/mocks"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrder_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		mockQueue := mocks.NewMockOrderQueue(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).Return(nil, nil)

		mockRepo.EXPECT().
			CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew).
			Return(nil).
			Once()

		mockQueue.EXPECT().AddOrderIDToQueue(orderID)

		mockLogger.EXPECT().Info("Заказ отправлен в очередь на обработку", mock2.Anything)

		o := &Order{
			repository: mockRepo,
			orderQueue: mockQueue,
			logger:     mockLogger,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Nil(t, err)
	})

	t.Run("error - validation luhn failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "order-123"
		userID := int32(123)

		o := &Order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Equal(t, "неверный формат номера заказа", err.Error())
		assert.Equal(t, entities.UnprocessableEntityErrorType, err.ErrorType)
	})

	t.Run("error - FindByOrderID failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to create order",
		)

		mockRepo.EXPECT().FindByOrderID(ctx, orderID).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - create order failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to create order",
		)

		mockRepo.EXPECT().FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew).
			Return(domainErr)

		o := &Order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})
}

func TestOrder_ValidateOrderID(t *testing.T) {
	t.Parallel()

	t.Run("success - valid order ID", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Nil(t, err)
	})

	t.Run("error - validation luhn failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "order-123"
		userID := int32(123)

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "неверный формат номера заказа", err.Error())
		assert.Equal(t, entities.UnprocessableEntityErrorType, err.ErrorType)
	})

	t.Run("error - FindByOrderID failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find order",
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - order already uploaded by this user", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		existingOrder := &entities.Order{
			ID:      1,
			OrderID: orderID,
			UserID:  userID,
			Status:  entities.OrdersStatusNew,
		}

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(existingOrder, nil)

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "номер заказа уже был загружен этим пользователем", err.Error())
		assert.Equal(t, entities.OkEntityErrorType, err.ErrorType)
	})

	t.Run("error - order already uploaded by another user", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		existingOrder := &entities.Order{
			ID:      1,
			OrderID: orderID,
			UserID:  int32(456),
			Status:  entities.OrdersStatusNew,
		}

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(existingOrder, nil)

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "номер заказа уже был загружен другим пользователем", err.Error())
		assert.Equal(t, entities.ConflictErrorType, err.ErrorType)
	})
}

func TestOrder_OrdersListUser(t *testing.T) {
	t.Parallel()

	t.Run("success - get orders list", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		expectedOrders := []entities.Order{
			{
				ID:      1,
				OrderID: "12345678903",
				UserID:  filter.UserID,
				Status:  entities.OrdersStatusNew,
			},
			{
				ID:      2,
				OrderID: "123456789015",
				UserID:  filter.UserID,
				Status:  entities.OrdersStatusProcessed,
			},
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, filter, int64(20), int64(0)).
			Return(expectedOrders, nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Nil(t, err)
		assert.Len(t, orders, 2)
	})

	t.Run("success - no orders found", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(0), nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Nil(t, err)
		assert.Nil(t, orders)
	})

	t.Run("error - count orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to count orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(0), domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - find orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, filter, pag.GetPerPage(), pag.GetOffset()).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}

func TestOrder_OrdersWithdrawalsListUser(t *testing.T) {
	t.Parallel()

	t.Run("success - get withdrawals list", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		expectedOrders := []entities.Order{
			{
				ID:       1,
				OrderID:  "12345678903",
				UserID:   userID,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -100.50,
			},
			{
				ID:       2,
				OrderID:  "123456789015",
				UserID:   userID,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -200.00,
			},
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, ordersFilter, pag.GetPerPage(), pag.GetOffset()).
			Return(expectedOrders, nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Nil(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, int64(2), pag.TotalCount())
	})

	t.Run("success - no withdrawals found", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(0), nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Nil(t, err)
		assert.Nil(t, orders)
	})

	t.Run("error - count orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to count orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(0), domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - find orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, ordersFilter, pag.GetPerPage(), pag.GetOffset()).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}

func TestOrder_BalanceWithdraw(t *testing.T) {
	t.Parallel()

	t.Run("success - balance withdrawn", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)

		ctx := t.Context()

		userID := int32(123)
		orderID := "12345678903"
		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(true),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockRepo.EXPECT().
			UserBalance(mock2.Anything, userID).
			Return(500, nil)

		mockRepo.EXPECT().
			CreateOrderWithWithdraw(
				mock2.Anything,
				userID,
				orderID,
				withdraw.Sum,
			).
			Return(nil)

		mockTx.EXPECT().
			Commit().
			Return(nil)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		o := &Order{
			repository: mockRepo,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		assert.Nil(t, errDomain)
		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("error - validate order failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		expectedErr := entities.NewInternalServerError(
			errors.New("repository error"),
			"",
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, expectedErr)

		o := &Order{
			repository: mockRepo,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)
	})

	t.Run("error - begin transaction failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		logger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		beginErr := errors.New("begin transaction error")

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(nil, beginErr)

		logger.EXPECT().Error("ошибка начала транзакции", mock2.Anything)

		o := &Order{
			repository: mockRepo,
			logger:     logger,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)
	})

	t.Run("error - advisory lock failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnError(errors.New("advisory lock error"))

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		mockLogger.EXPECT().Error("ошибка выполнения запроса блокировки", mock2.Anything)

		o := &Order{
			repository: mockRepo,
			logger:     mockLogger,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)

		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("error - advisory lock already acquired", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(false),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		mockLogger.EXPECT().
			Warn(
				"запрос отклонен: операция уже выполняется параллельно",
				mock2.Anything,
				mock2.Anything,
			)

		o := &Order{
			repository: mockRepo,
			logger:     mockLogger,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.TooManyRequestErrorType, errDomain.ErrorType)

		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("error - get user balance failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		expectedErr := entities.NewInternalServerError(
			errors.New("get balance error"),
			"",
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(true),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockRepo.EXPECT().
			UserBalance(mock2.Anything, userID).
			Return(0, expectedErr)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		o := &Order{
			repository: mockRepo,
			logger:     mockLogger,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, expectedErr, errDomain)

		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("error - withdraw validation failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   600,
		}

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(true),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockRepo.EXPECT().
			UserBalance(mock2.Anything, userID).
			Return(500, nil)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		o := &Order{
			repository: mockRepo,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.PaymentRequiredErrorType, errDomain.ErrorType)
	})

	t.Run("error - create order with withdraw failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		expectedErr := entities.NewInternalServerError(
			errors.New("create order error"),
			"",
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(true),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockRepo.EXPECT().
			UserBalance(mock2.Anything, userID).
			Return(500, nil)

		mockRepo.EXPECT().
			CreateOrderWithWithdraw(
				mock2.Anything,
				userID,
				orderID,
				withdraw.Sum,
			).
			Return(expectedErr)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		o := &Order{
			repository: mockRepo,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, expectedErr, errDomain)

		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("error - commit failed", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockTx := interfacesMocks.NewMockTx(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		userID := int32(123)
		orderID := "12345678903"

		withdraw := entities.BalanceWithdraw{
			Order: orderID,
			Sum:   100,
		}

		commitErr := errors.New("commit error")

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			BeginTransaction(ctx).
			Return(mockTx, nil)

		db, sqlMock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)

		sqlMock.ExpectQuery(`SELECT pg_try_advisory_xact_lock($1, $2)`).
			WithArgs(int64(lockTypeCreateOrderWithdraw), int64(userID)).
			WillReturnRows(
				sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).
					AddRow(true),
			)

		row := db.QueryRow(
			"SELECT pg_try_advisory_xact_lock($1, $2)",
			int64(lockTypeCreateOrderWithdraw),
			int64(userID),
		)

		mockTx.EXPECT().
			QueryRowContext(
				mock2.Anything,
				"SELECT pg_try_advisory_xact_lock($1, $2);",
				[]any{
					int64(lockTypeCreateOrderWithdraw),
					int64(userID),
				},
			).
			Return(row)

		mockRepo.EXPECT().
			UserBalance(mock2.Anything, userID).
			Return(500, nil)

		mockRepo.EXPECT().
			CreateOrderWithWithdraw(
				mock2.Anything,
				userID,
				orderID,
				withdraw.Sum,
			).
			Return(nil)

		mockTx.EXPECT().
			Commit().
			Return(commitErr)

		mockTx.EXPECT().
			Rollback().
			Return(nil)

		mockLogger.EXPECT().Error("ошибка Commit", mock2.Anything)

		o := &Order{
			repository: mockRepo,
			logger:     mockLogger,
		}

		errDomain := o.BalanceWithdraw(ctx, userID, withdraw)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)

		require.NoError(t, sqlMock.ExpectationsWereMet())
	})
}

func TestOrder_validateWithdraw(t *testing.T) {
	t.Parallel()

	t.Run("error - zero balance", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			0,
			entities.BalanceWithdraw{Sum: 100},
		)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.PaymentRequiredErrorType, errDomain.ErrorType)
		assert.Equal(t, "на счету недостаточно средств", errDomain.Text)
	})

	t.Run("error - negative balance", func(t *testing.T) {
		t.Parallel()

		mockLogger := interfacesMocks.NewMockLogger(t)

		mockLogger.EXPECT().
			Error(
				"У пользователя отрицательный баланс",
				mock2.Anything,
				mock2.Anything,
			).
			Once()

		o := &Order{
			logger: mockLogger,
		}

		errDomain := o.validateWithdraw(
			123,
			-100,
			entities.BalanceWithdraw{Sum: 50},
		)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.PaymentRequiredErrorType, errDomain.ErrorType)
		assert.Equal(t, "на счету недостаточно средств", errDomain.Text)
	})

	t.Run("error - zero withdraw sum", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			500,
			entities.BalanceWithdraw{Sum: 0},
		)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.BadRequestErrorType, errDomain.ErrorType)
		assert.Equal(t, "сумма для списания должна быть больше 0", errDomain.Text)
	})

	t.Run("error - negative withdraw sum", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			500,
			entities.BalanceWithdraw{Sum: -100},
		)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.BadRequestErrorType, errDomain.ErrorType)
		assert.Equal(t, "сумма для списания должна быть больше 0", errDomain.Text)
	})

	t.Run("error - withdraw sum greater than balance", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			500,
			entities.BalanceWithdraw{Sum: 600},
		)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.PaymentRequiredErrorType, errDomain.ErrorType)
		assert.Equal(t, "на счету недостаточно средств", errDomain.Text)
	})

	t.Run("success - withdraw sum equals balance", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			500,
			entities.BalanceWithdraw{Sum: 500},
		)

		assert.Nil(t, errDomain)
	})

	t.Run("success - withdraw sum less than balance", func(t *testing.T) {
		t.Parallel()

		o := &Order{}

		errDomain := o.validateWithdraw(
			123,
			500,
			entities.BalanceWithdraw{Sum: 100},
		)

		assert.Nil(t, errDomain)
	})
}
