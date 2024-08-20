package semaphore

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		max       uint64
		wantPanic bool
	}{
		{
			name: "valid",
			max:  10,
		},
		{
			name:      "invalid",
			max:       0,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					New(tt.max)
				})
			} else {
				semaphore := New(tt.max)
				assert.Len(t, semaphore.channel, 0)
			}
		})
	}
}

func TestSemaphore(t *testing.T) {
	semaphore := New(1)
	ctx := context.Background()
	require.NoError(t, semaphore.Acquire(ctx))
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	require.Error(t, semaphore.Acquire(cancelCtx))
	semaphore.Release()
	require.NoError(t, semaphore.Acquire(ctx))
}
