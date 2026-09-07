package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

type App interface {
	Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError)
	Login(ctx context.Context, request models.LoginRequest) (string, *entities.DomainError)
	CreateOrder(ctx context.Context, id string) *entities.DomainError
	UserOrdersList(ctx context.Context, page int32, perPage int32) ([]entities.Order, *entities.DomainError)
	BalanceWithDraw(ctx context.Context, request models.BalanceWithdrawRequest) *entities.DomainError
	UserWithdrawals(ctx context.Context, toInt32 int32, toInt33 int32) ([]entities.Order, *entities.DomainError)
	UserBalance(ctx context.Context) (entities.Balance, *entities.DomainError)
}

// Handler обрабатывает HTTP-запросы приложения.
type Handler struct {
	logger interfaces.Logger
	app    App
}

// NewHandler создает новый обработчик HTTP-запросов.
func NewHandler(logger interfaces.Logger, application App) *Handler {
	return &Handler{
		logger: logger,
		app:    application,
	}
}

// LoginUser Аутентификация пользователя.
//
// POST /api/user/login HTTP/1.1
//
// Тело запроса:
//
//	{
//	    "login": "user",
//	    "password": "password"
//	}
//
// Ответ:
//
//	{
//	    "token": "eyJhbGciOiJIUzI1NiIs..."
//	}
//
// При успешной аутентификации возвращает HTTP 200 OK.
func (h *Handler) LoginUser(w http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			h.logger.Error("body close error", zap.Error(err))
		}
	}()

	var loginRequest models.LoginRequest

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&loginRequest)
	if err != nil {
		h.errorHandler(w, err, http.StatusBadRequest)

		return
	}

	tokenJWT, errDomain := h.app.Login(request.Context(), loginRequest)
	if errDomain != nil {
		h.errorHandler(w, errDomain, 0)

		return
	}

	// для тестов
	w.Header().Set("Authorization", "Bearer "+tokenJWT)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenJWT,
	})
}

// RegisterUser Регистрация нового пользователя.
//
// POST /api/user/register HTTP/1.1
//
// Тело запроса:
//
//	{
//	    "login": "user",
//	    "password": "password"
//	}
//
// Ответ:
//
//	{
//	    "token": "eyJhbGciOiJIUzI1NiIs..."
//	}
//
// При успешной регистрации возвращает HTTP 200 OK.
func (h *Handler) RegisterUser(w http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			h.logger.Error("body close error", zap.Error(err))
		}
	}()

	var regRequest models.RegisterRequest

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&regRequest)
	if err != nil {
		h.errorHandler(w, err, http.StatusBadRequest)

		return
	}

	tokenJWT, errDomain := h.app.Register(request.Context(), regRequest)
	if errDomain != nil {
		h.errorHandler(w, errDomain, 0)

		return
	}

	// для тестов
	w.Header().Set("Authorization", "Bearer "+tokenJWT)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenJWT,
	})
}

func (h *Handler) errorHandler(writer http.ResponseWriter, err error, statusCode int) {
	if domainError, ok := errors.AsType[*entities.DomainError](err); ok {
		switch domainError.ErrorType {
		case entities.ConflictErrorType:
			statusCode = http.StatusConflict
		case entities.BadRequestErrorType:
			statusCode = http.StatusBadRequest
		case entities.UnauthorizedErrorType:
			statusCode = http.StatusUnauthorized
		case entities.UnprocessableEntityErrorType:
			statusCode = http.StatusUnprocessableEntity
		case entities.OkEntityErrorType:
			statusCode = http.StatusOK
		case entities.PaymentRequiredErrorType:
			statusCode = http.StatusPaymentRequired

		default:
			statusCode = http.StatusInternalServerError
		}
	}

	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(map[string]string{
		"error": err.Error(),
	})
}

// CreateOrder Регистрация нового заказа пользователя.
//
// POST /api/user/orders HTTP/1.1
//
// Тело запроса:
//
//	237722727171
//
// При успешной регистрации возвращает HTTP 202 Accepted.
func (h *Handler) CreateOrder(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.errorHandler(writer, err, http.StatusBadRequest)

		return
	}

	defer func() {
		_ = request.Body.Close()
	}()

	orderID := string(body)
	if orderID == "" {
		h.errorHandler(writer, errors.New("не указан номер заказа"), http.StatusBadRequest)

		return
	}

	errDomain := h.app.CreateOrder(request.Context(), orderID)
	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	writer.WriteHeader(http.StatusAccepted)
}

