package balance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	userContext "github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	managers "github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	_ "github.com/m1khal3v/gophermart-loyalty-service/pkg/validator"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContainer_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		contentType   string
		requestString string
		request       requests.Withdraw
		manager       func() userWithdrawalManager
		status        int
		response      *responses.Message
		errResponse   *responses.APIError
	}{
		{
			name:        "valid withdraw",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				manager := Mock[userWithdrawalManager]()
				When(manager.Withdraw(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
					Exact(123.321),
				)).ThenReturn(nil).
					Verify(Once())

				return manager
			},
			status: http.StatusOK,
			response: &responses.Message{
				Message: "withdrawal successfully registered",
			},
		},
		{
			name:        "cant get credentials",
			ctx:         context.Background(),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get request credentials",
			},
		},
		{
			name:        "invalid content-type",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "invalid",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid Content-Type",
			},
		},
		{
			name:          "invalid json",
			ctx:           userContext.WithUserID(context.Background(), 123),
			contentType:   "application/json",
			requestString: "{order: 123456, sum: 123.321}",
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid json received",
			},
		},
		{
			name:          "invalid request",
			ctx:           userContext.WithUserID(context.Background(), 123),
			contentType:   "application/json",
			requestString: `{"order": "123456", "money": 123.321}`,
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:        "invalid luhn",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 123456,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:        "invalid sum",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   0,
			},
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:        "invalid sum 2",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   -1.1,
			},
			manager: func() userWithdrawalManager {
				return Mock[userWithdrawalManager]()
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:        "insufficient funds",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				manager := Mock[userWithdrawalManager]()
				When(manager.Withdraw(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
					Exact(123.321),
				)).ThenReturn(managers.ErrInsufficientFunds).
					Verify(Once())

				return manager
			},
			status: http.StatusPaymentRequired,
			errResponse: &responses.APIError{
				Code:    http.StatusPaymentRequired,
				Message: "insufficient funds",
			},
		},
		{
			name:        "internal server error",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			request: requests.Withdraw{
				Order: 1234566,
				Sum:   123.321,
			},
			manager: func() userWithdrawalManager {
				manager := Mock[userWithdrawalManager]()
				When(manager.Withdraw(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
					Exact(123.321),
				)).ThenReturn(errors.New("some error")).
					Verify(Once())

				return manager
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t register withdrawal",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(Mock[userManager](), manager)
			recorder := httptest.NewRecorder()

			var requestBody *bytes.Buffer
			if tt.requestString != "" {
				requestBody = bytes.NewBuffer([]byte(tt.requestString))
			} else {
				jsonRequest, err := json.Marshal(tt.request)
				require.NoError(t, err)
				requestBody = bytes.NewBuffer(jsonRequest)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", requestBody).WithContext(tt.ctx)
			request.Header.Set("Content-Type", tt.contentType)

			container.Withdraw(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

			if tt.response != nil {
				response := &responses.Message{}
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
