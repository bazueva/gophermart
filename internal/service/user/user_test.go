package user

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/models/forms"
	"github.com/bazueva/gofermart/internal/service/user/mocks"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func Test_validate(t *testing.T) {
	t.Parallel()

	t.Run("empty login", func(t *testing.T) {
		t.Parallel()

		service := NewUserService(nil, nil, "")

		user := forms.UserForm{}

		err := service.validateRegister(t.Context(), user)

		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.BadRequestErrorType,
			Text:      "login: Поле обязательно для заполнения,password: Поле обязательно для заполнения",
		}, err)
	})

	t.Run("min rule", func(t *testing.T) {
		t.Parallel()

		service := NewUserService(nil, nil, "")

		user := forms.UserForm{
			Login:    "1",
			Password: "2",
		}

		err := service.validateRegister(t.Context(), user)

		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.BadRequestErrorType,
			Text:      "login: Поле должно содержать не менее 4 символов,password: Поле должно содержать не менее 8 символов",
		}, err)
	})

	t.Run("max rule", func(t *testing.T) {
		t.Parallel()

		service := NewUserService(nil, nil, "")

		user := forms.UserForm{
			Login:    strings.Repeat("1", 21),
			Password: strings.Repeat("2", 33),
		}

		err := service.validateRegister(t.Context(), user)

		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.BadRequestErrorType,
			Text:      "login: Поле должно содержать не более 20 символов,password: Поле должно содержать не более 32 символов",
		}, err)
	})

	t.Run("user exists", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			ExistLogin(mock.Anything, "existinglogin").
			Return(true, nil)

		service := NewUserService(mockRepo, nil, "")

		user := forms.UserForm{
			Login:    "existinglogin",
			Password: "validpassword123",
		}

		err := service.validateRegister(t.Context(), user)

		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.ConflictErrorType,
			Text:      "Такой login уже существует",
		}, err)
	})
}

func Test_createUser(t *testing.T) {
	t.Parallel()

	t.Run("success create user", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			CreateUser(mock.Anything, mock.Anything).
			Return(int32(19), nil)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "")

		userForm := forms.UserForm{
			Login:    "testuser",
			Password: "validpassword123",
		}

		userID, err := service.createUser(t.Context(), userForm)
		assert.Equal(t, int32(19), userID)

		assert.Nil(t, err)
	})

	t.Run("hash password error", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		logger := interfacesMocks.NewMockLogger(t)
		logger.EXPECT().
			Error("error hash password", mock.Anything)

		service := NewUserService(mockRepo, logger, "")

		userForm := forms.UserForm{
			Login:    "testuser",
			Password: strings.Repeat("a", 73),
		}

		userID, err := service.createUser(t.Context(), userForm)

		assert.Equal(t, int32(0), userID)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.InternalServerErrorType,
			Text:      "Internal Server Error",
			SourceErr: errors.New("bcrypt: password length exceeds 72 bytes"),
		}, err)
	})

	t.Run("repository returns domain error", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			CreateUser(mock.Anything, mock.Anything).
			Return(int32(0), entities.NewConflictError(nil, "login already exists"))

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "")

		userForm := forms.UserForm{
			Login:    "existinguser",
			Password: "validpassword123",
		}

		userID, err := service.createUser(t.Context(), userForm)

		assert.Equal(t, int32(0), userID)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.ConflictErrorType,
			Text:      "login already exists",
		}, err)
	})
}

