package retryafter

import (
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		want       time.Duration
	}{
		{
			name:       "success date",
			retryAfter: time.Now().Round(time.Hour).Add(time.Hour * 10).Format(time.RFC1123),
			want:       time.Until(time.Now().Round(time.Hour).Add(time.Hour * 10)),
		},
		{
			name:       "success second",
			retryAfter: "30",
			want:       30 * time.Second,
		},
		{
			name:       "failure",
			retryAfter: "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultValue := time.Second * time.Duration(rand.Uint64N(1000))
			got := Parse(tt.retryAfter, defaultValue)
			if tt.want > 0 {
				assert.Equal(t, tt.want.Nanoseconds()/time.Second.Nanoseconds(), got.Nanoseconds()/time.Second.Nanoseconds())
			} else {
				assert.Equal(t, defaultValue, got)
			}
		})
	}
}

func Test_parseHTTPDate(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		want       time.Time
		wantErr    bool
	}{
		{
			name:       "success",
			retryAfter: time.Now().UTC().Format(time.RFC1123),
			want:       time.Now().UTC(),
		},
		{
			name:       "failure",
			retryAfter: "abc",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHTTPDate(tt.retryAfter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.Round(time.Second*2), got.Round(time.Second*2))
		})
	}
}

func Test_parseSeconds(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		want       time.Duration
		wantErr    bool
	}{
		{
			name:       "success",
			retryAfter: "30",
			want:       30 * time.Second,
		},
		{
			name:       "failure",
			retryAfter: "abc",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSeconds(tt.retryAfter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
