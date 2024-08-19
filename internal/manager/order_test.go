package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOrderManager_Register(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() orderRepository
		want       *entity.Order
		wantErr    error
	}{
		{
			name: "valid register",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				order := &entity.Order{
					ID:     1,
					UserID: 11,
					Status: entity.OrderStatusNew,
				}
				When(repository.CreateOrFind(
					AnyContext(),
					Equal(order),
				)).ThenReturn(order, true, nil).
					Verify(Once())

				return repository
			},
			want: &entity.Order{
				ID:     1,
				UserID: 11,
				Status: entity.OrderStatusNew,
			},
		},
		{
			name: "already registered by current user",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				When(repository.CreateOrFind(
					AnyContext(),
					Equal(&entity.Order{
						ID:     2,
						UserID: 22,
						Status: entity.OrderStatusNew,
					}),
				)).ThenReturn(&entity.Order{
					ID:     2,
					UserID: 22,
					Status: entity.OrderStatusProcessed,
				}, false, nil).
					Verify(Once())

				return repository
			},
			want: &entity.Order{
				ID:     2,
				UserID: 22,
				Status: entity.OrderStatusProcessed,
			},
			wantErr: ErrOrderAlreadyRegisteredByCurrentUser,
		},
		{
			name: "already registered by another user",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				When(repository.CreateOrFind(
					AnyContext(),
					Equal(&entity.Order{
						ID:     3,
						UserID: 33,
						Status: entity.OrderStatusNew,
					}),
				)).ThenReturn(&entity.Order{
					ID:     3,
					UserID: 303,
					Status: entity.OrderStatusInvalid,
				}, false, nil).
					Verify(Once())

				return repository
			},
			wantErr: ErrOrderAlreadyRegisteredByAnotherUser,
		},
		{
			name: "db error",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				When(repository.CreateOrFind(
					AnyContext(),
					Equal(&entity.Order{
						ID:     4,
						UserID: 44,
						Status: entity.OrderStatusNew,
					}),
				)).ThenReturn(nil, false, someErr).
					Verify(Once())

				return repository
			},
			wantErr: someErr,
		},
	}
	for id, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id++
			SetUp(t)
			repository := tt.repository()
			manager := NewOrderManager(repository)
			order, err := manager.Register(context.Background(), uint64(id), uint32(id*11))
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, order)
		})
	}
}

func TestOrderManager_HasUser(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() orderRepository
		want       bool
		wantErr    error
	}{
		{
			name: "yes",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](1),
				)).ThenReturn(&entity.Order{}, nil).
					Verify(Once())

				return repository
			},
			want: true,
		},
		{
			name: "no",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](2),
				)).ThenReturn(nil, nil).
					Verify(Once())

				return repository
			},
			want: false,
		},
		{
			name: "error",
			repository: func() orderRepository {
				repository := Mock[orderRepository]()
				WhenDouble(repository.FindOneByUserID(
					AnyContext(),
					Exact[uint32](3),
				)).ThenReturn(nil, someErr).
					Verify(Once())

				return repository
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
			manager := NewOrderManager(repository)

			got, err := manager.HasUser(context.Background(), uint32(id))

			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
