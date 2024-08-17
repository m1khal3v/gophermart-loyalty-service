package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithdrawalManager_HasUser(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() withdrawalRepository
		verify     func(repository withdrawalRepository)
		want       bool
		wantErr    error
	}{
		{
			name: "yes",
			repository: func() withdrawalRepository {
				repository := Mock[withdrawalRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](1),
				)).ThenReturn(&entity.Withdrawal{}, nil)

				return repository
			},
			verify: func(repository withdrawalRepository) {
				Verify(repository, Once()).FindOneByUserID(
					AnyContext(),
					Exact[uint32](1),
				)
			},
			want: true,
		},
		{
			name: "no",
			repository: func() withdrawalRepository {
				repository := Mock[withdrawalRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](2),
				)).ThenReturn(nil, nil)

				return repository
			},
			verify: func(repository withdrawalRepository) {
				Verify(repository, Once()).FindOneByUserID(
					AnyContext(),
					Exact[uint32](2),
				)
			},
			want: false,
		},
		{
			name: "error",
			repository: func() withdrawalRepository {
				repository := Mock[withdrawalRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](3),
				)).ThenReturn(nil, someErr)

				return repository
			},
			verify: func(repository withdrawalRepository) {
				Verify(repository, Once()).FindOneByUserID(
					AnyContext(),
					Exact[uint32](3),
				)
			},
			want:    false,
			wantErr: someErr,
		},
	}
	for id, tt := range tests {
		id++
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			repository := tt.repository()
			manager := NewWithdrawalManager(repository)

			got, err := manager.HasUser(context.Background(), uint32(id))

			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			tt.verify(repository)
		})
	}
}
