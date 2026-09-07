package app

import (
	"errors"
	"testing"

	appMocks "github.com/bazueva/gofermart/internal/app/mocks"
	"github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestApp_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("ctx without userID", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("ctx without userID", mock2.Anything)

		appTest := NewApp(nil, nil, logger)
		err := appTest.CreateOrder(t.Context(), "test")

		assert.Equal(t, "Internal Server Error", err.Error())
	})

	t.Run("userID = 0", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 0)
		appTest := NewApp(nil, nil, logger)
		err := appTest.CreateOrder(ctx, "test")

		assert.Equal(t, "пользователь не аутентифицирован", err.Error())
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})

	t.Run("error create order", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 20)

		orderService := appMocks.NewMockOrderService(t)
		orderService.EXPECT().CreateOrder(ctx, "12323", int32(20)).
			Return(entities.NewInternalServerError(nil, ""))

		appTest := NewApp(nil, orderService, logger)
		err := appTest.CreateOrder(ctx, "12323")

		assert.Equal(t, "Internal Server Error", err.Error())
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 20)

		orderService := appMocks.NewMockOrderService(t)
		orderService.EXPECT().CreateOrder(ctx, "12323", int32(20)).
			Return(nil)

		appTest := NewApp(nil, orderService, logger)
		err := appTest.CreateOrder(ctx, "12323")

		assert.Nil(t, err)
	})
}

func TestApp_CheckJWTToken_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("error userService", func(t *testing.T) {
		mockUserService := appMocks.NewMockUserService(t)

		domainErr := entities.NewBadRequestError(
			errors.New("empty token"),
			"token cannot be empty",
		)
		mockUserService.EXPECT().CheckJWTToken("").
			Return(int32(0), domainErr).
			Once()

		appMocks := &App{
			userService: mockUserService,
		}

		userID, err := appMocks.CheckJWTToken("")

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.Equal(t, entities.BadRequestErrorType, err.ErrorType)

		mockUserService.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockUserService := appMocks.NewMockUserService(t)

		mockUserService.EXPECT().CheckJWTToken("token").
			Return(int32(1), nil).
			Once()

		appTest := &App{
			userService: mockUserService,
		}

		userID, err := appTest.CheckJWTToken("token")

		assert.Nil(t, err)
		assert.Equal(t, int32(1), userID)
	})
}

