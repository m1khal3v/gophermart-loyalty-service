package context

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		userID uint32
		ok     bool
	}{
		{
			name:   "valid",
			ctx:    context.WithValue(context.Background(), userIDKey, uint32(123)),
			userID: 123,
			ok:     true,
		},
		{
			name:   "invalid type",
			ctx:    context.WithValue(context.Background(), userIDKey, int32(123)),
			userID: 0,
			ok:     false,
		},
		{
			name:   "invalid key",
			ctx:    context.WithValue(context.Background(), "invalidID", uint32(123)),
			userID: 0,
			ok:     false,
		},
		{
			name:   "empty context",
			ctx:    context.Background(),
			userID: 0,
			ok:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, ok := UserIDFromContext(tt.ctx)
			assert.Equalf(t, tt.userID, userID, "UserIDFromContext(%v)", tt.ctx)
			assert.Equalf(t, tt.ok, ok, "UserIDFromContext(%v)", tt.ctx)
		})
	}
}

func TestWithUserID(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		userID uint32
	}{
		{
			name:   "valid",
			ctx:    context.Background(),
			userID: 123,
		},
		{
			name:   "valid replace",
			ctx:    WithUserID(context.Background(), uint32(123)),
			userID: 321,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(tt.ctx, tt.userID)
			assert.Equalf(t, tt.userID, ctx.Value(userIDKey), "WithUserID(%v)", tt.userID)
		})
	}
}
