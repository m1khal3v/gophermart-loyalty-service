package order

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	userContext "github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	managers "github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContainer_Register(t *testing.T) {
	tests := []struct {
		name            string
		ctx             context.Context
		contentType     string
		requestBody     string
		manager         func() orderManager
		verify          func(manager orderManager)
		status          int
		messageResponse *responses.Message
		errResponse     *responses.APIError
	}{
		{
			name:        "valid order",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "text/plain",
			requestBody: "1234566",
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)).ThenReturn(&entity.Order{
					ID:     1234566,
					UserID: 123,
					Status: entity.OrderStatusNew,
				}, nil)

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)
			},
			status: http.StatusAccepted,
			messageResponse: &responses.Message{
				Message: "order has been successfully registered for processing",
			},
		},
		{
			name:        "order already registered by current user",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "text/plain",
			requestBody: "1234566",
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)).ThenReturn(&entity.Order{
					ID:     1234566,
					UserID: 123,
					Status: entity.OrderStatusNew,
				}, managers.ErrOrderAlreadyRegisteredByCurrentUser)

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)
			},
			status: http.StatusOK,
			messageResponse: &responses.Message{
				Message: "order already registered by current user",
			},
		},
		{
			name:        "order already registered by another user",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "text/plain",
			requestBody: "1234566",
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)).ThenReturn(&entity.Order{
					ID:     1234566,
					UserID: 123,
					Status: entity.OrderStatusNew,
				}, managers.ErrOrderAlreadyRegisteredByAnotherUser)

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)
			},
			status: http.StatusConflict,
			errResponse: &responses.APIError{
				Code:    http.StatusConflict,
				Message: "order already registered by another user",
			},
		},
		{
			name:        "cant register order",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "text/plain",
			requestBody: "1234566",
			manager: func() orderManager {
				manager := Mock[orderManager]()
				WhenDouble(manager.Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)).ThenReturn(&entity.Order{
					ID:     1234566,
					UserID: 123,
					Status: entity.OrderStatusNew,
				}, errors.New("some error"))

				return manager
			},
			verify: func(manager orderManager) {
				Verify(manager, Once()).Register(
					AnyContext(),
					Exact(uint64(1234566)),
					Exact(uint32(123)),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t register order",
			},
		},
		{
			name:        "invalid content-type",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "application/json",
			requestBody: "1234566",
			manager: func() orderManager {
				return Mock[orderManager]()
			},
			verify: func(manager orderManager) {
				Verify(manager, Never()).Register(
					AnyContext(),
					Any[uint64](),
					Any[uint32](),
				)
			},
			status: http.StatusBadRequest,
			errResponse: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "invalid Content-Type",
			},
		},
		{
			name:        "invalid order id",
			ctx:         userContext.WithUserID(context.Background(), 123),
			contentType: "text/plain",
			requestBody: "123456",
			manager: func() orderManager {
				return Mock[orderManager]()
			},
			verify: func(manager orderManager) {
				Verify(manager, Never()).Register(
					AnyContext(),
					Any[uint64](),
					Any[uint32](),
				)
			},
			status: http.StatusUnprocessableEntity,
			errResponse: &responses.APIError{
				Code:    http.StatusUnprocessableEntity,
				Message: "invalid order id",
			},
		},
		{
			name:        "cant get credentials",
			ctx:         context.Background(),
			contentType: "text/plain",
			manager: func() orderManager {
				return Mock[orderManager]()
			},
			verify: func(manager orderManager) {
				Verify(manager, Never()).Register(
					AnyContext(),
					Any[uint64](),
					Any[uint32](),
				)
			},
			status: http.StatusInternalServerError,
			errResponse: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "can`t get request credentials",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			manager := tt.manager()
			container := NewContainer(manager, queue.New[uint64](1))
			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", bytes.NewBuffer([]byte(tt.requestBody))).WithContext(tt.ctx)
			request.Header.Set("Content-Type", tt.contentType)

			container.Register(recorder, request)

			require.Equal(t, tt.status, recorder.Code)

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
