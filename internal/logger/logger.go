package logger

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
)

var Logger = zap.NewNop()
var once sync.Once

func Init(name, level string) {
	once.Do(func() {
		atomicLevel, err := zap.ParseAtomicLevel(level)
		if err != nil {
			panic(err)
		}

		config := zap.NewProductionConfig()
		config.Level = atomicLevel
		logger, err := config.Build()
		if err != nil {
			panic(err)
		}

		Logger = logger.Named(name)
	})
}

func RecoverAndPanic() {
	if recovered := recover(); recovered != nil {
		Logger.Panic(fmt.Sprintf("%v", recovered))
	}
}
