package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name       string
		transport  roundTripFunction
		request    *requests.Register
		want       *responses.Auth
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			request: &requests.Register{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Auth{
					AccessToken: "t0k3n",
				}), nil
			}),
			want: &responses.Auth{
				AccessToken: "t0k3n",
			},
		},
		{
			name: "too many requests",
			request: &requests.Register{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				response := createResponse(t, http.StatusTooManyRequests, responses.APIError{
					Code:    http.StatusTooManyRequests,
					Message: "too many requests",
				})
				response.Header.Set("Retry-After", "5")

				return response, nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTooManyRequests,
				Message: "too many requests",
			},
			wantErr: newErrTooManyRequests(retryafter.Parse("5", defaultRetryAfter)),
		},
		{
			name: "bad request",
			request: &requests.Register{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, responses.APIError{
					Code:    http.StatusBadRequest,
					Message: "bad request",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "internal server error",
			request: &requests.Register{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "unexpected status",
			request: &requests.Register{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "/api/user/register", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Register(ctx, tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name       string
		transport  roundTripFunction
		request    *requests.Login
		want       *responses.Auth
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Auth{
					AccessToken: "t0k3n",
				}), nil
			}),
			want: &responses.Auth{
				AccessToken: "t0k3n",
			},
		},
		{
			name: "too many requests",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				response := createResponse(t, http.StatusTooManyRequests, responses.APIError{
					Code:    http.StatusTooManyRequests,
					Message: "too many requests",
				})
				response.Header.Set("Retry-After", "5")

				return response, nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTooManyRequests,
				Message: "too many requests",
			},
			wantErr: newErrTooManyRequests(retryafter.Parse("5", defaultRetryAfter)),
		},
		{
			name: "bad request",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, responses.APIError{
					Code:    http.StatusBadRequest,
					Message: "bad request",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "internal server error",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "invalid credentials",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unexpected status",
			request: &requests.Login{
				Login:    "ivan_ivanov",
				Password: "swordfish",
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "/api/user/login", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Login(ctx, tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_AddOrder(t *testing.T) {
	tests := []struct {
		name       string
		transport  roundTripFunction
		request    uint64
		want       *responses.Message
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name:    "valid",
			request: 1234566,
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Message{
					Message: "order successfully added",
				}), nil
			}),
			want: &responses.Message{
				Message: "order successfully added",
			},
		},
		{
			name:    "bad request",
			request: 123456,
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, responses.APIError{
					Code:    http.StatusBadRequest,
					Message: "bad request",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantErr: ErrBadRequest,
		},
		{
			name:    "internal server error",
			request: 1234566,
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name:    "invalid credentials",
			request: 1234566,
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "unexpected status",
			request: 1234566,
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))
				assert.Equal(t, fmt.Sprintf("Bearer t0k3n_%d", i), req.Header.Get("Authorization"))
				assert.Equal(t, "/api/user/orders", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.AddOrder(ctx, fmt.Sprintf("t0k3n_%d", i), tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_Orders(t *testing.T) {
	now := time.Now().UTC()
	accrual := 1.23
	tests := []struct {
		name       string
		transport  roundTripFunction
		want       []responses.Order
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, []responses.Order{
					{
						Number:     1,
						Status:     "STATUS1",
						UploadedAt: now,
					},
					{
						Number:     2,
						Status:     "STATUS2",
						Accrual:    &accrual,
						UploadedAt: now,
					},
				}), nil
			}),
			want: []responses.Order{
				{
					Number:     1,
					Status:     "STATUS1",
					UploadedAt: now,
				},
				{
					Number:     2,
					Status:     "STATUS2",
					Accrual:    &accrual,
					UploadedAt: now,
				},
			},
		},
		{
			name: "no content",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusNoContent, []responses.Order{}), nil
			}),
			want: []responses.Order{},
		},
		{
			name: "internal server error",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "invalid credentials",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unexpected status",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, fmt.Sprintf("Bearer t0k3n_%d", i), req.Header.Get("Authorization"))
				assert.Equal(t, "/api/user/orders", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Orders(ctx, fmt.Sprintf("t0k3n_%d", i))
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.EqualValues(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_Withdrawals(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name       string
		transport  roundTripFunction
		want       []responses.Withdrawal
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, []responses.Withdrawal{
					{
						Order:       1,
						Sum:         1.11,
						ProcessedAt: now,
					},
					{
						Order:       2,
						Sum:         2.22,
						ProcessedAt: now,
					},
				}), nil
			}),
			want: []responses.Withdrawal{
				{
					Order:       1,
					Sum:         1.11,
					ProcessedAt: now,
				},
				{
					Order:       2,
					Sum:         2.22,
					ProcessedAt: now,
				},
			},
		},
		{
			name: "no content",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusNoContent, []responses.Withdrawal{}), nil
			}),
			want: []responses.Withdrawal{},
		},
		{
			name: "internal server error",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "invalid credentials",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unexpected status",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, fmt.Sprintf("Bearer t0k3n_%d", i), req.Header.Get("Authorization"))
				assert.Equal(t, "/api/user/withdrawals", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Withdrawals(ctx, fmt.Sprintf("t0k3n_%d", i))
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.EqualValues(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_Balance(t *testing.T) {
	tests := []struct {
		name       string
		transport  roundTripFunction
		want       *responses.Balance
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Balance{
					Current:   1.23,
					Withdrawn: 3.21,
				}), nil
			}),
			want: &responses.Balance{
				Current:   1.23,
				Withdrawn: 3.21,
			},
		},
		{
			name: "internal server error",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "invalid credentials",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unexpected status",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, fmt.Sprintf("Bearer t0k3n_%d", i), req.Header.Get("Authorization"))
				assert.Equal(t, "/api/user/balance", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Balance(ctx, fmt.Sprintf("t0k3n_%d", i))
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

func TestClient_Withdraw(t *testing.T) {
	tests := []struct {
		name       string
		transport  roundTripFunction
		request    *requests.Withdraw
		want       *responses.Message
		wantAPIErr *responses.APIError
		wantErr    error
	}{
		{
			name: "valid",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Message{
					Message: "funds successfully withdrawn",
				}), nil
			}),
			want: &responses.Message{
				Message: "funds successfully withdrawn",
			},
		},
		{
			name: "bad request",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusBadRequest, responses.APIError{
					Code:    http.StatusBadRequest,
					Message: "bad request",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "insufficient balance",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusPaymentRequired, responses.APIError{
					Code:    http.StatusPaymentRequired,
					Message: "payment required",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusPaymentRequired,
				Message: "payment required",
			},
			wantErr: newErrUnexpectedStatus(http.StatusPaymentRequired),
		},
		{
			name: "internal server error",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusInternalServerError, responses.APIError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErr: ErrInternalServerError,
		},
		{
			name: "invalid credentials",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusUnauthorized, responses.APIError{
					Code:    http.StatusUnauthorized,
					Message: "invalid credentials",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusUnauthorized,
				Message: "invalid credentials",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unexpected status",
			request: &requests.Withdraw{
				Order: 123456,
				Sum:   1.23,
			},
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusTeapot, responses.APIError{
					Code:    http.StatusTeapot,
					Message: "unexpected status",
				}), nil
			}),
			wantAPIErr: &responses.APIError{
				Code:    http.StatusTeapot,
				Message: "unexpected status",
			},
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, fmt.Sprintf("Bearer t0k3n_%d", i), req.Header.Get("Authorization"))
				assert.Equal(t, "/api/user/balance/withdraw", req.URL.Path)

				return tt.transport(req)
			})
			response, apiErr, err := client.Withdraw(ctx, fmt.Sprintf("t0k3n_%d", i), tt.request)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantAPIErr != nil {
				assert.Equal(t, tt.wantAPIErr, apiErr)
			} else {
				require.Nil(t, apiErr)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, response)
			} else {
				require.Nil(t, response)
			}
		})
	}
}

type roundTripFunction func(req *http.Request) (*http.Response, error)

func (function roundTripFunction) RoundTrip(req *http.Request) (*http.Response, error) {
	return function(req)
}

func newTestClient(t *testing.T, function roundTripFunction) *Client {
	t.Helper()
	client, err := New(&Config{
		DisableRetry:             true,
		DisableAddressValidation: true,
		transport:                function,
	})
	require.NoError(t, err)
	require.NotNil(t, client)

	return client
}

func createResponse(t *testing.T, statusCode int, response any) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: statusCode,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: createResponseBody(t, response),
	}
}

func createResponseBody(t *testing.T, response any) io.ReadCloser {
	t.Helper()
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	return io.NopCloser(bytes.NewReader(jsonResponse))
}