// UserOrdersList Получение списка заказов пользователя.
//
// GET /api/user/orders?page=1&perPage=10 HTTP/1.1
//
// Ответ:
//
//	[
//	    {
//	        "number": "237722727171",
//	        "status": "PROCESSED",
//	        "accrual": 500.5,
//	        "uploaded_at": "2020-12-10T15:15:45+03:00"
//	    },
//	    {
//	        "number": "237722727172",
//	        "status": "PROCESSING",
//	        "accrual": 42,
//	        "uploaded_at": ""
//	    }
//	]
//
// Если заказов нет, возвращается HTTP 204 No Content.
func (h *Handler) UserOrdersList(writer http.ResponseWriter, request *http.Request) {
	orders, errDomain := h.app.UserOrdersList(
		request.Context(),
		cast.ToInt32(request.URL.Query().Get("page")),
		cast.ToInt32(request.URL.Query().Get("perPage")),
	)
	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	if len(orders) == 0 {
		writer.WriteHeader(http.StatusNoContent)

		return
	}

	result := make([]models.Order, len(orders))
	for i, order := range orders {
		result[i] = models.Order{
			Number:  order.OrderID,
			Status:  string(order.Status),
			Accrual: order.BonusSum,
			UploadedAt: func() string {
				if order.ProcessedAt != nil {
					return order.ProcessedAt.Format("2006-01-02T15:04:05-07:00")
				}

				return ""
			}(),
		}
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		h.logger.Error("error marshal []order", zap.Error(err))

		h.errorHandler(writer, entities.NewInternalServerError(err, ""), 0)

		return
	}

	writer.Write(resultJSON)
}

// BalanceWithdraw Списание бонусов с баланса пользователя.
//
// POST /api/user/balance/withdraw HTTP/1.1
//
// Запрос:
//
//	{
//	    "order": "237722727171",
//	    "sum": 500.5
//	}
//
// При успешном выполнении возвращает HTTP 200 OK.
func (h *Handler) BalanceWithdraw(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			h.logger.Error("body close error", zap.Error(err))
		}
	}()

	var balanceRequest models.BalanceWithdrawRequest

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&balanceRequest)
	if err != nil {
		h.errorHandler(writer, err, http.StatusBadRequest)

		return
	}

	errDomain := h.app.BalanceWithDraw(request.Context(), balanceRequest)
	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// UserWithdrawals Получение списка списаний пользователя.
//
// GET /api/user/withdrawals?page=1&perPage=10 HTTP/1.1
//
// Ответ:
//
//	[
//	    {
//	        "order": "237722727171",
//	        "sum": 500.5,
//	        "processed_at": "2020-12-10T15:15:45+03:00"
//	    },
//	    {
//	        "order": "237722727172",
//	        "sum": 42,
//	        "processed_at": "2020-12-11T10:20:30+03:00"
//	    }
//	]
func (h *Handler) UserWithdrawals(writer http.ResponseWriter, request *http.Request) {
	withdrawals, errDomain := h.app.UserWithdrawals(
		request.Context(),
		cast.ToInt32(request.URL.Query().Get("page")),
		cast.ToInt32(request.URL.Query().Get("perPage")),
	)

	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	if len(withdrawals) == 0 {
		writer.WriteHeader(http.StatusNoContent)

		return
	}

	result := make([]models.Withdrawal, len(withdrawals))
	for i, order := range withdrawals {
		result[i] = models.Withdrawal{
			Order: order.OrderID,
			Sum:   order.BonusSum * -1,
			ProcessedAt: func() string {
				if order.ProcessedAt != nil {
					return order.ProcessedAt.Format("2006-01-02T15:04:05-07:00")
				}

				return ""
			}(),
		}
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		h.logger.Error("error marshal []order", zap.Error(err))

		h.errorHandler(writer, entities.NewInternalServerError(err, ""), 0)

		return
	}

	writer.Write(resultJSON)
}

// UserBalance Получение текущего баланса пользователя.
//
// GET /api/user/balance HTTP/1.1
//
// Ответ:
//
//	{
//	    "current": 500.5,
//	    "withdrawn": 42
//	}
func (h *Handler) UserBalance(writer http.ResponseWriter, request *http.Request) {
	result, errDomain := h.app.UserBalance(request.Context())
	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	balance := models.Balance{
		Current:   result.Balance,
		Withdrawn: result.Withdrawn,
	}

	jsonBalance, err := json.Marshal(balance)
	if err != nil {
		h.logger.Error("error marshal balance", zap.Error(err))

		h.errorHandler(writer, entities.NewInternalServerError(err, ""), 0)

		return
	}

	writer.Write(jsonBalance)
}
