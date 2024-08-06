package container

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/processor"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/task"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/config"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Container struct {
	Server    *http.Server
	Retriever *processor.Retriever
	Updater   *processor.Updater
}

func New(config *config.Config) (*Container, error) {
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
	router := router.New(userManager, orderManager, taskManager, withdrawalManager, userWithdrawalManager, jwt)
	server := &http.Server{
		Addr:    config.RunAddress,
		Handler: router,
	}

	return &Container{
		Server:    server,
		Retriever: processor.NewRetriever(client, taskManager),
		Updater:   processor.NewUpdater(taskManager, orderManager),
	}, nil
}
