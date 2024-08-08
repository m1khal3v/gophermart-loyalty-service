package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/config"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/auth"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/balance"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/order"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/withdrawal"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	processor2 "github.com/m1khal3v/gophermart-loyalty-service/internal/processor"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/router"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type app struct {
	server    *http.Server
	retriever *processor2.Retriever
	updater   *processor2.Updater
}

// New function acts as the simplest configuration-based dependency injector
func New(config *config.Config) (*app, error) {
	// JWT
	jwt := jwt.New(config.AppSecret)

	// DB
	gorm, err := gorm.Open(postgres.Open(config.DatabaseURI), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	userRepository := repository.NewUserRepository(gorm)
	withdrawalRepository := repository.NewWithdrawalRepository(gorm)
	orderRepository := repository.NewOrderRepository(gorm)
	userWithdrawalRepository := repository.NewUserWithdrawalRepository(gorm)
	userOrderRepository := repository.NewUserOrderRepository(gorm)

	// Managers
	userManager := manager.NewUserManager(userRepository, jwt)
	withdrawalManager := manager.NewWithdrawalManager(withdrawalRepository)
	orderManager := manager.NewOrderManager(orderRepository)
	userWithdrawalManager := manager.NewUserWithdrawalManager(userWithdrawalRepository)
	userOrderManager := manager.NewUserOrderManager(userOrderRepository)

	// Queue
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	unprocessedIDs, err := orderRepository.FindUnprocessedIDs(ctx)
	if err != nil {
		return nil, err
	}
	unprocessedQueue := queue.New[uint64](10000)
	for unprocessedID := range unprocessedIDs {
		unprocessedQueue.Push(unprocessedID)
	}
	processedQueue := queue.New[*responses.Accrual](10000)

	// Router
	authRoutes := auth.NewContainer(userManager)
	orderRoutes := order.NewContainer(orderManager, unprocessedQueue)
	balanceRoutes := balance.NewContainer(userManager, withdrawalManager, userWithdrawalManager)
	withdrawalRoutes := withdrawal.NewContainer(withdrawalManager)
	router := router.New(config.AppEnv, authRoutes, orderRoutes, balanceRoutes, withdrawalRoutes, jwt)

	// Accrual
	client, err := client.New(&client.Config{
		Address: config.AccrualSystemAddress,
	})
	if err != nil {
		return nil, err
	}

	return &app{
		server: &http.Server{
			Addr:    config.RunAddress,
			Handler: router,
		},
		retriever: processor2.NewRetriever(client, unprocessedQueue, processedQueue, config.RetrieverConcurrency),
		updater:   processor2.NewUpdater(unprocessedQueue, processedQueue, orderManager, userOrderManager, config.UpdaterConcurrency),
	}, nil
}

func (app *app) Run() {
	ctx := context.Background()
	suspendCtx, suspendCancel := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer suspendCancel()

	errCtx, errCancel := context.WithCancelCause(ctx)
	defer errCancel(nil)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := app.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCancel(fmt.Errorf("server error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.retriever.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("retriever error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.updater.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("updater error: %w", err))
		}
	}()

	select {
	case <-errCtx.Done():
		suspendCancel()
		logger.Logger.Error("An error occurred while the application was running", zap.Error(context.Cause(errCtx)))
	case <-suspendCtx.Done():
		logger.Logger.Info("Received suspend signal.")
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	logger.Logger.Info("Trying to shutdown server gracefully...")
	if err := app.server.Shutdown(timeoutCtx); err != nil {
		logger.Logger.Error("Failed to shutdown server", zap.Error(err))
	} else {
		logger.Logger.Info("Server was shutdown successfully")
	}

	logger.Logger.Info("Waiting for all goroutines to finish...")
	wg.Wait()
}