func TestUserService_generateJWTToken(t *testing.T) {
	t.Parallel()

	secretKey := "test-secret-key"
	service := &UserService{
		secretKey: secretKey,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		userID := int32(123)
		token, err := service.generateJWTToken(userID)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(userID), claims["user_id"])
		assert.NotEmpty(t, claims["exp"])
	})

	t.Run("token contains correct user_id", func(t *testing.T) {
		t.Parallel()

		userID := int32(456)
		token, err := service.generateJWTToken(userID)
		require.NoError(t, err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		require.NoError(t, err)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(456), claims["user_id"])
	})

	t.Run("token expires after 24 hours", func(t *testing.T) {
		t.Parallel()

		userID := int32(789)
		token, err := service.generateJWTToken(userID)
		require.NoError(t, err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		require.NoError(t, err)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)

		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)

		expTime := time.Unix(int64(exp), 0)
		expectedTime := time.Now().Add(24 * time.Hour)
		diff := expTime.Sub(expectedTime)
		assert.Less(t, diff.Abs(), 5*time.Second)
	})

	t.Run("different tokens for different users", func(t *testing.T) {
		t.Parallel()

		token1, err := service.generateJWTToken(1)
		require.NoError(t, err)

		token2, err := service.generateJWTToken(2)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})
}

func TestUserService_Login(t *testing.T) {
	t.Parallel()

	correctPassword := "correctPassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success login", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			FindByLogin(mock.Anything, "testuser").
			Return(entities.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
			}, nil)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "testuser",
			Password: correctPassword,
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Nil(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("empty login", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		logger := interfacesMocks.NewMockLogger(t)
		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "",
			Password: "password123",
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.BadRequestErrorType,
			Text:      "login: Поле обязательно для заполнения",
		}, err)
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		logger := interfacesMocks.NewMockLogger(t)
		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "testuser",
			Password: "",
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.BadRequestErrorType,
			Text:      "password: Поле обязательно для заполнения",
		}, err)
	})

	t.Run("error findLogin", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			FindByLogin(mock.Anything, "nonexistent").
			Return(entities.User{}, entities.NewInternalServerError(nil, ""))

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "nonexistent",
			Password: "password123",
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.InternalServerErrorType,
			Text:      "Internal Server Error",
		}, err)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			FindByLogin(mock.Anything, "nonexistent").
			Return(entities.User{}, nil)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "nonexistent",
			Password: "password123",
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.UnauthorizedErrorType,
			Text:      "неверная пара логин/пароль",
		}, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			FindByLogin(mock.Anything, "testuser").
			Return(entities.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
			}, nil)

		logger := interfacesMocks.NewMockLogger(t)
		service := NewUserService(mockRepo, logger, "test-secret-key")

		loginForm := forms.LoginForm{
			Login:    "testuser",
			Password: "wrongPassword",
		}

		token, err := service.Login(t.Context(), loginForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.UnauthorizedErrorType,
			Text:      "неверная пара логин/пароль",
		}, err)
	})
}

func TestUserService_CheckJWTToken(t *testing.T) {
	t.Parallel()

	secretKey := "test-secret-key"

	t.Run("valid token returns user_id", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"user_id": float64(123),
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Nil(t, errDomain)
		assert.Equal(t, int32(123), userID)
	})

	t.Run("expired token returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"user_id": float64(123),
			"exp":     time.Now().Add(-1 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Equal(t, int32(0), userID)
		assert.NotNil(t, errDomain)
		assert.Equal(t, entities.UnauthorizedErrorType, errDomain.ErrorType)
		assert.Equal(t, "токен недействителен", errDomain.Text)
	})

	t.Run("invalid signature returns unauthorized error", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"user_id": float64(123),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("wrong-secret-key"))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Equal(t, int32(0), userID)
		assert.NotNil(t, errDomain)
		assert.Equal(t, entities.UnauthorizedErrorType, errDomain.ErrorType)
	})

	t.Run("malformed token returns unauthorized error", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		userID, errDomain := service.CheckJWTToken("invalid-token-format")

		assert.Equal(t, int32(0), userID)
		assert.NotNil(t, errDomain)
		assert.Equal(t, entities.UnauthorizedErrorType, errDomain.ErrorType)
		assert.Equal(t, "токен недействителен", errDomain.Text)
	})

	t.Run("missing user_id claim returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Equal(t, int32(0), userID)
		assert.NotNil(t, errDomain)
		assert.Equal(t, entities.UnauthorizedErrorType, errDomain.ErrorType)
		assert.Equal(t, "пользователь не авторизован", errDomain.Text)
	})

	t.Run("user_id is zero returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"user_id": float64(0),
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Equal(t, int32(0), userID)
		assert.NotNil(t, errDomain)
		assert.Equal(t, entities.UnauthorizedErrorType, errDomain.ErrorType)
		assert.Equal(t, "пользователь не авторизован", errDomain.Text)
	})

	t.Run("user_id as string", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, mockLogger, secretKey)

		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		userID, errDomain := service.CheckJWTToken(tokenString)

		assert.Equal(t, int32(123), userID)
		assert.Nil(t, errDomain)
	})
}

