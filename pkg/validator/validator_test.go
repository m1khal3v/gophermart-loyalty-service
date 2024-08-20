package validator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsLuhn(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "valid string",
			value: "1234566",
			want:  true,
		},
		{
			name:  "invalid string",
			value: "123456",
			want:  false,
		},
		{
			name:  "valid int",
			value: 1234566,
			want:  true,
		},
		{
			name:  "invalid int",
			value: 123456,
			want:  false,
		},
		{
			name:  "negative int",
			value: -1234566,
			want:  false,
		},
		{
			name:  "valid uint",
			value: uint(1234566),
			want:  true,
		},
		{
			name:  "invalid uint",
			value: uint(123456),
			want:  false,
		},
		{
			name:  "incorrect string",
			value: "abc",
			want:  false,
		},
		{
			name:  "incorrect float",
			value: 123.123,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsLuhn(tt.value), "IsLuhn(%v)", tt.value)
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "valid string",
			value: "123",
			want:  true,
		},
		{
			name:  "invalid string",
			value: "-123",
			want:  false,
		},
		{
			name:  "incorrect string",
			value: "abc",
			want:  false,
		},
		{
			name:  "valid int",
			value: 123,
			want:  true,
		},
		{
			name:  "invalid int",
			value: -123,
			want:  false,
		},
		{
			name:  "valid uint",
			value: uint(123),
			want:  true,
		},
		{
			name:  "valid float",
			value: 123.123,
			want:  true,
		},
		{
			name:  "invalid float",
			value: -123.123,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsPositive(tt.value), "IsPositive(%v)", tt.value)
		})
	}
}
