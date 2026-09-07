package queries

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/enum"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/go-jet/jet/v2/postgres"
)

// hydrateDomainToOrdersStatusEnum преобразование статуса заказа из доменной модели в тип enum.
func hydrateDomainToOrdersStatusEnum(status entities.OrderStatus) postgres.Expression {
	switch status {
	case entities.OrdersStatusNew:
		return enum.OrdersStatus.New
	case entities.OrdersStatusProcessing:
		return enum.OrdersStatus.Processing
	case entities.OrdersStatusInvalid:
		return enum.OrdersStatus.Invalid
	case entities.OrdersStatusProcessed:
		return enum.OrdersStatus.Processed
	default:
		return enum.OrdersStatus.New
	}
}

// hydrateDomainToOrdersStatus преобразование статуса заказа из доменной модели в тип БД.
func hydrateDomainToOrdersStatus(status entities.OrderStatus) dbModel.OrdersStatus {
	switch status {
	case entities.OrdersStatusNew:
		return dbModel.OrdersStatus_New
	case entities.OrdersStatusProcessing:
		return dbModel.OrdersStatus_Processing
	case entities.OrdersStatusInvalid:
		return dbModel.OrdersStatus_Invalid
	case entities.OrdersStatusProcessed:
		return dbModel.OrdersStatus_Processed
	default:
		return dbModel.OrdersStatus(status)
	}
}
