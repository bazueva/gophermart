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
	errorsPkg "github.com/pkg/errors"
)

type repository struct {
	db     interfaces.DB
}

// FindByLogin поиск пользователя по логину.
func (r *repository) FindByLogin(ctx context.Context, login string) (entities.User, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result model.Users

	err := queries.NewFindByLogin(login).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		return entities.User{}, entities.NewInternalServerError(errorsPkg.Wrap(err, "error repository FindByLoginPassword"), "")
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
		return 0, entities.NewInternalServerError(errorsPkg.Wrap(err, "error repository CreateUser"), "")
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
		return false, entities.NewInternalServerError(errorsPkg.Wrap(err, "error ExistLogin"), "")
	}

	return response.Exists, nil
}

// NewRepository создает новый репозиторий для работы с пользователями.
func NewRepository(db interfaces.DB) *repository {
	return &repository{
		db:     db,
	}
}
