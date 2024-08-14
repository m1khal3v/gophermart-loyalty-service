package order

import (
	"context"
	"encoding/json"
	"errors"
	userContext "github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestContainer_List(t *testing.T) {
	accrual := 1.23
	tests := []struct {
		name            string
		ctx             context.Context
		manager         func() orderManager
		verify          func(manager orderManager)
		status          int
		listResponse    []responses.Order
		messageResponse *responses.Message
		errResponse     *responses.APIError
	}{
		{
			name: "valid orders",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() orderManager {
				channel := make(chan *entity.Order, 4)
				for i := 1; i <= 4; i++ {
					var status string
					switch i {
					case 1:
						status = entity.OrderStatusNew
					case 2:
						status = entity.OrderStatusProcessing
					case 3:
						status = entity.OrderStatusInvalid
					case 4:
						status = entity.OrderStatusProcessed
					}
					order := &entity.Order{
						ID:        uint64(i),
						UserID:    123,
						Status:    status,
						CreatedAt: time.Now().Round(time.Duration(i) * time.Minute),
						UpdatedAt: time.Now().Round(time.Duration(i) * time.Minute),
					}
					if status == entity.OrderStatusProcessed {
						order.Accrual = money.New(accrual)
					}
					channel <- order
				}
				close(channel)
				manager := Mock[orderManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(true, nil)
				WhenDouble(manager.FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(channel, nil)

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)
				Verify(manager, Once()).FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)
			},
			status: http.StatusOK,
			listResponse: []responses.Order{
				{
					Number:     1,
					Status:     entity.OrderStatusNew,
					UploadedAt: time.Now().Round(time.Minute),
				},
				{
					Number:     2,
					Status:     entity.OrderStatusProcessing,
					UploadedAt: time.Now().Round(2 * time.Minute),
				},
				{
					Number:     3,
					Status:     entity.OrderStatusInvalid,
					UploadedAt: time.Now().Round(3 * time.Minute),
				},
				{
					Number:     4,
					Status:     entity.OrderStatusProcessed,
					Accrual:    &accrual,
					UploadedAt: time.Now().Round(4 * time.Minute),
				},
			},
		},
		{
			name: "no orders",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(false, nil)

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)
				Verify(manager, Never()).FindByUser(
					AnyContext(),
					Any[uint32](),
				)
			},
			status: http.StatusNoContent,
			messageResponse: &responses.Message{
				Message: "orders not found",
			},
		},
		{
			name: "cant get credentials",
			ctx:  context.Background(),
			manager: func() orderManager {
				return Mock[orderManager]()
			},
			verify: func(manager orderManager) {
				Verify(manager, Never()).HasUser(
					AnyContext(),
					Any[uint32](),
				)
				Verify(manager, Never()).FindByUser(
					AnyContext(),
					Any[uint32](),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get request credentials",
			},
		},
		{
			name: "cant check user has orders",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(false, errors.New("some error"))

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)
				Verify(manager, Never()).FindByUser(
					AnyContext(),
					Any[uint32](),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t check that user has orders",
			},
		},
		{
			name: "cant get user orders",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(true, nil)
				WhenDouble(manager.FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(nil, errors.New("some error"))

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)
				Verify(manager, Once()).FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get user orders",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(manager, queue.New[uint64](1))
			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil).WithContext(tt.ctx)

			container.List(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

			if tt.listResponse != nil {
				response := []responses.Order{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				assert.Equal(t, tt.listResponse, response)
			}

			if tt.errResponse != nil {
				response := &responses.APIError{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
				assert.Equal(t, tt.errResponse, response)
			}

			if tt.messageResponse != nil {
				response := &responses.Message{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
				assert.Equal(t, tt.messageResponse, response)
			}

			tt.verify(manager)
		})
	}
}
