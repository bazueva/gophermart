package queries

import (
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/samber/lo"
)

// NewFindStaleOrders создание запроса для поиска заказов, которые необходимо обработать.
func NewFindStaleOrders(statuses []entities.OrderStatus, limit int64) postgres.SelectStatement {
	staleThreshold := time.Now().Add(-2 * time.Minute)

	statusesArgs := lo.Map(statuses, func(item entities.OrderStatus, index int) postgres.Expression {
		return hydrateDomainToOrdersStatusEnum(item)
	})

	return postgres.
		SELECT(table.Orders.OrderID).
		FROM(table.Orders).
		WHERE(
			table.Orders.Status.IN(statusesArgs...).
				AND(
					table.Orders.CreatedAt.LT(postgres.TimestampzT(staleThreshold)),
				).AND(
				table.Orders.NextCheckAt.IS_NULL().
					OR(
						table.Orders.NextCheckAt.LT(
							postgres.TimestampzT(time.Now()),
						),
					),
			),
		).
		ORDER_BY(table.Orders.CreatedAt.ASC()).
		LIMIT(limit)
}
