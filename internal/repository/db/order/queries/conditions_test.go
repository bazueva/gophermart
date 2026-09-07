package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/stretchr/testify/assert"
)

func Test_buildOrderFilterCondition(t *testing.T) {
	t.Parallel()

	t.Run("filter by user ID only", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
		}

		condition := buildOrderFilterCondition(filter)
		stmt := postgres.SELECT(postgres.Raw("*")).FROM(table.Orders).WHERE(condition)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT * 
FROM public.orders 
WHERE orders.user_id = $1::integer;`
		expectedArgs := []interface{}{int32(123)}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("filter by user ID and statuses", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusNew,
				entities.OrdersStatusProcessed,
			},
		}

		condition := buildOrderFilterCondition(filter)
		stmt := postgres.SELECT(postgres.Raw("*")).FROM(table.Orders).WHERE(condition)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT * FROM public.orders WHERE (orders.user_id = $1::integer) AND (orders.status IN ('NEW', 'PROCESSED'));`
		expectedArgs := []interface{}{
			int32(123),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("filter by user ID and order type (add balance)", func(t *testing.T) {
		addType := entities.OrderFilterAddBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &addType,
		}

		condition := buildOrderFilterCondition(filter)
		stmt := postgres.SELECT(postgres.Raw("*")).FROM(table.Orders).WHERE(condition)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT * FROM public.orders WHERE (orders.user_id = $1::integer) AND ((orders.bonus_sum >= $2) OR (orders.bonus_sum IS NULL));`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("filter by user ID and order type (write off balance)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &writeOffType,
		}

		condition := buildOrderFilterCondition(filter)
		stmt := postgres.SELECT(postgres.Raw("*")).FROM(table.Orders).WHERE(condition)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT * FROM public.orders WHERE (orders.user_id = $1::integer) AND (orders.bonus_sum < $2);`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("filter by user ID, statuses and order type (write off)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusProcessed,
			},
			OrderType: &writeOffType,
		}

		condition := buildOrderFilterCondition(filter)
		stmt := postgres.SELECT(postgres.Raw("*")).FROM(table.Orders).WHERE(condition)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT * FROM public.orders WHERE ((orders.user_id = $1::integer) AND (orders.status IN ('PROCESSED'))) AND (orders.bonus_sum < $2);`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
