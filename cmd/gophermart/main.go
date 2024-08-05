package main

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/config"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/container"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/server"
	"go.uber.org/zap"
)

func main() {
	config := config.ParseConfig()
	logger.Init("server", config.LogLevel)
	defer logger.Logger.Sync()
	defer logger.RecoverAndPanic()
	logger.Logger.Info(
		"Starting",
		zap.String("run_address", config.RunAddress),
		zap.String("accrual_system_address", config.AccrualSystemAddress),
		zap.String("database_uri", config.DatabaseURI),
		zap.String("log_level", config.LogLevel),
	)

	container, err := container.New(config)
	if err != nil {
		logger.Logger.Panic(err.Error())
	}

	processorCtx, processorCancel := context.WithCancel(context.Background())
	defer processorCancel()
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	go func() {
		defer processorCancel()
		if err := server.Start(serverCtx, container.Server); err != nil {
			logger.Logger.Panic(err.Error())
		}
	}()
	go func() {
		defer serverCancel()
		if err := container.Retriever.Process(processorCtx, config.RetrieverConcurrency); err != nil {
			logger.Logger.Error(err.Error())
		}
	}()
	go func() {
		defer serverCancel()
		if err := container.Updater.Process(processorCtx, config.UpdaterConcurrency); err != nil {
			logger.Logger.Error(err.Error())
		}
	}()

	<-processorCtx.Done()
}
