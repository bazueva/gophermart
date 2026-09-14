package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_NewExistLogin(t *testing.T) {
	t.Parallel()

	t.Run("check login exists", func(t *testing.T) {
		login := "test-user"

		stmt := NewExistLogin(login)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT (EXISTS (SELECT users.login AS "users.login" FROM public.users WHERE users.login = $1::text)) AS "exists";`
		expectedArgs := []interface{}{
			login,
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
