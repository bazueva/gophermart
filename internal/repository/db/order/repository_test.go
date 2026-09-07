package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	dbPkg "github.com/bazueva/gofermart/internal/repository/db"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestRepository(t *testing.T) (*repository, sqlmock.Sqlmock, *mocks.MockLogger) {
	t.Helper()

	db, mock, err := helpers.SQLMockTest(t)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	logger := mocks.NewMockLogger(t)

	repo := NewRepository(
		dbPkg.NewSQLDBWrapper(db),
		logger,
	)

	return repo, mock, logger
}

func TestRepository_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		t.Parallel()

		repo, mock, _ := newTestRepository(t)

		ctx := t.Context()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusNew

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status)
        VALUES ($1, $2::integer, 'NEW')
        RETURNING orders.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := repo.CreateOrder(ctx, orderID, userID, status)

		assert.Nil(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().Error("error repository CreateOrder", mock2.Anything)

		ctx := t.Context()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusProcessing

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status)
        VALUES ($1, $2::integer, 'PROCESSING')
        RETURNING orders.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID, status).
			WillReturnError(errorDB)

		err := repo.CreateOrder(ctx, orderID, userID, status)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		logger.EXPECT().Error("error repository CreateOrder", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusInvalid

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status)
        VALUES ($1, $2::integer, 'INVALID')
        RETURNING orders.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID, status).
			WillReturnError(context.DeadlineExceeded)

		err := repo.CreateOrder(ctx, orderID, userID, status)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_FindByOrderID(t *testing.T) {
	t.Parallel()

	t.Run("success - find order", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		rows := sqlmock.NewRows([]string{"orders.id", "orders.order_id", "orders.status", "orders.user_id", "orders.created_at"}).
			AddRow(1, "order-123", "PROCESSED", 123, time.Now())

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnRows(rows)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(1), result.ID)
		assert.Equal(t, "order-123", result.OrderID)
		assert.Equal(t, int32(123), result.UserID)
		assert.Equal(t, entities.OrdersStatusProcessed, result.Status)
	})

	t.Run("order not found", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		orderID := "order-not-found"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Nil(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		ctx := t.Context()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(errorDB)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_CountOrdersByUserID(t *testing.T) {
	t.Parallel()

	t.Run("success - find orders", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		rows := sqlmock.NewRows([]string{"count"}).
			AddRow(1)

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID).
			WillReturnRows(rows)

		result, err := repo.CountOrdersByUserID(ctx, filter)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(1), result)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository FindByOrderID", mock2.Anything)

		ctx := t.Context()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID).
			WillReturnError(errorDB)

		result, err := repo.CountOrdersByUserID(ctx, filter)

		assert.Equal(t, int32(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.CountOrdersByUserID(ctx, filter)

		assert.Error(t, err)
		assert.Equal(t, int32(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_FindByUserID(t *testing.T) {
	t.Parallel()

	t.Run("success - find orders by user ID", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at",
             orders.processed_at AS "orders.processed_at",
             orders.bonus_sum AS "orders.bonus_sum"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		createdAt := time.Now()
		rows := sqlmock.NewRows([]string{"orders.id", "orders.order_id", "orders.status", "orders.user_id", "orders.created_at"}).
			AddRow(1, "12345678903", "NEW", 123, createdAt).
			AddRow(2, "123456789015", "PROCESSED", 123, createdAt)

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID, limit, offset).
			WillReturnRows(rows)

		result, err := repo.FindByUserID(ctx, filter, limit, offset)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, entities.Order{
			ID:        1,
			OrderID:   "12345678903",
			Status:    entities.OrdersStatusNew,
			UserID:    123,
			CreatedAt: &createdAt,
		}, result[0])

		assert.Equal(t, entities.Order{
			ID:        2,
			OrderID:   "123456789015",
			Status:    entities.OrdersStatusProcessed,
			UserID:    123,
			CreatedAt: &createdAt,
		}, result[1])
	})

	t.Run("success - no orders found", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at",
             orders.processed_at AS "orders.processed_at",
             orders.bonus_sum AS "orders.bonus_sum"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID, limit, offset).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindByUserID(ctx, filter, limit, offset)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository FindByOrderID", mock2.Anything)

		ctx := t.Context()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at",
             orders.processed_at AS "orders.processed_at",
             orders.bonus_sum AS "orders.bonus_sum"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID, limit, offset).
			WillReturnError(errorDB)

		result, err := repo.FindByUserID(ctx, filter, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		filter := entities.OrderFilter{
			UserID: int32(123),
		}
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(filter.UserID, limit, offset).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindByUserID(ctx, filter, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_FindStaleOrders(t *testing.T) {
	t.Parallel()

	t.Run("success - find stale orders", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE ((orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)) 
          AND ((orders.next_check_at IS NULL) OR (orders.next_check_at < $2::timestamp with time zone))
        ORDER BY orders.created_at ASC
        LIMIT $3;`

		rows := sqlmock.NewRows([]string{"orders.order_id"}).
			AddRow("12345678903").
			AddRow("123456789015")

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), limit).
			WillReturnRows(rows)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "12345678903", result[0])
		assert.Equal(t, "123456789015", result[1])
	})

	t.Run("success - no stale orders found", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE ((orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)) 
          AND ((orders.next_check_at IS NULL) OR (orders.next_check_at < $2::timestamp with time zone))
        ORDER BY orders.created_at ASC
        LIMIT $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), limit).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository FindStaleOrders", mock2.Anything)

		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE ((orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)) 
          AND ((orders.next_check_at IS NULL) OR (orders.next_check_at < $2::timestamp with time zone))
        ORDER BY orders.created_at ASC
        LIMIT $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnError(errorDB)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().Error("error repository FindStaleOrders", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE (orders.status IN ($1::text, $2::text)) AND (orders.created_at < $3::timestamp with time zone)
        LIMIT $4;`

		mock.ExpectQuery(expectedSQL).
			WithArgs("NEW", "PROCESSING", sqlmock.AnyArg(), limit).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_UpdateStatusAndBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - update status and bonus", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		order := entities.Order{
			OrderID:  "12345678903",
			Status:   entities.OrdersStatusProcessing,
			BonusSum: float64(100),
		}

		expectedSQL := `UPDATE public.orders
        SET (bonus_sum, updated_at, next_check_at, status) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs(order.BonusSum, sqlmock.AnyArg(), sqlmock.AnyArg(), model.OrdersStatus_Processing, order.OrderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateStatusAndBonus(ctx, order)

		assert.Nil(t, err)
	})

	t.Run("success - update status invalid and bonus", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)
		ctx := t.Context()

		order := entities.Order{
			OrderID:  "12345678903",
			Status:   entities.OrdersStatusInvalid,
			BonusSum: float64(100),
		}

		expectedSQL := `UPDATE public.orders
        SET (bonus_sum, updated_at, next_check_at, status) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs(order.BonusSum, sqlmock.AnyArg(), sqlmock.AnyArg(), model.OrdersStatus_Invalid, order.OrderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateStatusAndBonus(ctx, order)

		assert.Nil(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository UpdateStatusAndBonus", mock2.Anything)

		ctx := t.Context()

		order := entities.Order{
			OrderID:  "12345678903",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: float64(100),
		}

		expectedSQL := `UPDATE public.orders
        SET (bonus_sum, updated_at, next_check_at, processed_at, status) = ($1, $2, $3, $4, $5)
        WHERE orders.order_id = $6::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs("PROCESSED", order.BonusSum, sqlmock.AnyArg(), order.OrderID).
			WillReturnError(errorDB)

		err := repo.UpdateStatusAndBonus(ctx, order)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().Error("error repository UpdateStatusAndBonus", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		order := entities.Order{
			OrderID:  "12345678903",
			Status:   entities.OrdersStatusProcessed,
			BonusSum: float64(100),
		}

		expectedSQL := `UPDATE public.orders
        SET (status, bonus_sum, updated_at, processed_at) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs("PROCESSED", order.BonusSum, order.OrderID).
			WillReturnError(context.DeadlineExceeded)

		err := repo.UpdateStatusAndBonus(ctx, order)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})
}

func TestRepository_UserBalanceWithWithdrawn(t *testing.T) {
	t.Parallel()

	t.Run("success - get balance with withdrawn", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)

		ctx := t.Context()

		userID := int32(1)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "balance",
             (COALESCE(SUM((CASE WHEN orders.bonus_sum < $2 THEN orders.bonus_sum ELSE $3 END)), $4) * $5) AS "withdrawn"
        FROM public.orders
        WHERE (orders.user_id = $6::integer) AND (orders.status = 'PROCESSED');`

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0), float64(0), float64(0), float64(0), float64(-1), userID,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"balance",
					"withdrawn",
				}).AddRow(1000.50, 250.25),
			)

		result, errDomain := repo.UserBalanceWithWithdrawn(ctx, userID)

		require.Nil(t, errDomain)
		assert.Equal(t, entities.Balance{
			Balance:   1000.50,
			Withdrawn: 250.25,
		}, result)
	})

	t.Run("success - no rows", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)

		ctx := t.Context()

		userID := int32(1)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "balance",
			(COALESCE(SUM((CASE WHEN orders.bonus_sum < $2 THEN orders.bonus_sum ELSE $3 END)), $4) * $5) AS "withdrawn"
		FROM public.orders
		WHERE (orders.user_id = $6::integer) AND (orders.status = 'PROCESSED');`

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0), float64(0), float64(0), float64(0), float64(-1), userID,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"balance",
					"withdrawn",
				}),
			)

		result, errDomain := repo.UserBalanceWithWithdrawn(ctx, userID)

		require.Nil(t, errDomain)
		assert.Equal(t, entities.Balance{}, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().
			Error("error repository UserBalance", mock2.Anything)

		ctx := t.Context()

		userID := int32(1)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "balance",
			(COALESCE(SUM((CASE WHEN orders.bonus_sum < $2 THEN orders.bonus_sum ELSE $3 END)), $4) * $5) AS "withdrawn"
		FROM public.orders
		WHERE (orders.user_id = $6::integer) AND (orders.status = 'PROCESSED');`

		dbErr := errors.New("database error")

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0), float64(0), float64(0), float64(0), float64(-1), userID,
			).
			WillReturnError(dbErr)

		result, errDomain := repo.UserBalanceWithWithdrawn(ctx, userID)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.Balance{}, result)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)
	})

	t.Run("error - context canceled", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)
		logger.EXPECT().
			Error("error repository UserBalance", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		userID := int32(1)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "balance",
			(COALESCE(SUM((CASE WHEN orders.bonus_sum < $2 THEN orders.bonus_sum ELSE $3 END)), $4) * $5) AS "withdrawn"
		FROM public.orders
		WHERE (orders.user_id = $6::integer) AND (orders.status = 'PROCESSED');`

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0), float64(0), float64(0), float64(0), float64(-1), userID,
			).
			WillReturnError(context.Canceled)

		result, errDomain := repo.UserBalanceWithWithdrawn(ctx, userID)

		require.NotNil(t, errDomain)
		assert.Equal(t, entities.Balance{}, result)
		assert.Equal(t, entities.InternalServerErrorType, errDomain.ErrorType)
	})
}

