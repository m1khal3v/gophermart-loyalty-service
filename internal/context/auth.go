package context

import "context"

type key string

const userIDKey key = "userID"

func WithUserID(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (uint32, bool) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return 0, false
	}

	userID, ok := value.(uint32)

	return userID, ok
}
