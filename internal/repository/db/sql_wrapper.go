package db

import (
	"context"
	"database/sql"

	"github.com/bazueva/gofermart/internal/interfaces"
)

// SQLDBWrapper оборачивает стандартное подключение к базе данных *sql.DB
// и предоставляет интерфейс для работы с базой данных.
type SQLDBWrapper struct {
	*sql.DB
}

// NewSQLDBWrapper создает обертку над подключением к базе данных.
func NewSQLDBWrapper(db *sql.DB) interfaces.DB {
	return &SQLDBWrapper{db}
}

// BeginTx начинает транзакцию с указанными параметрами.
func (w *SQLDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (interfaces.Tx, error) {
	tx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
