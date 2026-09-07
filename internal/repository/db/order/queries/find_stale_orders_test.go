package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewFindStaleOrders(t *testing.T) {
	t.Parallel()

	t.Run("find stale orders with one status", func(t *testing.T) {
		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessed,
		}
		limit := int64(10)

		stmt := NewFindStaleOrders(statuses, limit)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.order_id AS "orders.order_id" 
FROM public.orders 
WHERE ((orders.status IN ('NEW','PROCESSED')) AND (orders.created_at < $1::timestamp with time zone)) 
  AND ((orders.next_check_at IS NULL) OR (orders.next_check_at < $2::timestamp with time zone)) 
ORDER BY orders.created_at ASC 
LIMIT $3;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))

		assert.Len(t, args, 3)
		assert.Equal(t, int64(10), args[2])
	})
}
