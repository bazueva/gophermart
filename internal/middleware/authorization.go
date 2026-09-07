package middleware

import (
	"net/http"
	"strings"

	"github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
)

type CheckerJWTToken interface {
	CheckJWTToken(token string) (int32, *entities.DomainError)
}

// Authorization middleware для проверки авторизации пользователя.
func Authorization(checkerToken CheckerJWTToken, logger interfaces.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "пользователь не авторизован", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "невалидные данные авторизации", http.StatusBadRequest)
				return
			}

			if parts[1] == "" {
				http.Error(w, "не указаны данные для авторизации", http.StatusBadRequest)
				return
			}

			userID, err := checkerToken.CheckJWTToken(parts[1])
			if err != nil {
				if err.ErrorType == entities.UnauthorizedErrorType {
					http.Error(w, err.Text, http.StatusUnauthorized)
					return
				}

				http.Error(w, err.Text, http.StatusInternalServerError)
				return
			}

			auth := context.WithUserID(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(auth))
		}

		return http.HandlerFunc(fn)
	}
}
