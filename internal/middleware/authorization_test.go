package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appCtx "github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/middleware/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthorization(t *testing.T) {
	t.Parallel()

	t.Run("success authorization", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		expectedUserID := int32(123)
		mockChecker.EXPECT().CheckJWTToken("valid-token").
			Return(expectedUserID, nil).Once()

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := appCtx.UserIDFromContext(r.Context())

			assert.Equal(t, expectedUserID, userID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockChecker.AssertExpectations(t)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "пользователь не авторизован")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Invalid")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "невалидные данные авторизации")
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "не указаны данные для авторизации")
	})

	t.Run("unauthorized error from checker", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		domainErr := entities.NewUnauthorizedError(nil, "токен недействителен")
		mockChecker.EXPECT().
			CheckJWTToken("invalid-token").
			Return(int32(0), domainErr).
			Once()

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "токен недействителен")
	})

	t.Run("internal server error from checker", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		domainErr := entities.NewInternalServerError(nil, "")
		mockChecker.EXPECT().
			CheckJWTToken("token-with-error").
			Return(int32(0), domainErr).
			Once()

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer token-with-error")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})

	t.Run("case insensitive Bearer", func(t *testing.T) {
		t.Parallel()

		mockChecker := mocks.NewMockCheckerJWTToken(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		expectedUserID := int32(456)
		mockChecker.EXPECT().
			CheckJWTToken("valid-token").
			Return(expectedUserID, nil).
			Once()

		handler := Authorization(mockChecker, mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := appCtx.UserIDFromContext(r.Context())

			assert.Equal(t, expectedUserID, userID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "bEaReR valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
