package queries

import (
	"testing"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateStatusAndBonus(t *testing.T) {
	t.Parallel()

	t.Run("update status and bonus for PROCESSED order", func(t *testing.T) {
		orderID := "12345678903"
		status := entities.OrdersStatusProcessed
		bonusSum := 150.50

		stmt := NewUpdateStatusAndBonus(orderID, status, bonusSum, nil)
		sql, args := stmt.Sql()

		expectedSQL := `UPDATE public.orders 
SET (bonus_sum,updated_at,next_check_at,processed_at,status) = ($1,$2,$3,$4,$5) 
WHERE orders.order_id = $6::text;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Len(t, args, 6)
		assert.Equal(t, bonusSum, args[0])
		assert.IsType(t, time.Time{}, args[1]) // UpdatedAt
		assert.Nil(t, args[2])                 // next_check_at
		assert.IsType(t, time.Time{}, args[3]) // ProcessedAt
		assert.Equal(t, model.OrdersStatus("PROCESSED"), args[4])
		assert.Equal(t, orderID, args[5])
	})

	t.Run("update status and bonus for NEW order (without processed_at)", func(t *testing.T) {
		orderID := "12345678903"
		status := entities.OrdersStatusNew
		bonusSum := 0.0

		stmt := NewUpdateStatusAndBonus(orderID, status, bonusSum, nil)
		sql, args := stmt.Sql()

		expectedSQL := `UPDATE public.orders 
SET (bonus_sum,updated_at,next_check_at,status) = ($1,$2,$3,$4)
WHERE orders.order_id = $5::text;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Len(t, args, 5)
		assert.Equal(t, bonusSum, args[0])
		assert.IsType(t, time.Time{}, args[1]) // UpdatedAt
		assert.Nil(t, args[2])                 // next_check_at
		assert.Equal(t, model.OrdersStatus("NEW"), args[3])
		assert.Equal(t, orderID, args[4])

	})

	t.Run("update status and bonus for NEW order with next_check_at", func(t *testing.T) {
		orderID := "12345678903"
		status := entities.OrdersStatusNew
		bonusSum := 0.0

		stmt := NewUpdateStatusAndBonus(orderID, status, bonusSum, new(time.Now()))
		sql, args := stmt.Sql()

		expectedSQL := `UPDATE public.orders 
SET (bonus_sum,updated_at,next_check_at,status) = ($1,$2,$3,$4)
WHERE orders.order_id = $5::text;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Len(t, args, 5)
		assert.Equal(t, bonusSum, args[0])
		assert.IsType(t, time.Time{}, args[1]) // UpdatedAt
		assert.IsType(t, time.Time{}, args[2]) // next_check_at
		assert.Equal(t, model.OrdersStatus("NEW"), args[3])
		assert.Equal(t, orderID, args[4])
	})
}
