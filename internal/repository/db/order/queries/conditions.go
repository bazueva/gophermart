package queries

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/samber/lo"
)

// buildOrderFilterCondition построение условия фильтрации заказов.
func buildOrderFilterCondition(filter entities.OrderFilter) postgres.BoolExpression {
	condition := table.Orders.UserID.EQ(postgres.Int32(filter.UserID))

	if len(filter.Statuses) > 0 {
		statusesArgs := lo.Map(filter.Statuses, func(item entities.OrderStatus, _ int) postgres.Expression {
			return hydrateDomainToOrdersStatusEnum(item)
		})

		condition = condition.AND(table.Orders.Status.IN(statusesArgs...))
	}

	if filter.OrderType != nil {
		switch *filter.OrderType {
		case entities.OrderFilterAddBalanceType:
			condition = condition.AND(
				table.Orders.BonusSum.GT_EQ(postgres.Float(0)).
					OR(table.Orders.BonusSum.IS_NULL()),
			)
		case entities.OrderFilterWriteOffBalanceType:
			condition = condition.AND(
				table.Orders.BonusSum.LT(postgres.Float(0)),
			)
		}
	}

	return condition
}
