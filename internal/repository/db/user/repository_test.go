package user

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	dbPkg "github.com/bazueva/gofermart/internal/repository/db"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRepository_ExistLogin(t *testing.T) {
	t.Parallel()

	t.Run("exist login", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).
				AddRow(true))

		exists, err := repo.ExistLogin(ctx, login)

		assert.Nil(t, err)
		assert.True(t, exists)
	})

	t.Run("not exist login", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).
				AddRow(false))

		exists, err := repo.ExistLogin(ctx, login)

		assert.Nil(t, err)
		assert.False(t, exists)
	})

	t.Run("error", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("ошибка БД")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error ExistLogin", mock2.Anything)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnError(errorDB)

		exists, err := repo.ExistLogin(ctx, login)

		assert.Equal(t, "Internal Server Error", err.Error())
		assert.False(t, exists)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error ExistLogin", mock2.Anything)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		expectedSQL := `SELECT (EXISTS (
                  SELECT users.login AS "users.login"
                  FROM public.users
                  WHERE users.login = $1::text
             )) AS "exists";`

		login := "testlogin"

		mock.ExpectQuery(expectedSQL).
			WithArgs(login).
			WillReturnError(context.DeadlineExceeded)

		exists, err := repo.ExistLogin(ctx, login)

		assert.False(t, exists)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)
		assert.Equal(t, "jet: context canceled", err.(*entities.DomainError).SourceErr.Error())
	})
}

func TestRepository_CreateUser(t *testing.T) {
	t.Run("success create user", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		user := entities.User{
			Login:        "testuser",
			PasswordHash: "password",
		}

		expectedSQL := `INSERT INTO public.users (login, password_hash) VALUES ($1, $2) RETURNING users.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.PasswordHash).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(100))

		userID, err := repo.CreateUser(ctx, user)

		assert.Nil(t, err)
		assert.Equal(t, int32(100), userID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error repository", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateUser", mock2.Anything)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		user := entities.User{
			Login:        "testuser",
			PasswordHash: "password",
		}

		expectedSQL := `INSERT INTO public.users (login, password_hash) VALUES ($1, $2) RETURNING users.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.PasswordHash).
			WillReturnError(errors.New("connection refused"))

		userID, err := repo.CreateUser(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateUser", mock2.Anything)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		user := entities.User{
			Login:        "testuser",
			PasswordHash: "hashed_password",
		}

		expectedSQL := `INSERT INTO public.users (login, password) VALUES ($1, $2);`

		mock.ExpectQuery(expectedSQL).
			WithArgs(user.Login, user.PasswordHash).
			WillReturnError(context.DeadlineExceeded)

		userID, err := repo.CreateUser(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, err.(*entities.DomainError).ErrorType, entities.InternalServerErrorType)
		assert.Equal(t, "jet: context canceled", err.(*entities.DomainError).SourceErr.Error())
	})
}

func TestRepository_FindByLogin(t *testing.T) {
	t.Parallel()

	t.Run("success - user found", func(t *testing.T) {
		t.Parallel()

		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer db.Close()

		logger := mocks.NewMockLogger(t)
		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT users.id AS "users.id",
             users.login AS "users.login",
             users.password_hash AS "users.password_hash"
        FROM public.users
        WHERE (users.login = $1::text)
        LIMIT $2;`

		mock.ExpectQuery(expectedSQL).
			WithArgs("testuser", 1).
			WillReturnRows(sqlmock.NewRows([]string{"users.id", "users.login", "users.password_hash"}).
				AddRow(123, "testuser", "hashed_password"))

		user, domainErr := repo.FindByLogin(ctx, "testuser")

		assert.Nil(t, domainErr)
		assert.Equal(t, int32(123), user.ID)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "hashed_password", user.PasswordHash)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found - returns empty user", func(t *testing.T) {
		t.Parallel()

		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer db.Close()

		logger := mocks.NewMockLogger(t)
		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT users.id AS "users.id",
             users.login AS "users.login",
             users.password_hash AS "users.password_hash"
        FROM public.users
        WHERE (users.login = $1::text)
        LIMIT $2;`

		mock.ExpectQuery(expectedSQL).
			WithArgs("nonexistent", 1).
			WillReturnRows(sqlmock.NewRows([]string{"users.id", "users.login", "users.password_hash"}))

		user, domainErr := repo.FindByLogin(ctx, "nonexistent")

		assert.Nil(t, domainErr)
		assert.Equal(t, int32(0), user.ID)
		assert.Empty(t, user.Login)
		assert.Empty(t, user.PasswordHash)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error - returns internal error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := helpers.SQLMockTest(t)
		require.NoError(t, err)
		defer db.Close()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindByLoginPassword", mock2.Anything)

		dbWrapper := dbPkg.NewSQLDBWrapper(db)
		repo := NewRepository(dbWrapper, logger)
		ctx := t.Context()

		expectedSQL := `SELECT users.id AS "users.id",
             users.login AS "users.login",
             users.password_hash AS "users.password_hash"
        FROM public.users
        WHERE (users.login = $1::text)
        LIMIT $2;`

		mock.ExpectQuery(expectedSQL).
			WithArgs("testuser", 1).
			WillReturnError(errors.New("connection refused"))

		user, domainErr := repo.FindByLogin(ctx, "testuser")

		assert.Error(t, domainErr)
		assert.Equal(t, entities.InternalServerErrorType, domainErr.ErrorType)
		assert.Equal(t, int32(0), user.ID)
		assert.Empty(t, user.Login)
		assert.Empty(t, user.PasswordHash)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
