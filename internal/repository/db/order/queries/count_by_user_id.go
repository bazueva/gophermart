package queries

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewCountByUserID создание запроса для подсчета количества заказов по идентификатору пользователя.
func NewCountByUserID(filter entities.OrderFilter) postgres.SelectStatement {
	return postgres.SELECT(
		postgres.COUNT(table.Orders.ID).AS("count"),
	).FROM(table.Orders).
		WHERE(buildOrderFilterCondition(filter))
}
