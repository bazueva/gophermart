package db

import (
	"context"

	"github.com/bazueva/gofermart/internal/interfaces"
)

// txContextKey ключ для хранения транзакции в контексте.
type txContextKey struct{}

// WithTx добавляет транзакцию в контекст.
func WithTx(ctx context.Context, tx interfaces.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext извлекает транзакцию из контекста.
func TxFromContext(ctx context.Context) (interfaces.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(interfaces.Tx)

	return tx, ok
}