func TestApp_Login(t *testing.T) {
	t.Parallel()

	t.Run("success - valid credentials", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		request := models.LoginRequest{
			Login:    "testuser",
			Password: "correct_password",
		}
		expectedToken := "valid.jwt.token"

		mockUserService.EXPECT().
			Login(t.Context(), forms.LoginForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return(expectedToken, nil)

		appTest := &App{
			userService: mockUserService,
		}

		token, err := appTest.Login(t.Context(), request)

		assert.Nil(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("error - invalid credentials", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.LoginRequest{
			Login:    "testuser",
			Password: "wrong_password",
		}
		domainErr := entities.NewUnauthorizedError(
			errors.New("invalid credentials"),
			"invalid login or password",
		)

		mockUserService.EXPECT().
			Login(ctx, forms.LoginForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return("", domainErr)

		appTest := &App{
			userService: mockUserService,
		}

		token, err := appTest.Login(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, "", token)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
		mockUserService.AssertExpectations(t)
	})
}

func TestApp_Register(t *testing.T) {
	t.Parallel()

	t.Run("success - valid registration", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.RegisterRequest{
			Login:    "newuser",
			Password: "password123",
		}
		expectedToken := "valid.jwt.token"

		mockUserService.EXPECT().
			Register(ctx, forms.UserForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return(expectedToken, nil)

		appTest := &App{
			userService: mockUserService,
		}

		token, err := appTest.Register(ctx, request)

		assert.Nil(t, err)
		assert.Equal(t, expectedToken, token)
		mockUserService.AssertExpectations(t)
	})

	t.Run("error - user already exists", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.RegisterRequest{
			Login:    "existinguser",
			Password: "password123",
		}
		domainErr := entities.NewConflictError(
			errors.New("user already exists"),
			"user with this login already exists",
		)

		mockUserService.EXPECT().
			Register(ctx, forms.UserForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return("", domainErr).
			Once()

		appTest := &App{
			userService: mockUserService,
		}

		token, err := appTest.Register(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, "", token)
		assert.Equal(t, entities.ConflictErrorType, err.ErrorType)
		mockUserService.AssertExpectations(t)
	})
}

func TestApp_UserOrdersList(t *testing.T) {
	t.Parallel()

	t.Run("success - get user orders", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)

		userID := int32(123)
		ctx := context.WithUserID(t.Context(), userID)

		page := int32(1)
		perPage := int32(20)

		expectedOrders := []entities.Order{
			{
				ID:      1,
				OrderID: "12345678903",
				UserID:  userID,
				Status:  entities.OrdersStatusNew,
			},
		}

		mockOrderService.EXPECT().
			OrdersListUser(ctx, userID, pagination.NewPagination(int64(page), int64(perPage))).
			Return(expectedOrders, nil)

		a := &App{
			orderService: mockOrderService,
		}

		orders, err := a.UserOrdersList(ctx, page, perPage)

		assert.Nil(t, err)
		assert.Equal(t, expectedOrders, orders)
	})

	t.Run("error - user not authorized", func(t *testing.T) {
		ctx := t.Context()
		page := int32(1)
		perPage := int32(20)

		domainErr := entities.NewInternalServerError(
			errors.New("ctx without userID"),
			"",
		)

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("ctx without userID", mock2.Anything)

		a := &App{
			logger: logger,
		}

		orders, err := a.UserOrdersList(ctx, page, perPage)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - order service failed", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)

		userID := int32(123)
		ctx := context.WithUserID(t.Context(), userID)
		page := int32(1)
		perPage := int32(20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to get orders",
		)

		mockOrderService.EXPECT().
			OrdersListUser(ctx, userID, pagination.NewPagination(int64(page), int64(perPage))).
			Return(nil, domainErr)

		a := &App{
			orderService: mockOrderService,
		}

		orders, err := a.UserOrdersList(ctx, page, perPage)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}

func TestApp_UserIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("success - get userID from context", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)
		ctx := context.WithUserID(t.Context(), 123)

		a := &App{
			logger: logger,
		}

		userID, err := a.userIDFromContext(ctx, false)

		assert.Nil(t, err)
		assert.Equal(t, int32(123), userID)
	})

	t.Run("success - userID 0 with errorIfEmpty false", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)
		ctx := context.WithUserID(t.Context(), 0)

		a := &App{
			logger: logger,
		}

		userID, err := a.userIDFromContext(ctx, false)

		assert.Nil(t, err)
		assert.Equal(t, int32(0), userID)
	})

	t.Run("error - context without Auth", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("ctx without userID", mock2.Anything)

		ctx := t.Context()

		a := &App{
			logger: logger,
		}

		userID, err := a.userIDFromContext(ctx, false)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("error - userID 0 with errorIfEmpty true", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)
		ctx := context.WithUserID(t.Context(), 0)

		a := &App{
			logger: logger,
		}

		userID, err := a.userIDFromContext(ctx, true)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})
}

func TestApp_BalanceWithDraw(t *testing.T) {
	t.Parallel()

	t.Run("success - balance withdraw", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)
		request := models.BalanceWithdrawRequest{
			Order: "12345678903",
			Sum:   100.50,
		}

		mockOrderService.EXPECT().
			BalanceWithdraw(ctx, int32(123), entities.BalanceWithdraw{
				Order: request.Order,
				Sum:   request.Sum,
			}).
			Return(nil)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		err := a.BalanceWithDraw(ctx, request)

		assert.Nil(t, err)
	})

	t.Run("error - user not authorized", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 0)
		request := models.BalanceWithdrawRequest{
			Order: "12345678903",
			Sum:   100.50,
		}

		a := &App{
			logger: logger,
		}

		err := a.BalanceWithDraw(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})

	t.Run("error - context without Auth", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("ctx without userID", mock2.Anything)

		ctx := t.Context()
		request := models.BalanceWithdrawRequest{
			Order: "12345678903",
			Sum:   100.50,
		}

		a := &App{
			logger: logger,
		}

		err := a.BalanceWithDraw(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("error - order service failed", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)
		request := models.BalanceWithdrawRequest{
			Order: "12345678903",
			Sum:   100.50,
		}

		domainErr := entities.NewBadRequestError(
			errors.New("insufficient balance"),
			"insufficient balance",
		)

		mockOrderService.EXPECT().
			BalanceWithdraw(ctx, int32(123), entities.BalanceWithdraw{
				Order: request.Order,
				Sum:   request.Sum,
			}).
			Return(domainErr)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		err := a.BalanceWithDraw(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})
}

