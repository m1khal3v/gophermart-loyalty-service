package balance

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
)

func TestContainer_Balance(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		manager     func() userManager
		status      int
		response    *responses.Balance
		errResponse *responses.APIError
	}{
		{
			name: "valid balance",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() userManager {
				manager := Mock[userManager]()
				WhenDouble(manager.FindByID(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(&entity.User{
					Balance:   money.New(123.32),
					Withdrawn: money.New(321.12),
				}, nil).
					Verify(Once())

				return manager
			},
			status: http.StatusOK,
			response: &responses.Balance{
				Current:   123.32,
				Withdrawn: 321.12,
			},
		},
		{
			name: "cant get credentials",
			ctx:  context.Background(),
			manager: func() userManager {
				return Mock[userManager]()
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get request credentials",
			},
		},
		{
			name: "cant get user",
			ctx:  userContext.WithUserID(context.Background(), 123),
			manager: func() userManager {
				manager := Mock[userManager]()
				WhenDouble(manager.FindByID(
					AnyContext(),
					Exact(uint32(123)),
				)).ThenReturn(nil, errors.New("db unavailable")).
					Verify(Once())

				return manager
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(manager, Mock[userWithdrawalManager]())
			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil).WithContext(tt.ctx)

			container.Balance(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

			if tt.response != nil {
				response := &responses.Balance{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
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
