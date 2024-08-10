package config

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	AppEnv                string `env:"APP_ENV"`
	AppSecret             string `env:"APP_SECRET"`
	RunAddress            string `env:"RUN_ADDRESS"`
	AccrualSystemAddress  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI           string `env:"DATABASE_URI"`
	RetrieverConcurrency  uint64 `env:"RETRIEVER_CONCURRENCY"`
	RouterConcurrency     uint64 `env:"ROUTER_CONCURRENCY"`
	ProcessingConcurrency uint64 `env:"PROCESSING_CONCURRENCY"`
	InvalidConcurrency    uint64 `env:"INVALID_CONCURRENCY"`
	ProcessedConcurrency  uint64 `env:"PROCESSED_CONCURRENCY"`
	UpdateBatchSize       uint64 `env:"UPDATE_BATCH_SIZE"`
	LogLevel              string `env:"LOG_LEVEL"`
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVarP(&config.AppEnv, "env", "e", "dev", "app environment")
	flag.StringVarP(&config.AppSecret, "app-secret", "s", "aPp$eCr3t", "app secret for jwt")
	flag.StringVarP(&config.RunAddress, "address", "a", ":8080", "address of gophermart-loyalty-service server")
	flag.StringVarP(&config.AccrualSystemAddress, "accrual-system-address", "r", "localhost:8081", "address of gophermart-accrual-service server")
	flag.StringVarP(&config.DatabaseURI, "database-uri", "d", "", "database uri")
	flag.Uint64Var(&config.RetrieverConcurrency, "retriever-concurrency", 10, "retriever concurrency")
	flag.Uint64Var(&config.RouterConcurrency, "router-concurrency", 10, "router concurrency")
	flag.Uint64Var(&config.ProcessingConcurrency, "processing-concurrency", 10, "processing concurrency")
	flag.Uint64Var(&config.InvalidConcurrency, "invalid-concurrency", 10, "invalid concurrency")
	flag.Uint64Var(&config.ProcessedConcurrency, "processed-concurrency", 10, "processed concurrency")
	flag.Uint64Var(&config.UpdateBatchSize, "update-batch-size", 100, "update batch size")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Parse()
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return config
}
