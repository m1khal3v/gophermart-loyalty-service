package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	managers "github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContainer_Login(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		requestString string
		request       requests.Auth
		manager       func() UserManager
		verify        func(manager UserManager)
		status        int
		token         string
		response      *responses.Auth
		errResponse   *responses.APIError
	}{
		{
			name:        "valid login",
			contentType: "application/json",
			request: requests.Auth{
				Login:    "ivan_ivanov",
				Password: "$uP3R$3cR3t",
			},
			manager: func() UserManager {
				manager := Mock[UserManager]()
				WhenDouble(manager.Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)).ThenReturn("t0k3n", nil)

				return manager
			},
			verify: func(manager UserManager) {
				Verify(manager, Once()).Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)
			},
			status: http.StatusOK,
			token:  "Bearer t0k3n",
			response: &responses.Auth{
				AccessToken: "t0k3n",
			},
		},
		{
			name:        "invalid content-type",
			contentType: "invalid",
			request: requests.Auth{
				Login:    "ivan_ivanov",
				Password: "$uP3R$3cR3t",
			},
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid Content-Type",
			},
		},
		{
			name:          "invalid json",
			contentType:   "application/json",
			requestString: "{login: ivan_ivanov, password: $uP3R$3cR3t}",
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid json received",
			},
		},
		{
			name:          "invalid request",
			contentType:   "application/json",
			requestString: `{"username": "ivan_ivanov", "password": "$uP3R$3cR3t"}`,
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:          "invalid password length",
			contentType:   "application/json",
			requestString: `{"login": "ivan_ivanov", "password": "$uP3"}`,
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:          "invalid login length",
			contentType:   "application/json",
			requestString: `{"login": "ok", "password": "$uP3rS3cr3t"}`,
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:          "invalid login symbol",
			contentType:   "application/json",
			requestString: `{"login": "ivan=ivanov", "password": "$uP3rS3cr3t"}`,
			manager: func() UserManager {
				return Mock[UserManager]()
			},
			verify: func(manager UserManager) {
				Verify(manager, Never()).Authorize(
					AnyContext(),
					AnyString(),
					AnyString(),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request received",
			},
		},
		{
			name:        "invalid credentials",
			contentType: "application/json",
			request: requests.Auth{
				Login:    "ivan_ivanov",
				Password: "$uP3R$3cR3t",
			},
			manager: func() UserManager {
				manager := Mock[UserManager]()
				WhenDouble(manager.Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)).ThenReturn("", managers.ErrInvalidCredentials)

				return manager
			},
			verify: func(manager UserManager) {
				Verify(manager, Once()).Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)
			},
			status: http.StatusUnauthorized,
			errResponse: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
		},
		{
			name:        "internal server error",
			contentType: "application/json",
			request: requests.Auth{
				Login:    "ivan_ivanov",
				Password: "$uP3R$3cR3t",
			},
			manager: func() UserManager {
				manager := Mock[UserManager]()
				WhenDouble(manager.Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)).ThenReturn("", errors.New("some error"))

				return manager
			},
			verify: func(manager UserManager) {
				Verify(manager, Once()).Authorize(
					AnyContext(),
					Exact("ivan_ivanov"),
					Exact("$uP3R$3cR3t"),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(manager)
			recorder := httptest.NewRecorder()

			var requestBody *bytes.Buffer
			if tt.requestString != "" {
				requestBody = bytes.NewBuffer([]byte(tt.requestString))
			} else {
				jsonRequest, err := json.Marshal(tt.request)
				require.NoError(t, err)
				requestBody = bytes.NewBuffer(jsonRequest)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/user/login", requestBody)
			request.Header.Set("Content-Type", tt.contentType)

			container.Login(recorder, request)

			require.Equal(t, tt.status, recorder.Code)
			assert.Equal(t, tt.token, recorder.Header().Get("Authorization"))

			if tt.response != nil {
				response := &responses.Auth{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
				assert.Equal(t, tt.response, response)
			}

			if tt.errResponse != nil {
				response := &responses.APIError{}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), response))
				assert.Equal(t, tt.errResponse, response)
			}

			tt.verify(manager)
		})
	}
}