func TestApp_UserWithdrawals(t *testing.T) {
	t.Parallel()

	t.Run("success - get user withdrawals", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)
		page := int32(1)
		perPage := int32(20)

		expectedOrders := []entities.Order{
			{
				ID:       1,
				OrderID:  "12345678903",
				UserID:   123,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -100.50,
			},
			{
				ID:       2,
				OrderID:  "123456789015",
				UserID:   123,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -200.00,
			},
		}

		mockOrderService.EXPECT().
			OrdersWithdrawalsListUser(ctx, int32(123), pagination.NewPagination(int64(page), int64(perPage))).
			Return(expectedOrders, nil)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		orders, err := a.UserWithdrawals(ctx, page, perPage)

		assert.Nil(t, err)
		assert.Equal(t, expectedOrders, orders)
	})

	t.Run("success - empty withdrawals list", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)
		page := int32(1)
		perPage := int32(20)

		mockOrderService.EXPECT().
			OrdersWithdrawalsListUser(ctx, int32(123), pagination.NewPagination(int64(page), int64(perPage))).
			Return([]entities.Order{}, nil)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		orders, err := a.UserWithdrawals(ctx, page, perPage)

		assert.Nil(t, err)
		assert.Empty(t, orders)
	})

	t.Run("error - user not authorized", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 0)
		page := int32(1)
		perPage := int32(20)

		a := &App{
			logger: logger,
		}

		orders, err := a.UserWithdrawals(ctx, page, perPage)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})

	t.Run("error - context without Auth", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)

		ctx := t.Context()
		page := int32(1)
		perPage := int32(20)

		a := &App{
			logger: logger,
		}

		logger.EXPECT().Error("ctx without userID", mock2.Anything).Return()

		orders, err := a.UserWithdrawals(ctx, page, perPage)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("error - order service failed", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)
		page := int32(1)
		perPage := int32(20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to get withdrawals",
		)

		mockOrderService.EXPECT().
			OrdersWithdrawalsListUser(ctx, int32(123), pagination.NewPagination(int64(page), int64(perPage))).
			Return(nil, domainErr)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		orders, err := a.UserWithdrawals(ctx, page, perPage)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}

func TestApp_UserBalance(t *testing.T) {
	t.Parallel()

	t.Run("success balance", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)

		mockOrderService.EXPECT().
			UserBalance(ctx, int32(123)).
			Return(entities.Balance{Balance: 1, Withdrawn: 2}, nil)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		balance, err := a.UserBalance(ctx)

		assert.Nil(t, err)
		assert.Equal(t, entities.Balance{
			Balance:   1,
			Withdrawn: 2,
		}, balance)
	})

	t.Run("error - user not authorized", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 0)

		a := &App{
			logger: logger,
		}

		balance, err := a.UserBalance(ctx)

		assert.Error(t, err)
		assert.Empty(t, balance)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})

	t.Run("error - context without Auth", func(t *testing.T) {
		logger := mocks.NewMockLogger(t)

		ctx := t.Context()

		a := &App{
			logger: logger,
		}

		logger.EXPECT().Error("ctx without userID", mock2.Anything).Return()

		balance, err := a.UserBalance(ctx)

		assert.Error(t, err)
		assert.Empty(t, balance)
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("error - order service failed", func(t *testing.T) {
		mockOrderService := appMocks.NewMockOrderService(t)
		logger := mocks.NewMockLogger(t)

		ctx := context.WithUserID(t.Context(), 123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to get withdrawals",
		)

		mockOrderService.EXPECT().
			UserBalance(ctx, int32(123)).
			Return(entities.Balance{}, domainErr)

		a := &App{
			orderService: mockOrderService,
			logger:       logger,
		}

		balance, err := a.UserBalance(ctx)

		assert.Error(t, err)
		assert.Empty(t, balance)
		assert.Equal(t, domainErr, err)
	})
}
