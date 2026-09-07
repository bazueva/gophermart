package user

import (
	"context"
	"errors"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/repository/db/user/queries"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/qrm"
	"go.uber.org/zap"
)

type repository struct {
	db     interfaces.DB
	logger interfaces.Logger
}

// FindByLogin поиск пользователя по логину.
func (r *repository) FindByLogin(ctx context.Context, login string) (entities.User, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result model.Users

	err := queries.NewFindByLogin(login).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		r.logger.Error("error repository FindByLoginPassword", zap.Error(err))

		return entities.User{}, entities.NewInternalServerError(err, "")
	}

	return entities.User{
		ID:           result.ID,
		Login:        result.Login,
		PasswordHash: result.PasswordHash,
	}, nil
}

const (
	// defaultTimeout таймаут для выполнения запросов к базе данных.
	defaultTimeout = 1 * time.Second
)

// CreateUser создает нового пользователя.
func (r *repository) CreateUser(ctx context.Context, user entities.User) (int32, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	query := table.Users.
		INSERT(table.Users.Login, table.Users.PasswordHash).
		VALUES(user.Login, user.PasswordHash).
		RETURNING(table.Users.ID.AS("id"))

	var result struct {
		ID int32
	}
	err := query.QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil {
		r.logger.Error("error repository CreateUser", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	return result.ID, nil
}

// ExistLogin проверяет наличие пользователя с указанным логином.
func (r *repository) ExistLogin(ctx context.Context, login string) (bool, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var response struct {
		Exists bool
	}
	err := queries.NewExistLogin(login).
		QueryContext(ctxWithTimeout, r.db, &response)
	if err != nil {
		r.logger.Error("error ExistLogin", zap.Error(err))

		return false, entities.NewInternalServerError(err, "")
	}

	return response.Exists, nil
}

// NewRepository создает новый репозиторий для работы с пользователями.
func NewRepository(db interfaces.DB, logger interfaces.Logger) *repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}