func TestUserService_Register(t *testing.T) {
	t.Parallel()

	t.Run("success register", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		mockRepo.EXPECT().
			ExistLogin(mock.Anything, "testuser").
			Return(false, nil)

		mockRepo.EXPECT().
			CreateUser(mock.Anything, mock.MatchedBy(func(user entities.User) bool {
				return user.Login == "testuser" &&
					user.PasswordHash != ""
			})).
			Return(int32(19), nil)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret")

		userForm := forms.UserForm{
			Login:    "testuser",
			Password: "validpassword123",
		}

		token, err := service.Register(t.Context(), userForm)

		assert.NotEmpty(t, token)
		assert.Nil(t, err)

		userID, errDomain := service.CheckJWTToken(token)

		assert.Nil(t, errDomain)
		assert.Equal(t, int32(19), userID)
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)
		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret")

		userForm := forms.UserForm{
			Login:    "",
			Password: "validpassword123",
		}

		token, err := service.Register(t.Context(), userForm)

		assert.Empty(t, token)
		require.NotNil(t, err)
		assert.Equal(t, entities.BadRequestErrorType, err.ErrorType)
	})

	t.Run("error - login already exists", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		mockRepo.EXPECT().
			ExistLogin(mock.Anything, "existinguser").
			Return(true, nil)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret")

		userForm := forms.UserForm{
			Login:    "existinguser",
			Password: "validpassword123",
		}

		token, err := service.Register(t.Context(), userForm)

		assert.Empty(t, token)
		assert.Equal(t, &entities.DomainError{
			ErrorType: entities.ConflictErrorType,
			Text:      "Такой login уже существует",
		}, err)
	})

	t.Run("error - repository check unique login", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		expectedErr := entities.NewInternalServerError(
			errors.New("repository error"),
			"",
		)

		mockRepo.EXPECT().
			ExistLogin(mock.Anything, "testuser").
			Return(false, expectedErr)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret")

		userForm := forms.UserForm{
			Login:    "testuser",
			Password: "validpassword123",
		}

		token, err := service.Register(t.Context(), userForm)

		assert.Empty(t, token)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("error - create user", func(t *testing.T) {
		t.Parallel()

		mockRepo := mocks.NewMockRepository(t)

		expectedErr := entities.NewConflictError(
			nil,
			"login already exists",
		)

		mockRepo.EXPECT().
			ExistLogin(mock.Anything, "testuser").
			Return(false, nil)

		mockRepo.EXPECT().
			CreateUser(
				mock.Anything,
				mock.MatchedBy(func(user entities.User) bool {
					return user.Login == "testuser" &&
						user.PasswordHash != ""
				}),
			).
			Return(int32(0), expectedErr)

		logger := interfacesMocks.NewMockLogger(t)

		service := NewUserService(mockRepo, logger, "test-secret")

		userForm := forms.UserForm{
			Login:    "testuser",
			Password: "validpassword123",
		}

		token, err := service.Register(t.Context(), userForm)

		assert.Empty(t, token)
		assert.Equal(t, expectedErr, err)
	})
}
