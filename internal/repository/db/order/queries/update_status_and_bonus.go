package queries

import (
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewUpdateStatusAndBonus создание запроса для обновления статуса заказа и бонусов.
func NewUpdateStatusAndBonus(
	orderID string,
	status entities.OrderStatus,
	bonusSum float64,
	nextCheckAt *time.Time,
) postgres.UpdateStatement {
	now := time.Now()

	columns := postgres.ColumnList{
		table.Orders.BonusSum,
		table.Orders.UpdatedAt,
		table.Orders.NextCheckAt,
	}

	if status == entities.OrdersStatusProcessed {
		columns = append(columns, table.Orders.ProcessedAt)
	}

	if status != "" {
		columns = append(columns, table.Orders.Status)
	}

	model := dbModel.Orders{
		Status:      hydrateDomainToOrdersStatus(status),
		BonusSum:    new(bonusSum),
		ProcessedAt: &now,
		UpdatedAt:   &now,
		NextCheckAt: nextCheckAt,
	}

	return table.Orders.
		UPDATE(columns).
		MODEL(model).
		WHERE(table.Orders.OrderID.EQ(postgres.String(orderID)))
}
