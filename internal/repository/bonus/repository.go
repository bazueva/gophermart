package bonus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Repository структура для работы с сервисом бонусов.
type repository struct {
	client *resty.Client
	addr   string
	logger interfaces.Logger
}

// Order структура для хранения информации о заказе.
type order struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// GetOrder получает информацию о заказе из сервиса бонусов.
func (r *repository) GetOrder(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError) {
	url := fmt.Sprintf("%s/api/orders/%s", r.addr, orderID)

	response, err := r.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Get(url)
	if err != nil {
		r.logger.Error("bonus repository error GetOrder", zap.Error(err), zap.String("order_id", orderID))

		return nil, entities.NewInternalServerError(err, "")
	}

	domainErr := r.checkResponseStatus(response.StatusCode(), orderID)
	if domainErr != nil {
		return nil, domainErr
	}

	var result order
	err = json.Unmarshal(response.Body(), &result)
	if err != nil {
		r.logger.Error("json unmarshal error", zap.Error(err), zap.String("order_id", orderID))

		return nil, entities.NewInternalServerError(err, "")
	}

	if result.Order != orderID || result.Order == "" {
		err = errors.New("заказ не соответствует запрошенному")
		r.logger.Error("неверный заказ",
			zap.Error(err),
			zap.String("order_id", orderID),
			zap.String("result.order_id", result.Order),
		)

		return nil, entities.NewInternalServerError(err, "")
	}

	return &entities.Order{
		OrderID:  result.Order,
		Status:   hydrateOrdersStatusToDomain(result.Status),
		BonusSum: result.Accrual,
	}, nil
}

// checkResponseStatus проверяет статус ответа.
func (r *repository) checkResponseStatus(code int, orderID string) *entities.DomainError {
	switch code {
	case http.StatusNoContent:
		r.logger.Error("Заказ не найден",
			zap.String("order_id", orderID),
		)

		return entities.NewNoContentError(nil, "Заказ не найден")
	case http.StatusOK:
	default:
		r.logger.Error("Ошибка получения заказа из bonus",
			zap.String("order_id", orderID),
			zap.Int("status_code", code),
		)

		return entities.NewInternalServerError(
			nil,
			fmt.Sprintf("Получени статус из bonus - %d", code),
		)
	}

	return nil
}

// NewRepository создает новый экземпляр репозитория бонусов.
func NewRepository(addr string, logger interfaces.Logger) (*repository, error) {
	if addr == "" {
		return nil, fmt.Errorf("не указан адрес сервера")
	}

	return &repository{
		client: createClient(logger),
		logger: logger,
		addr:   addr,
	}, nil
}

// createClient создает новый экземпляр клиента для работы с сервисом бонусов.
func createClient(logger interfaces.Logger) *resty.Client {
	return resty.New().
		SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			if response != nil && response.StatusCode() == http.StatusTooManyRequests {
				if retryAfterHeader := response.Header().Get("Retry-After"); retryAfterHeader != "" {
					if seconds, err := strconv.Atoi(retryAfterHeader); err == nil && seconds > 0 {
						logger.Warn("Сервер ответил 429. Ждем согласно Retry-After",
							zap.Int("seconds", seconds),
							zap.String("url", response.Request.URL),
						)

						return time.Duration(seconds) * time.Second, nil
					}
				}
			}

			attempt := response.Request.Attempt

			return time.Duration(2*attempt-1) * time.Second, nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}

			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		AddRetryHook(
			func(r *resty.Response, err error) {
				logger.Info("Повторная попытка...",
					zap.Error(err),
					zap.String("url", r.Request.URL),
				)
			},
		)
}
