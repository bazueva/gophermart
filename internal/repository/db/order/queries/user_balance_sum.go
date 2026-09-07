package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/enum"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewUserBalanceSum создание запроса для получения суммы бонусов пользователя.
func NewUserBalanceSum(userID int32) postgres.SelectStatement {
	return postgres.SELECT(
		postgres.COALESCE(
			postgres.SUM(table.Orders.BonusSum),
			postgres.Float(0),
		).AS("sum"),
	).FROM(table.Orders).
		WHERE(
			table.Orders.UserID.EQ(postgres.Int32(userID)).
				AND(table.Orders.Status.EQ(enum.OrdersStatus.Processed)),
		)
}