func TestRepository_CreateOrderWithWithdraw(t *testing.T) {
	t.Run("success - create order with withdraw", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)

		ctx := t.Context()

		userID := int32(123)
		orderID := "order-123"
		bonusSum := 100.50

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status, bonus_sum, processed_at)
        VALUES ($1, $2::integer, 'PROCESSED', $3::double precision, $4::timestamp with time zone)
        RETURNING orders.id AS "id";`

		mock.ExpectExec(expectedSQL).
			WithArgs(
				orderID,
				userID,
				-bonusSum,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateOrderWithWithdraw(
			ctx,
			userID,
			orderID,
			bonusSum,
		)

		require.Nil(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository CreateOrderWithWithdraw", mock2.Anything)

		ctx := t.Context()

		userID := int32(123)
		orderID := "order-123"
		bonusSum := 100.50

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status, bonus_sum, processed_at) 
VALUES ($1, $2::integer, 'PROCESSED', $3::double precision, $4::timestamp with time zone) 
RETURNING orders.id AS "id";`

		mock.ExpectExec(expectedSQL).
			WithArgs(
				orderID,
				userID,
				-bonusSum,
				sqlmock.AnyArg(),
			).
			WillReturnError(errorDB)

		err := repo.CreateOrderWithWithdraw(
			ctx,
			userID,
			orderID,
			bonusSum,
		)

		require.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(
			t,
			entities.InternalServerErrorType,
			err.ErrorType,
		)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - context canceled", func(t *testing.T) {
		repo, _, logger := newTestRepository(t)

		logger.EXPECT().
			Error("error repository CreateOrderWithWithdraw", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		userID := int32(123)
		orderID := "order-123"
		bonusSum := 100.50

		err := repo.CreateOrderWithWithdraw(
			ctx,
			userID,
			orderID,
			bonusSum,
		)

		require.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(
			t,
			entities.InternalServerErrorType,
			err.ErrorType,
		)
	})
}

