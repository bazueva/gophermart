package context

import "context"

type userIDContextKey struct{}

// WithUserID добавляет userID в контекст.
func WithUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext извлекает userID из контекста.
func UserIDFromContext(ctx context.Context) (int32, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int32)

	return userID, ok
}
