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
		name        string
		ctx         context.Context
		manager     func() withdrawalManager
		status      int
		response    []responses.Withdrawal
		errResponse *responses.APIError
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
						CreatedAt: time.Unix(int64(i), int64(i)).UTC(),
					}
				}
				close(channel)
				manager := Mock[withdrawalManager]()
				WhenDouble(manager.HasUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(true, nil).
					Verify(Once())
				WhenDouble(manager.FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(channel, nil).
					Verify(Once())

				return manager
			},
			status: http.StatusOK,
			response: []responses.Withdrawal{
				{
					Order:       1,
					Sum:         1.11,
					ProcessedAt: time.Unix(1, 1).UTC(),
				},
				{
					Order:       2,
					Sum:         2.22,
					ProcessedAt: time.Unix(2, 2).UTC(),
				},
				{
					Order:       3,
					Sum:         3.33,
					ProcessedAt: time.Unix(3, 3).UTC(),
				},
				{
					Order:       4,
					Sum:         4.44,
					ProcessedAt: time.Unix(4, 4).UTC(),
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
				)).ThenReturn(false, nil).
					Verify(Once())

				return manager
			},
			status:   http.StatusNoContent,
			response: []responses.Withdrawal{},
		},
		{
			name: "cant get credentials",
			ctx:  context.Background(),
			manager: func() withdrawalManager {
				return Mock[withdrawalManager]()
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
				)).ThenReturn(false, errors.New("some error")).
					Verify(Once())

				return manager
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
				)).ThenReturn(true, nil).
					Verify(Once())
				WhenDouble(manager.FindByUser(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(nil, errors.New("some error")).
					Verify(Once())

				return manager
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

			if tt.response != nil {
				response := []responses.Withdrawal{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				assert.Equal(t, tt.response, response)
			}

			if tt.errResponse != nil {
				response := &responses.APIError{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
				assert.Equal(t, tt.errResponse, response)
			}
		})
	}
}
