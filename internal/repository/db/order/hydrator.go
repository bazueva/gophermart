package order

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
)

// hydrateOrdersStatusToDomain преобразование статуса заказа из базы данных в доменный статус.
func hydrateOrdersStatusToDomain(status dbModel.OrdersStatus) entities.OrderStatus {
	switch status {
	case dbModel.OrdersStatus_New:
		return entities.OrdersStatusNew
	case dbModel.OrdersStatus_Processing:
		return entities.OrdersStatusProcessing
	case dbModel.OrdersStatus_Invalid:
		return entities.OrdersStatusInvalid
	case dbModel.OrdersStatus_Processed:
		return entities.OrdersStatusProcessed
	default:
		return entities.OrdersStatusNew
	}
}
