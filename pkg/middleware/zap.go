package middleware

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func ZapLogRequest(logger *zap.Logger, name string) func(next http.Handler) http.Handler {
	logger = logger.Named(name).WithOptions(zap.WithCaller(false))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			wrapper := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			timestamp := time.Now()
			defer func() {
				logger.Info(
					"Request processed",
					zap.String("method", request.Method),
					zap.String("url", request.URL.String()),
					zap.Int("status", wrapper.Status()),
					zap.Int("size", wrapper.BytesWritten()),
					zap.Duration("duration", time.Since(timestamp)),
				)
			}()
			next.ServeHTTP(wrapper, request)
		})
	}
}

func ZapLogPanic(logger *zap.Logger, name string) func(next http.Handler) http.Handler {
	logger = logger.Named(name).WithOptions(zap.WithCaller(false))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				recovered := recover()

				switch recovered {
				case nil:
					return
				case http.ErrAbortHandler:
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(recovered)
				}

				logger.Error(
					fmt.Sprintf("%v", recovered),
					zap.String("method", request.Method),
					zap.String("url", request.URL.String()),
				)

				panic(recovered)
			}()

			next.ServeHTTP(writer, request)
		})
	}
}
