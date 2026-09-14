package db

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
)

// PGErrorClassification представляет классификацию ошибок PostgreSQL.
type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota

	Retriable
)

// PostgresErrorClassifier классифицирует ошибки PostgreSQL.
type PostgresErrorClassifier struct {
}

// NewPostgresErrorClassifier создает новый классификатор ошибок PostgreSQL.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// ClassifyRetry классифицирует ошибку PostgreSQL для повторной попытки.
func (pe *PostgresErrorClassifier) ClassifyRetry(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var connectErr *pgconn.ConnectError
	if errors.As(err, &connectErr) {
		return Retriable
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return ClassifyPgError(pgError)
	}

	return NonRetriable
}

// ClassifyPgError классифицирует ошибку PostgreSQL по коду SQLSTATE.
func ClassifyPgError(err *pgconn.PgError) PGErrorClassification {
	if pgerrcode.IsConnectionException(err.Code) {
		return Retriable
	}

	if lo.Contains([]string{
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.AdminShutdown,
		pgerrcode.TooManyConnections,
		pgerrcode.QueryCanceled,
		pgerrcode.IOError,
	}, err.Code) {
		return Retriable
	}

	return NonRetriable
}
