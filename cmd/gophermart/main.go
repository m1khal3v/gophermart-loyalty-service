package main

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/app"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/config"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	config := config.ParseConfig()
	logger.Init("server", config.LogLevel)
	defer logger.Logger.Sync()
	defer logger.RecoverAndPanic()
	logger.Logger.Info(
		"Starting",
		zap.String("app_env", config.AppEnv),
		zap.String("run_address", config.RunAddress),
		zap.String("accrual_system_address", config.AccrualSystemAddress),
		zap.String("database_uri", config.DatabaseURI),
		zap.Uint64("retriever_concurrency", config.RetrieverConcurrency),
		zap.Uint64("updater_concurrency", config.UpdaterConcurrency),
		zap.String("log_level", config.LogLevel),
	)

	if app, err := app.New(config); err == nil {
		app.Run()
	} else {
		logger.Logger.Fatal("Can't start application", zap.Error(err))
	}
}
