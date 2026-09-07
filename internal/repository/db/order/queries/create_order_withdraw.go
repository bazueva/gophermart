package queries

import (
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewCreateOrderWithdraw создание запроса для создания заказа со списанием бонусов.
func NewCreateOrderWithdraw(
	orderID string,
	userID int32,
	bonusSum float64,
	status entities.OrderStatus,
) postgres.InsertStatement {
	return table.Orders.
		INSERT(
			table.Orders.OrderID,
			table.Orders.UserID,
			table.Orders.Status,
			table.Orders.BonusSum,
			table.Orders.ProcessedAt,
		).
		VALUES(
			orderID,
			postgres.Int32(userID),
			hydrateDomainToOrdersStatusEnum(status),
			postgres.Double(bonusSum),
			postgres.TimestampzT(time.Now()),
		).
		RETURNING(table.Orders.ID.AS("id"))
}
