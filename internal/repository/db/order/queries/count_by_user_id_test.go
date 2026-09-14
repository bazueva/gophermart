package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_NewCountByUserID(t *testing.T) {
	t.Parallel()

	t.Run("count by user ID only", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
		}

		stmt := NewCountByUserID(filter)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COUNT(orders.id) AS "count" FROM public.orders WHERE orders.user_id = $1::integer;`
		expectedArgs := []interface{}{int32(123)}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("count by user ID and statuses", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusNew,
				entities.OrdersStatusProcessed,
			},
		}

		stmt := NewCountByUserID(filter)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COUNT(orders.id) AS "count" FROM public.orders WHERE (orders.user_id = $1::integer) AND (orders.status IN ('NEW', 'PROCESSED'));`
		expectedArgs := []interface{}{
			int32(123),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("count by user ID and order type (add balance)", func(t *testing.T) {
		addType := entities.OrderFilterAddBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &addType,
		}

		stmt := NewCountByUserID(filter)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COUNT(orders.id) AS "count" FROM public.orders WHERE (orders.user_id = $1::integer) AND ((orders.bonus_sum >= $2) OR (orders.bonus_sum IS NULL));`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("count by user ID and order type (write off balance)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &writeOffType,
		}

		stmt := NewCountByUserID(filter)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COUNT(orders.id) AS "count" FROM public.orders WHERE (orders.user_id = $1::integer) AND (orders.bonus_sum < $2);`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("count by user ID, statuses and order type (write off)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusProcessed,
			},
			OrderType: &writeOffType,
		}

		stmt := NewCountByUserID(filter)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COUNT(orders.id) AS "count" FROM public.orders WHERE ((orders.user_id = $1::integer) AND (orders.status IN ('PROCESSED'))) AND (orders.bonus_sum < $2);`
		expectedArgs := []interface{}{
			int32(123),
			float64(0),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
