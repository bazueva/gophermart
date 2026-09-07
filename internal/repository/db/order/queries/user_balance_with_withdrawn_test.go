package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewUserBalanceWithWithdrawn(t *testing.T) {
	t.Parallel()

	t.Run("get user balance with withdraw", func(t *testing.T) {
		userID := int32(123)

		stmt := NewUserBalanceWithWithdrawn(userID)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT COALESCE(SUM(orders.bonus_sum), $1) AS "balance", 
       (COALESCE(
			SUM(
				(CASE WHEN orders.bonus_sum < $2 
					THEN orders.bonus_sum 
					ELSE $3 END)
			), 
			$4
       ) * $5) AS "withdrawn"
FROM public.orders 
WHERE (orders.user_id = $6::integer) AND (orders.status = 'PROCESSED');`
		expectedArgs := []interface{}{float64(0), float64(0), float64(0), float64(0), float64(-1), userID}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
