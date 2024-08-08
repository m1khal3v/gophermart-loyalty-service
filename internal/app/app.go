package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/processor"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/task"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/config"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/auth"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/balance"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/order"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller/withdrawal"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/router"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	server    *http.Server
	retriever *processor.Retriever
	updater   *processor.Updater
}

func New(config *config.Config) (*App, error) {
	gorm, err := gorm.Open(postgres.Open(config.DatabaseURI), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	client, err := client.New(&client.Config{
		Address: config.AccrualSystemAddress,
	})
	if err != nil {
		return nil, err
	}

	jwt := jwt.New(config.AppSecret)
	userRepository := repository.NewUserRepository(gorm)
	userManager := manager.NewUserManager(userRepository, jwt)
	withdrawalRepository := repository.NewWithdrawalRepository(gorm)
	withdrawalManager := manager.NewWithdrawalManager(withdrawalRepository)
	orderRepository := repository.NewOrderRepository(gorm)
	orderManager := manager.NewOrderManager(orderRepository)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	unprocessedIDs, err := orderRepository.FindUnprocessedIDs(ctx)
	if err != nil {
		return nil, err
	}
	taskManager := task.NewTaskManager(unprocessedIDs)
	userWithdrawalRepository := repository.NewUserWithdrawalRepository(gorm)
	userWithdrawalManager := manager.NewUserWithdrawalManager(userWithdrawalRepository)
	userOrderRepository := repository.NewUserOrderRepository(gorm)
	userOrderManager := manager.NewUserOrderManager(userOrderRepository)
	authRoutes := auth.NewContainer(userManager)
	orderRoutes := order.NewContainer(orderManager, taskManager)
	balanceRoutes := balance.NewContainer(userManager, withdrawalManager, userWithdrawalManager)
	withdrawalRoutes := withdrawal.NewContainer(withdrawalManager)
	router := router.New(config.AppEnv, authRoutes, orderRoutes, balanceRoutes, withdrawalRoutes, jwt)

	return &App{
		server: &http.Server{
			Addr:    config.RunAddress,
			Handler: router,
		},
		retriever: processor.NewRetriever(client, taskManager, config.RetrieverConcurrency),
		updater:   processor.NewUpdater(taskManager, orderManager, userOrderManager, config.UpdaterConcurrency),
	}, nil
}

func (app *App) Run() {
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
		app.shutdown()
	case <-suspendCtx.Done():
		logger.Logger.Info("Received suspend signal.")
		app.shutdown()
	}

	logger.Logger.Info("Waiting for all goroutines to finish...")
	wg.Wait()
}

func (app *App) shutdown() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	logger.Logger.Info("Trying to shutdown server gracefully...")
	if err := app.server.Shutdown(timeoutCtx); err != nil {
		logger.Logger.Error("Failed to shutdown server", zap.Error(err))
	} else {
		logger.Logger.Info("Server was shutdown successfully")
	}
}
