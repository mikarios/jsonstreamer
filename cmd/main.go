package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/elasticclient"
	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
	"github.com/mikarios/jsonstreamer/internal/services/databasesvc"
	"github.com/mikarios/jsonstreamer/internal/services/jsonstreamersvc"
	"github.com/mikarios/jsonstreamer/internal/services/portcollectorsvc"
	"github.com/mikarios/jsonstreamer/internal/services/portdomainsvc"
)

func main() {
	bgCTX := context.Background()
	cfg := config.Init("")

	if err := setupLogger(cfg.LOG.Level, cfg.LOG.Format, cfg.LOG.Trace); err != nil {
		logger.Panic(bgCTX, err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	elasticclient.Init()

	jsonStreamerSvc, err := jsonstreamersvc.New[*portmodel.PortData](cfg.PortsFileLocation, 0)
	if err != nil {
		logger.Panic(bgCTX, err, "could not initialise streamer service")
	}

	dbSVC := databasesvc.New()
	portCollectorSVC := portcollectorsvc.New(dbSVC)
	portDomainService := portdomainsvc.New(jsonStreamerSvc, portCollectorSVC)

	go portDomainService.Start(bgCTX)

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
