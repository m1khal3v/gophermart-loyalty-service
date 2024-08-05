package queue

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueue(t *testing.T) {
	tests := []struct {
		name  string
		items []bool
	}{
		{
			"empty",
			[]bool{},
		},
		{
			"not empty",
			[]bool{
				true,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := uint64(len(tt.items)) * 3
			queue := New[bool](count)
			channel := make(chan bool, count/3)
			for _, item := range tt.items {
				queue.Push(item)
				channel <- item
			}
			close(channel)
			queue.PushBatch(tt.items)
			queue.PushChannel(channel)

			require.Equal(t, count, queue.Count())
			items := make([]bool, 0, count)
			items = append(items, queue.Pop(count)...)
			expectedItems := make([]bool, 0, count)
			expectedItems = append(expectedItems, tt.items...)
			expectedItems = append(expectedItems, tt.items...)
			expectedItems = append(expectedItems, tt.items...)
			require.Equal(t, expectedItems, items)
			require.Equal(t, uint64(0), queue.Count())
			queue.PushBatch(tt.items)
			err := queue.RemoveBatch(2, func(items []bool) error {
				for _, item := range items {
					if !item {
						return errors.New("do not remove")
					}
				}

				return nil
			})
			if len(tt.items) > 0 {
				require.Error(t, err)
				require.Equal(t, uint64(2), queue.Count())
			} else {
				require.NoError(t, err)
				require.Equal(t, uint64(0), queue.Count())
			}
			require.NoError(t, queue.RemoveBatch(2, func(items []bool) error {
				return nil
			}))
			require.Equal(t, uint64(0), queue.Count())
		})
	}
}
