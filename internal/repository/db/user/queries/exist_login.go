package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

// NewExistLogin создает запрос для проверки наличия пользователя с указанным логином.
func NewExistLogin(login string) postgres.SelectStatement {
	return postgres.SELECT(
		postgres.EXISTS(
			postgres.SELECT(
				table.Users.Login,
			).
				FROM(table.Users).
				WHERE(table.Users.Login.EQ(postgres.String(login))),
		).AS("exists"),
	)
}
