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
	retrieverProcessor "github.com/m1khal3v/gophermart-loyalty-service/internal/processor/retriever"
	routerProcessor "github.com/m1khal3v/gophermart-loyalty-service/internal/processor/router"
	invalidProcessor "github.com/m1khal3v/gophermart-loyalty-service/internal/processor/status/invalid"
	processedProcessor "github.com/m1khal3v/gophermart-loyalty-service/internal/processor/status/processed"
	processingProcessor "github.com/m1khal3v/gophermart-loyalty-service/internal/processor/status/processing"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/router"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/server"
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
	server              *http.Server
	retrieverProcessor  *retrieverProcessor.Processor
	routerProcessor     *routerProcessor.Processor
	processingProcessor *processingProcessor.Processor
	invalidProcessor    *invalidProcessor.Processor
	processedProcessor  *processedProcessor.Processor
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
	orderQueue := queue.New[uint64](10000)
	for unprocessedID := range unprocessedIDs {
		orderQueue.Push(unprocessedID)
	}
	routerQueue := queue.New[*responses.Accrual](10000)
	invalidQueue := queue.New[*responses.Accrual](10000)
	processingQueue := queue.New[*responses.Accrual](10000)
	processedQueue := queue.New[*responses.Accrual](10000)

	// Router
	authRoutes := auth.NewContainer(userManager)
	orderRoutes := order.NewContainer(orderManager, orderQueue)
	balanceRoutes := balance.NewContainer(userManager, userWithdrawalManager)
	withdrawalRoutes := withdrawal.NewContainer(withdrawalManager)
	router := router.New(config.AppEnv == "prod", authRoutes, orderRoutes, balanceRoutes, withdrawalRoutes, jwt)

	// Accrual
	client, err := client.New(&client.Config{
		Address: config.AccrualSystemAddress,
	})
	if err != nil {
		return nil, err
	}

	return &app{
		server: server.New(config.RunAddress, router),
		retrieverProcessor: retrieverProcessor.NewProcessor(client, orderQueue, routerQueue, &retrieverProcessor.Config{
			Concurrency: config.RetrieverConcurrency,
		}),
		routerProcessor: routerProcessor.NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &routerProcessor.Config{
			Concurrency: config.RouterConcurrency,
		}),
		processingProcessor: processingProcessor.NewProcessor(orderQueue, processingQueue, orderManager, config.ProcessingConcurrency, config.UpdateBatchSize),
		invalidProcessor:    invalidProcessor.NewProcessor(invalidQueue, orderManager, config.InvalidConcurrency, config.UpdateBatchSize),
		processedProcessor:  processedProcessor.NewProcessor(processedQueue, userOrderManager, config.ProcessedConcurrency, config.UpdateBatchSize),
	}, nil
}

func (app *app) Run() {
	ctx := context.Background()
	suspendCtx, suspendCancel := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer suspendCancel()

	errCtx, errCancel := context.WithCancelCause(ctx)
	defer errCancel(nil)

	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		if err := app.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCancel(fmt.Errorf("server error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.retrieverProcessor.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("retriever processor error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.routerProcessor.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("router processor error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.processingProcessor.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("processing processor error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.invalidProcessor.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("invalid processor error: %w", err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := app.processedProcessor.Process(suspendCtx); !errors.Is(err, context.Canceled) {
			errCancel(fmt.Errorf("processed processor error: %w", err))
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
