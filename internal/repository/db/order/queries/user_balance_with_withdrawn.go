package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/enum"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewUserBalanceWithWithdrawn создание запроса для получения информации о балансе и списанных бонусах пользователя.
func NewUserBalanceWithWithdrawn(userID int32) postgres.SelectStatement {
	sumWithdrawals := postgres.COALESCE(
		postgres.SUM(
			postgres.CASE().WHEN(table.Orders.BonusSum.LT(postgres.Float(0))).
				THEN(table.Orders.BonusSum).
				ELSE(postgres.Float(0)),
		),
		postgres.Float(0),
	)

	return postgres.SELECT(
		postgres.COALESCE(
			postgres.SUM(table.Orders.BonusSum),
			postgres.Float(0),
		).AS("balance"),
		postgres.FloatExp(sumWithdrawals).
			MUL(postgres.Float(-1)).
			AS("withdrawn"),
	).FROM(table.Orders).
		WHERE(
			table.Orders.UserID.EQ(postgres.Int32(userID)).
				AND(table.Orders.Status.EQ(enum.OrdersStatus.Processed)),
		)
}
