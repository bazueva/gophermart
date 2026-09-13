package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewUserBalanceSum(t *testing.T) {
	t.Parallel()

	t.Run("get user balance sum", func(t *testing.T) {
		userID := int32(123)

		stmt := NewUserBalanceSum(userID)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COALESCE(SUM(locked_orders."orders.bonus_sum"),$1) AS "sum" 
FROM (
	SELECT orders.bonus_sum AS "orders.bonus_sum" 
	FROM public.orders 
	WHERE (orders.user_id = $2::integer) AND (orders.status = 'PROCESSED') 
		FOR UPDATE
	) AS locked_orders;`
		expectedArgs := []interface{}{float64(0), userID}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
