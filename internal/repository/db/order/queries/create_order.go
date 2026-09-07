package queries

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewCreateOrder создание запроса для создания заказа.
func NewCreateOrder(orderID string, userID int32, status entities.OrderStatus) postgres.InsertStatement {
	return table.Orders.
		INSERT(
			table.Orders.OrderID,
			table.Orders.UserID,
			table.Orders.Status,
		).
		VALUES(
			orderID,
			postgres.Int32(userID),
			hydrateDomainToOrdersStatusEnum(status),
		).
		RETURNING(table.Orders.ID.AS("id"))
}
