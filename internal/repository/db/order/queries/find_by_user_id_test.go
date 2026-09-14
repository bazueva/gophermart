package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewFindByUserID(t *testing.T) {
	t.Parallel()

	t.Run("find by user ID only", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
		}
		limit := int64(20)
		offset := int64(0)

		stmt := NewFindByUserID(filter, limit, offset)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id", 
       orders.status AS "orders.status", 
       orders.user_id AS "orders.user_id", 
       orders.created_at AS "orders.created_at", 
       orders.processed_at AS "orders.processed_at", 
       orders.bonus_sum AS "orders.bonus_sum" 
FROM public.orders 
WHERE orders.user_id = $1::integer 
ORDER BY orders.created_at DESC 
LIMIT $2 
    OFFSET $3;`
		expectedArgs := []interface{}{int32(123), limit, offset}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("find by user ID and statuses", func(t *testing.T) {
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusNew,
				entities.OrdersStatusProcessed,
			},
		}
		limit := int64(10)
		offset := int64(20)

		stmt := NewFindByUserID(filter, limit, offset)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id", 
       orders.status AS "orders.status", 
       orders.user_id AS "orders.user_id", 
       orders.created_at AS "orders.created_at", 
       orders.processed_at AS "orders.processed_at", 
       orders.bonus_sum AS "orders.bonus_sum" 
FROM public.orders 
WHERE (orders.user_id = $1::integer) AND (orders.status IN ('NEW', 'PROCESSED')) 
ORDER BY orders.created_at DESC 
LIMIT $2 
    OFFSET $3;
`
		expectedArgs := []interface{}{int32(123), limit, offset}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("find by user ID and order type (add balance)", func(t *testing.T) {
		addType := entities.OrderFilterAddBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &addType,
		}
		limit := int64(20)
		offset := int64(0)

		stmt := NewFindByUserID(filter, limit, offset)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id", 
       orders.status AS "orders.status", 
       orders.user_id AS "orders.user_id", 
       orders.created_at AS "orders.created_at", 
       orders.processed_at AS "orders.processed_at", 
       orders.bonus_sum AS "orders.bonus_sum"
FROM public.orders 
WHERE (orders.user_id = $1::integer) AND ((orders.bonus_sum >= $2) OR (orders.bonus_sum IS NULL)) 
ORDER BY orders.created_at DESC 
LIMIT $3 
    OFFSET $4;
`
		expectedArgs := []interface{}{int32(123), float64(0), limit, offset}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("find by user ID and order type (write off balance)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID:    123,
			OrderType: &writeOffType,
		}
		limit := int64(20)
		offset := int64(0)

		stmt := NewFindByUserID(filter, limit, offset)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id", 
       orders.status AS "orders.status", 
       orders.user_id AS "orders.user_id", 
       orders.created_at AS "orders.created_at", 
       orders.processed_at AS "orders.processed_at", 
       orders.bonus_sum AS "orders.bonus_sum"
FROM public.orders 
WHERE (orders.user_id = $1::integer) AND (orders.bonus_sum < $2)
ORDER BY orders.created_at DESC 
LIMIT $3 
    OFFSET $4;
`
		expectedArgs := []interface{}{int32(123), float64(0), limit, offset}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("find by user ID, statuses and order type (write off)", func(t *testing.T) {
		writeOffType := entities.OrderFilterWriteOffBalanceType
		filter := entities.OrderFilter{
			UserID: 123,
			Statuses: []entities.OrderStatus{
				entities.OrdersStatusProcessed,
			},
			OrderType: &writeOffType,
		}
		limit := int64(15)
		offset := int64(30)

		stmt := NewFindByUserID(filter, limit, offset)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id", 
       orders.status AS "orders.status", 
       orders.user_id AS "orders.user_id", 
       orders.created_at AS "orders.created_at", 
       orders.processed_at AS "orders.processed_at", 
       orders.bonus_sum AS "orders.bonus_sum"
FROM public.orders 
WHERE ((orders.user_id = $1::integer) AND (orders.status IN ('PROCESSED'))) AND (orders.bonus_sum < $2) 
ORDER BY orders.created_at DESC 
LIMIT $3 
    OFFSET $4;
`
		expectedArgs := []interface{}{int32(123), float64(0), limit, offset}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
