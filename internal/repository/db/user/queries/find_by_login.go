package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewFindByLogin создает запрос для поиска пользователя по логину.
func NewFindByLogin(login string) postgres.SelectStatement {
	return postgres.
		SELECT(
			table.Users.ID,
			table.Users.Login,
			table.Users.PasswordHash,
		).
		FROM(table.Users).
		WHERE(postgres.AND(
			table.Users.Login.EQ(postgres.String(login)),
		)).
		LIMIT(1)
}
