package db

import (
	"context"

	"github.com/bazueva/gofermart/internal/interfaces"
)

// TryLock пытается установить транзакционную advisory-блокировку для указанного ключа.
// Возвращает true, если блокировка успешно установлена, и false, если она уже занята.
func TryLock(
	ctx context.Context,
	executor interfaces.Executor,
	lockType int64,
	key int64,
) (bool, error) {
	var ok bool

	err := executor.QueryRowContext(
		ctx,
		"SELECT pg_try_advisory_xact_lock($1, $2);",
		lockType,
		key,
	).Scan(&ok)

	if err != nil {
		return false, err
	}

	return ok, nil
}
