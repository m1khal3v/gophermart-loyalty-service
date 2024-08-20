package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserWithdrawalManager_Withdraw(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() userWithdrawalRepository
		wantErr    error
	}{
		{
			name: "ok",
			repository: func() userWithdrawalRepository {
				repository := Mock[userWithdrawalRepository]()
				WhenDouble(repository.Withdraw(
					AnyContext(),
					Exact[uint64](1),
					Exact[uint32](11),
					Exact(1.11),
				)).ThenReturn(&entity.Withdrawal{
					OrderID: 1,
					UserID:  1,
					Sum:     money.New(1.11),
				}, nil).
					Verify(Once())

				return repository
			},
		},
		{
			name: "insufficient funds",
			repository: func() userWithdrawalRepository {
				repository := Mock[userWithdrawalRepository]()
				WhenDouble(repository.Withdraw(
					AnyContext(),
					Exact[uint64](2),
					Exact[uint32](22),
					Exact(2.22),
				)).ThenReturn(nil, nil).
					Verify(Once())

				return repository
			},
			wantErr: ErrInsufficientFunds,
		},
		{
			name: "db error",
			repository: func() userWithdrawalRepository {
				repository := Mock[userWithdrawalRepository]()
				WhenDouble(repository.Withdraw(
					AnyContext(),
					Exact[uint64](3),
					Exact[uint32](33),
					Exact(3.33),
				)).ThenReturn(nil, someErr).
					Verify(Once())

				return repository
			},
			wantErr: someErr,
		},
	}
	for id, tt := range tests {
		id++
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			repository := tt.repository()
			manager := NewUserWithdrawalManager(repository)

			err := manager.Withdraw(context.Background(), uint64(id), uint32(id)*11, float64(id)*1.11)

			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
