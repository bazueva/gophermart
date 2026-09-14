package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_NewCreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("create new order", func(t *testing.T) {
		orderID := "1234567890"
		userID := int32(42)
		status := entities.OrdersStatusNew

		stmt := NewCreateOrder(orderID, userID, status)
		sql, args := stmt.Sql()

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status) VALUES ($1, $2::integer, 'NEW') RETURNING orders.id AS "id";`
		expectedArgs := []interface{}{
			"1234567890",
			int32(42),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})

	t.Run("create processed order", func(t *testing.T) {
		orderID := "9876543210"
		userID := int32(100)
		status := entities.OrdersStatusProcessed

		stmt := NewCreateOrder(orderID, userID, status)
		sql, args := stmt.Sql()

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status) VALUES ($1, $2::integer, 'PROCESSED') RETURNING orders.id AS "id";`
		expectedArgs := []interface{}{
			"9876543210",
			int32(100),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
