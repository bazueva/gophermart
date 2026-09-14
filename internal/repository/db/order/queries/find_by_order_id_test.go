package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewFindByOrderID(t *testing.T) {
	t.Parallel()

	t.Run("find by existing order ID", func(t *testing.T) {
		orderID := "12345678903"
		stmt := NewFindByOrderID(orderID)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT orders.id AS "orders.id", 
       orders.order_id AS "orders.order_id",
       orders.status AS "orders.status",
       orders.user_id AS "orders.user_id",
       orders.created_at AS "orders.created_at" 
FROM public.orders 
WHERE orders.order_id = $1::text;`
		expectedArgs := []interface{}{orderID}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
