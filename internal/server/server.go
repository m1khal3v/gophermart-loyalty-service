package server

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Start(ctx context.Context, server *http.Server) error {
	suspendCtx, suspendCancel := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer suspendCancel()

	errCtx, errCancel := context.WithCancelCause(ctx)
	defer errCancel(nil)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCancel(err)
		}
	}()

	select {
	case <-errCtx.Done():
		return context.Cause(errCtx)
	case <-suspendCtx.Done():
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		logger.Logger.Info("Received suspend signal. Trying to shutdown gracefully...")

		if err := server.Shutdown(timeoutCtx); err != nil {
			logger.Logger.Error("Failed to shutdown server", zap.Error(err))
		} else {
			logger.Logger.Info("Server was shutdown successfully")
		}

		return nil
	}
}
