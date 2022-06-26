package portdomainsvc

import (
	"context"
	"sync"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
	"github.com/mikarios/jsonstreamer/internal/services/databasesvc"
	"github.com/mikarios/jsonstreamer/internal/services/jsonstreamersvc"
	"github.com/mikarios/jsonstreamer/internal/services/portcollectorsvc"
)

type portDomainService struct {
	jsonStreamingSvc     *jsonstreamersvc.JSONStreamer[*portmodel.PortData]
	portCollectorService *portcollectorsvc.PortCollectorService
	dbService            *databasesvc.DBService
	notifyOnFinish       <-chan interface{}
}

var doneCh = make(chan interface{})

func New(
	jsonStreamingSvc *jsonstreamersvc.JSONStreamer[*portmodel.PortData],
	portCollectorService *portcollectorsvc.PortCollectorService,
	dbService *databasesvc.DBService,
	notifyOnFinish <-chan interface{},
) *portDomainService {
	return &portDomainService{
		jsonStreamingSvc:     jsonStreamingSvc,
		portCollectorService: portCollectorService,
		dbService:            dbService,
		notifyOnFinish:       notifyOnFinish,
	}
}

func (pd *portDomainService) Start(ctx context.Context) (doneIndicatorChannel chan interface{}) {
	go pd.jsonStreamingSvc.Start()
	go pd.portCollectorService.Start()
	go pd.sendItemsToPortCollector(ctx)
	go pd.waitForProcessingToFinish()

	return doneCh
}

// If we catch this error there is a problem with interrupt signal not exiting. Apparently I don't have time to find
// the issue
// defer func() {
// 	if r := recover(); r != nil {
// 		err, _ := r.(error)
// 		logger.Error(context.Background(), err, "could not send port to channel. Is system shutting down?")
// 		return
// 	}
// }()

func (pd *portDomainService) sendItemsToPortCollector(ctx context.Context) {
	stream := pd.jsonStreamingSvc.Watch()
	portChan := pd.portCollectorService.AddPorts()

	for item := range stream {
		if item.Err != nil {
			logger.Error(ctx, item.Err, "problem getting item from stream")
			continue
		}

		portChan <- &portmodel.Port{Key: item.Key, Data: item.Data}
	}

	close(pd.portCollectorService.AddPorts())
}

func (pd *portDomainService) waitForProcessingToFinish() {
	<-pd.notifyOnFinish
	<-pd.portCollectorService.WaitToFinish()
	<-pd.dbService.GracefulShutdown()

	doneCh <- struct{}{}
}

func (pd *portDomainService) GracefulShutdown() {
	wg := &sync.WaitGroup{}

	wg.Add(3)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		<-pd.jsonStreamingSvc.GracefulShutdown()
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		<-pd.portCollectorService.GracefulShutdown()
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		<-pd.dbService.GracefulShutdown()
	}(wg)

	wg.Wait()

	doneCh <- struct{}{}
}
