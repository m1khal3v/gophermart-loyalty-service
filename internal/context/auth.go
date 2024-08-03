package context

import "context"

func WithUserId(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, "userId", id)
}

func UserIdFromContext(ctx context.Context) (uint32, bool) {
	value := ctx.Value("userId")
	if value == nil {
		return 0, false
	}

	userId, ok := value.(uint32)

	return userId, ok
}
