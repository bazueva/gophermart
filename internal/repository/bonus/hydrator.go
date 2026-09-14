package bonus

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
)

// hydrateOrdersStatusToDomain преобразует статус заказа из сервиса бонусов в доменный статус.
func hydrateOrdersStatusToDomain(status string) entities.OrderStatus {
	switch status {
	case "REGISTERED":
		return entities.OrdersStatusNew
	case "PROCESSING":
		return entities.OrdersStatusProcessing
	case "INVALID":
		return entities.OrdersStatusInvalid
	case "PROCESSED":
		return entities.OrdersStatusProcessed
	default:
		return entities.OrdersStatusNew
	}
}