func TestRepository_UserBalance(t *testing.T) {
	t.Run("success - get user balance", func(t *testing.T) {
		repo, mock, _ := newTestRepository(t)

		ctx := t.Context()

		userID := int32(123)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "sum"
FROM public.orders
WHERE (orders.user_id = $2::integer) AND (orders.status = 'PROCESSED');`

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0),
				userID,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"sum",
				}).AddRow(1000.50),
			)

		result, err := repo.UserBalance(ctx, userID)

		require.Nil(t, err)
		assert.Equal(t, 1000.50, result)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - database error", func(t *testing.T) {
		repo, mock, logger := newTestRepository(t)

		errorDB := errors.New("database connection failed")

		logger.EXPECT().
			Error("error repository UserBalance", mock2.Anything)

		ctx := t.Context()

		userID := int32(123)

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "sum"
FROM public.orders
WHERE (orders.user_id = $2::integer) AND (orders.status = 'PROCESSED');`

		mock.ExpectQuery(expectedSQL).
			WithArgs(
				float64(0),
				userID,
			).
			WillReturnError(errorDB)

		result, err := repo.UserBalance(ctx, userID)

		require.Error(t, err)
		assert.Equal(t, float64(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(
			t,
			entities.InternalServerErrorType,
			err.ErrorType,
		)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - context canceled", func(t *testing.T) {
		repo, _, logger := newTestRepository(t)

		logger.EXPECT().
			Error("error repository UserBalance", mock2.Anything)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		userID := int32(123)

		result, err := repo.UserBalance(ctx, userID)

		require.Error(t, err)
		assert.Equal(t, float64(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(
			t,
			entities.InternalServerErrorType,
			err.ErrorType,
		)
	})
}
