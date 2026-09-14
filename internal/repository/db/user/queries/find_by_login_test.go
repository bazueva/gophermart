package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/stretchr/testify/assert"
)

func Test_NewFindByLogin(t *testing.T) {
	t.Parallel()

	t.Run("find user by login", func(t *testing.T) {
		login := "test-user"

		stmt := NewFindByLogin(login)
		sql, args := stmt.Sql()

		expectedSQL := `SELECT users.id AS "users.id",
       users.login AS "users.login",
       users.password_hash AS "users.password_hash" 
FROM public.users 
WHERE (users.login = $1::text) 
LIMIT $2;
		`
		expectedArgs := []interface{}{
			login,
			int64(1),
		}

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Equal(t, expectedArgs, args)
	})
}
