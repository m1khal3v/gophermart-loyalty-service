package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogRequest(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		path   string
		method string
		status int
	}{
		{
			name:   "ok",
			body:   "Hello, World!",
			path:   "/ok",
			method: http.MethodGet,
			status: http.StatusOK,
		},
		{
			name:   "not found",
			body:   "Not found!",
			path:   "/404",
			method: http.MethodPost,
			status: http.StatusNotFound,
		},
		{
			name:   "bad request",
			body:   "Bad request!",
			path:   "/400",
			method: http.MethodPut,
			status: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			core, logs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core)
			router.Use(ZapLogRequest(logger, tt.name))
			router.MethodFunc(tt.method, tt.path, func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(tt.status)
				writer.Write([]byte(tt.body))
			})
			httpServer := httptest.NewServer(router)
			defer httpServer.Close()

			request, err := http.NewRequest(tt.method, httpServer.URL+tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()

			allLogs := logs.All()
			require.Len(t, allLogs, 1)
			log := allLogs[0]

			assert.Equal(t, "Request processed", log.Message)
			assert.Equal(t, zapcore.InfoLevel, log.Level)
			assert.NotEmpty(t, log.Time)
			assert.Equal(t, tt.name, log.LoggerName)
			assert.False(t, log.Caller.Defined)

			fields := log.Context
			assert.Len(t, fields, 5)
			assert.Equal(t, tt.method, fieldByKey(t, fields, "method").String)
			assert.Equal(t, tt.path, fieldByKey(t, fields, "url").String)
			assert.Equal(t, int64(tt.status), fieldByKey(t, fields, "status").Integer)
			assert.Equal(t, int64(len(tt.body)), fieldByKey(t, fields, "size").Integer)
			assert.NotZero(t, fieldByKey(t, fields, "duration").Integer)
		})
	}
}

func TestZapLogPanic(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "get",
			path:   "/get",
			method: http.MethodGet,
		},
		{
			name:   "post",
			path:   "/post",
			method: http.MethodPost,
		},
		{
			name:   "put",
			path:   "/put",
			method: http.MethodPut,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			core, logs := observer.New(zapcore.ErrorLevel)
			logger := zap.New(core)
			router.Use(ZapLogPanic(logger, tt.name))
			router.MethodFunc(tt.method, tt.path, func(writer http.ResponseWriter, request *http.Request) {
				panic(tt.method + " panic")
			})
			httpServer := httptest.NewServer(router)
			defer httpServer.Close()

			request, err := http.NewRequest(tt.method, httpServer.URL+tt.path, nil)
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(request)
			if response != nil {
				response.Body.Close() // go vet fix
			}
			require.Nil(t, response)
			require.Error(t, err)

			allLogs := logs.All()
			require.Len(t, allLogs, 1)
			log := allLogs[0]

			assert.Equal(t, tt.method+" panic", log.Message)
			assert.Equal(t, zapcore.ErrorLevel, log.Level)
			assert.NotEmpty(t, log.Time)
			assert.Equal(t, tt.name, log.LoggerName)
			assert.False(t, log.Caller.Defined)

			fields := log.Context
			assert.Len(t, fields, 2)
			assert.Equal(t, tt.method, fieldByKey(t, fields, "method").String)
			assert.Equal(t, tt.path, fieldByKey(t, fields, "url").String)
		})
	}
}

func fieldByKey(t *testing.T, fields []zapcore.Field, key string) *zapcore.Field {
	t.Helper()
	for _, field := range fields {
		if field.Key == key {
			return &field
		}
	}

	t.Fatalf("field %s not found", key)
	return nil
}
