package withdrawal

import (
	"context"
	"encoding/json"
	"errors"
	userContext "github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
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
	tests := []struct {
		name            string
		ctx             context.Context
		manager         func() withdrawalManager
		verify          func(manager withdrawalManager)
		status          int
		listResponse    []responses.Withdrawal
		messageResponse *responses.Message
		errResponse     *responses.APIError
	}{
		{
			name: "valid withdrawals",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() withdrawalManager {
				channel := make(chan *entity.Withdrawal, 4)
				for i := 1; i <= 4; i++ {
					channel <- &entity.Withdrawal{
						OrderID:   uint64(i),
						UserID:    123,
						Sum:       money.New(float64(i) * 1.11),
						CreatedAt: time.Now().Round(time.Duration(i) * time.Minute),
					}
				}
				close(channel)
				manager := Mock[withdrawalManager]()
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
			verify: func(manager withdrawalManager) {
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
			listResponse: []responses.Withdrawal{
				{
					Order:       1,
					Sum:         1.11,
					ProcessedAt: time.Now().Round(time.Minute),
				},
				{
					Order:       2,
					Sum:         2.22,
					ProcessedAt: time.Now().Round(2 * time.Minute),
				},
				{
					Order:       3,
					Sum:         3.33,
					ProcessedAt: time.Now().Round(3 * time.Minute),
				},
				{
					Order:       4,
					Sum:         4.44,
					ProcessedAt: time.Now().Round(4 * time.Minute),
				},
			},
		},
		{
			name: "no withdrawals",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() withdrawalManager {
				manager := Mock[withdrawalManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(false, nil)

				return manager
			},
			verify: func(manager withdrawalManager) {
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
				Message: "withdrawals not found",
			},
		},
		{
			name: "cant get credentials",
			ctx:  context.Background(),
			manager: func() withdrawalManager {
				return Mock[withdrawalManager]()
			},
			verify: func(manager withdrawalManager) {
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
			name: "cant check user has withdrawals",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() withdrawalManager {
				manager := Mock[withdrawalManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(false, errors.New("some error"))

				return manager
			},
			verify: func(manager withdrawalManager) {
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
				Message: "can`t check that user has withdrawals",
			},
		},
		{
			name: "cant get user withdrawals",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() withdrawalManager {
				manager := Mock[withdrawalManager]()
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
			verify: func(manager withdrawalManager) {
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
				Message: "can`t get user withdrawals",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(manager)
			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawal", nil).WithContext(tt.ctx)

			container.List(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

			if tt.listResponse != nil {
				response := []responses.Withdrawal{}
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
