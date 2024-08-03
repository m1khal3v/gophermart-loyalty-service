package context

import "context"

const userIDKey = "userID"

func WithUserId(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIdFromContext(ctx context.Context) (uint32, bool) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return 0, false
	}

	userId, ok := value.(uint32)

	return userId, ok
}
