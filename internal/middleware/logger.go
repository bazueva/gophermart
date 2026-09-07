package middleware

import (
	"net/http"
	"time"

	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// ServerLogger middleware для логирования запросов.
func ServerLogger(logger interfaces.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			end := time.Since(start)

			logger.Info("request data",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("time", end),
				zap.Int("statusCode", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
			)
		}

		return http.HandlerFunc(fn)
	}
}
