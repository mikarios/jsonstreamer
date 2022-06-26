package config

import (
	"context"
	"sync"

	"github.com/kelseyhightower/envconfig"

	"github.com/mikarios/golib/logger"
)

var (
	once     sync.Once
	instance *Config
)

func GetInstance() *Config {
	if instance == nil {
		instance = Init("")
	}

	return instance
}

func Init(prefix string) *Config {
	once.Do(func() {
		instance = &Config{}
		if err := envconfig.Process(prefix, instance); err != nil {
			logger.Panic(context.Background(), err, "could not initialise config")
		}
	})

	return instance
}
