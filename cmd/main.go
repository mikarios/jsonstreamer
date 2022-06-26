package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
)

func main() {
	bgCTX := context.Background()
	cfg := config.Init("")

	if err := setupLogger(cfg.LOG.Level, cfg.LOG.Format, cfg.LOG.Trace); err != nil {
		logger.Panic(bgCTX, err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	event := <-quit
	logger.Warning(bgCTX, fmt.Sprintf("RECEIVED SIGNAL: %v exiting", event))
	gracefulShutdown()
}

func setupLogger(level, formatter string, trace bool) error {
	if err := logger.SetFormatter(formatter); err != nil {
		return err
	}

	logger.SetLogTrace(trace)

	return logger.SetLogLevel(level)
}

func gracefulShutdown() {
	// todo: implement this
}