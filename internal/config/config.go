package config

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Config struct {
	AppEnv               string `env:"APP_ENV"`
	AppSecret            string `env:"APP_SECRET"`
	RunAddress           string `env:"RUN_ADDRESS"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	RetrieverConcurrency uint64 `env:"RETRIEVER_CONCURRENCY"`
	UpdaterConcurrency   uint64 `env:"UPDATER_CONCURRENCY"`
	LogLevel             string `env:"LOG_LEVEL"`
}

func ParseConfig() *Config {
	config := &Config{}
	flag.StringVarP(&config.AppEnv, "env", "e", "dev", "app environment")
	flag.StringVarP(&config.AppSecret, "app-secret", "s", "aPp$eCr3t", "app secret for jwt")
	flag.StringVarP(&config.RunAddress, "address", "a", ":8080", "address of gophermart-loyalty-service server")
	flag.StringVarP(&config.AccrualSystemAddress, "accrual-system-address", "r", "localhost:8081", "address of gophermart-accrual-service server")
	flag.StringVarP(&config.DatabaseURI, "database-uri", "d", "", "database uri")
	flag.Uint64VarP(&config.RetrieverConcurrency, "retriever-concurrency", "c", 10, "retriever concurrency")
	flag.Uint64VarP(&config.UpdaterConcurrency, "updater-concurrency", "u", 10, "updater concurrency")
	flag.StringVarP(&config.LogLevel, "log-level", "l", "info", "log level")
	flag.Parse()
	if err := env.Parse(config); err != nil {
		panic(err)
	}

	return config
}
