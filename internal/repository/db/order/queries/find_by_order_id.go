package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewFindByOrderID создание запроса для поиска заказа по идентификатору заказа.
func NewFindByOrderID(orderID string) postgres.SelectStatement {
	return postgres.SELECT(
		table.Orders.ID,
		table.Orders.OrderID,
		table.Orders.Status,
		table.Orders.UserID,
		table.Orders.CreatedAt,
	).FROM(table.Orders).
		WHERE(table.Orders.OrderID.EQ(postgres.String(orderID)))
}
