package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models/forms"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	ExistLogin(ctx context.Context, login string) (bool, *entities.DomainError)
	CreateUser(ctx context.Context, user entities.User) (int32, *entities.DomainError)
	FindByLogin(ctx context.Context, login string) (entities.User, *entities.DomainError)
}

// UserService Сервис пользователей
type UserService struct {
	repository    Repository
	formValidator *validator.Validate
	logger        interfaces.Logger
	secretKey     string
}

// CheckJWTToken проверяет валидность JWT-токена и возвращает идентификатор пользователя.
func (u *UserService) CheckJWTToken(token string) (int32, *entities.DomainError) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.secretKey), nil
	})
	if err != nil {
		return 0, entities.NewUnauthorizedError(err, "токен недействителен")
	}

	if !parsedToken.Valid {
		return 0, entities.NewUnauthorizedError(nil, "токен недействителен")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, entities.NewUnauthorizedError(nil, "токен недействителен")
	}

	userIDRaw, exists := claims["user_id"]
	if !exists {
		return 0, entities.NewUnauthorizedError(nil, "пользователь не авторизован")
	}

	userID := cast.ToInt32(userIDRaw)
	if userID == 0 {
		return 0, entities.NewUnauthorizedError(nil, "пользователь не авторизован")
	}

	return userID, nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT-токен.
func (u *UserService) Login(ctx context.Context, loginForm forms.LoginForm) (string, *entities.DomainError) {
	errDomain := u.validateForm(loginForm)
	if errDomain != nil {
		return "", errDomain
	}

	user, errDomain := u.repository.FindByLogin(ctx, loginForm.Login)
	if errDomain != nil {
		return "", errDomain
	}

	if user.ID == 0 {
		return "", entities.NewUnauthorizedError(nil, "неверная пара логин/пароль")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginForm.Password))
	if err != nil {
		return "", entities.NewUnauthorizedError(nil, "неверная пара логин/пароль")
	}

	token, err := u.generateJWTToken(user.ID)
	if err != nil {
		u.logger.Error("error generateJWTToken", zap.Error(err))

		return "", entities.NewInternalServerError(err, "")
	}

	return token, nil
}

// Register выполняет регистрацию пользователя и возвращает JWT-токен.
func (u *UserService) Register(ctx context.Context, userForm forms.UserForm) (string, *entities.DomainError) {
	errDomain := u.validateRegister(ctx, userForm)
	if errDomain != nil {
		return "", errDomain
	}

	userID, errDomain := u.createUser(ctx, userForm)
	if errDomain != nil {
		return "", errDomain
	}

	token, err := u.generateJWTToken(userID)
	if err != nil {
		u.logger.Error("error generateJWTToken", zap.Error(err))

		return "", entities.NewInternalServerError(err, "")
	}

	return token, nil
}

// ValidateForm выполняет валидацию формы.
func (u *UserService) validateForm(form interface{}) *entities.DomainError {
	err := u.formValidator.Struct(form)

	validationErrors := entities.ConvertValidatorErrors(err)
	if len(validationErrors) > 0 {
		errorsResult := lo.Map(validationErrors, func(item entities.ValidateError, index int) string {
			return fmt.Sprintf("%s: %s", item.Field, item.Error)
		})

		return entities.NewBadRequestError(nil, strings.Join(errorsResult, ","))
	}

	return nil
}

// ValidateRegister выполняет валидацию формы регистрации.
func (u *UserService) validateRegister(ctx context.Context, userForm forms.UserForm) *entities.DomainError {
	errDomain := u.validateForm(userForm)
	if errDomain != nil {
		return errDomain
	}

	exists, errDomain := u.checkUniqueLogin(ctx, userForm.Login)
	if errDomain != nil {
		return errDomain
	}

	if exists {
		return entities.NewConflictError(nil, "Такой login уже существует")
	}

	return nil
}

// CheckUniqueLogin проверяет уникальность логина.
func (u *UserService) checkUniqueLogin(ctx context.Context, login string) (bool, *entities.DomainError) {
	exist, err := u.repository.ExistLogin(ctx, login)
	if err != nil {
		return false, err
	}

	return exist, nil
}

// CreateUser создает нового пользователя.
func (u *UserService) createUser(ctx context.Context, userForm forms.UserForm) (int32, *entities.DomainError) {
	hashPass, err := hashPassword(userForm.Password)
	if err != nil {
		u.logger.Error("error hash password", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	userID, errDomain := u.repository.CreateUser(ctx, entities.User{
		Login:        userForm.Login,
		PasswordHash: hashPass,
	})
	if errDomain != nil {
		return 0, errDomain
	}

	return userID, nil
}

// GenerateJWTToken генерирует JWT-токен для пользователя.
func (u *UserService) generateJWTToken(userID int32) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.secretKey))

	return tokenString, err
}

// NewUserService создает новый экземпляр сервиса пользователей.
func NewUserService(repository Repository, logger interfaces.Logger, secretKey string) *UserService {
	service := &UserService{
		logger:        logger,
		repository:    repository,
		formValidator: validator.New(validator.WithRequiredStructEnabled()),
		secretKey:     secretKey,
	}

	return service
}

// HashPassword хеширует пароль.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
