package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetAccrual(t *testing.T) {
	accrual := 1.23
	someErr := errors.New("some error")
	tests := []struct {
		name      string
		transport roundTripFunction
		want      *responses.Accrual
		wantErr   error
	}{
		{
			name: "valid",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Accrual{
					Status:  responses.AccrualStatusProcessed,
					OrderID: 1,
					Accrual: &accrual,
				}), nil
			}),
			want: &responses.Accrual{
				Status:  responses.AccrualStatusProcessed,
				OrderID: 1,
				Accrual: &accrual,
			},
		},
		{
			name: "valid no accrual",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return createResponse(t, http.StatusOK, responses.Accrual{
					Status:  responses.AccrualStatusProcessing,
					OrderID: 2,
				}), nil
			}),
			want: &responses.Accrual{
				Status:  responses.AccrualStatusProcessing,
				OrderID: 2,
			},
		},
		{
			name: "no content",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNoContent,
				}, nil
			}),
			wantErr: ErrOrderNotFound,
		},
		{
			name: "internal server error",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil
			}),
			wantErr: ErrInternalServerError,
		},
		{
			name: "too many requests seconds",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header: http.Header{
						"Retry-After": []string{"5"},
					},
				}, nil
			}),
			wantErr: newErrTooManyRequests(retryafter.Parse("5", defaultRetryAfter)),
		},
		{
			name: "too many requests date",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header: http.Header{
						"Retry-After": []string{time.Now().Add(5 * time.Second).Format(time.RFC1123)},
					},
				}, nil
			}),
			wantErr: newErrTooManyRequests(retryafter.Parse(time.Now().Add(5*time.Second).Format(time.RFC1123), defaultRetryAfter)),
		},
		{
			name: "too many requests empty",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header:     map[string][]string{},
				}, nil
			}),
			wantErr: newErrTooManyRequests(retryafter.Parse("", defaultRetryAfter)),
		},
		{
			name: "unexpected status code",
			transport: roundTripFunction(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTeapot,
				}, nil
			}),
			wantErr: newErrUnexpectedStatus(http.StatusTeapot),
		},
		{
			name: "transport error",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, someErr
			},
			wantErr: someErr,
		},
	}
	for orderID, tt := range tests {
		orderID++
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := newTestClient(t, func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, fmt.Sprintf("/api/orders/%d", orderID), req.URL.Path)

				return tt.transport(req)
			})
			response, err := client.GetAccrual(ctx, uint64(orderID))
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr)
				assert.ObjectsAreEqual(tt.want, response)
			} else {
				require.NoError(t, err)
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
