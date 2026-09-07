package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_NewCreateOrderWithdraw(t *testing.T) {
	t.Parallel()

	t.Run("create order withdraw with processed status", func(t *testing.T) {
		orderID := "777888999"
		userID := int32(15)
		bonusSum := -500.50
		status := entities.OrdersStatusProcessed

		stmt := NewCreateOrderWithdraw(orderID, userID, bonusSum, status)
		sql, args := stmt.Sql()

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status, bonus_sum, processed_at)
VALUES ($1, $2::integer, 'PROCESSED', $3::double precision, $4::timestamp with time zone) 
RETURNING orders.id AS "id";`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))

		assert.Len(t, args, 4)
		assert.Equal(t, "777888999", args[0])
		assert.Equal(t, int32(15), args[1])
		assert.Equal(t, -500.50, args[2])
	})

	t.Run("create order withdraw with processing status", func(t *testing.T) {
		orderID := "111222333"
		userID := int32(99)
		bonusSum := -10.00
		status := entities.OrdersStatusProcessing

		stmt := NewCreateOrderWithdraw(orderID, userID, bonusSum, status)
		sql, args := stmt.Sql()

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status, bonus_sum, processed_at)
VALUES ($1, $2::integer, 'PROCESSING', $3::double precision, $4::timestamp with time zone) 
RETURNING orders.id AS "id";`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))

		assert.Len(t, args, 4)
		assert.Equal(t, "111222333", args[0])
		assert.Equal(t, int32(99), args[1])
		assert.Equal(t, -10.00, args[2])
	})
}
