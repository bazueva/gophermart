package interfaces

import (
	"context"
	"database/sql"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DB interface {
	Executor
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type Tx interface {
	Executor

	Commit() error
	Rollback() error
}
