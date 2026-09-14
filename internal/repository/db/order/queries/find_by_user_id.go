package queries

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewFindByUserID создание запроса для поиска заказов по фильтру.
func NewFindByUserID(filter entities.OrderFilter, limit, offset int64) postgres.SelectStatement {
	query := postgres.SELECT(
		table.Orders.ID,
		table.Orders.OrderID,
		table.Orders.Status,
		table.Orders.UserID,
		table.Orders.CreatedAt,
		table.Orders.ProcessedAt,
		table.Orders.BonusSum,
	).FROM(table.Orders).
		WHERE(buildOrderFilterCondition(filter)).
		ORDER_BY(table.Orders.CreatedAt.DESC()).
		LIMIT(limit).
		OFFSET(offset)

	return query
}
