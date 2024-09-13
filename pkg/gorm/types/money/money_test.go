package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		want     float64
		wantUint uint64
	}{
		{
			name:     "zero",
			amount:   0,
			want:     0,
			wantUint: 0,
		},
		{
			name:     "without round",
			amount:   123.43,
			want:     123.43,
			wantUint: 12343,
		},
		{
			name:     "with round down",
			amount:   123.433,
			want:     123.43,
			wantUint: 12343,
		},
		{
			name:     "with round up",
			amount:   123.436,
			want:     123.44,
			wantUint: 12344,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money := New(tt.amount)
			assert.Equal(t, tt.want, money.AsFloat())
			assert.Equal(t, tt.wantUint, uint64(money))
		})
	}
}
