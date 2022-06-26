package portcollectorsvc

import (
	"context"
	"runtime"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
)

var (
	workerFinishedChan = make(chan interface{})
	shutdownChan       = make(chan interface{})
	noOfWorkers        int
)

type idb interface {
	Store(ctx context.Context, key string, data interface{}) error
}

type PortCollectorService struct {
	portChan chan *portmodel.Port
	db       idb
}

func New(dbService idb) *PortCollectorService {
	portCh := make(chan *portmodel.Port)
	return &PortCollectorService{portChan: portCh, db: dbService}
}

func (pc *PortCollectorService) AddPorts() chan<- *portmodel.Port {
	return pc.portChan
}

func (pc *PortCollectorService) Start() {
	maxProcesses := runtime.GOMAXPROCS(0)
	noOfWorkers = 2 * maxProcesses

	if cfg := config.GetInstance(); cfg.PortCollectorWorkers > 0 {
		noOfWorkers = cfg.PortCollectorWorkers
	}

	for i := 0; i < noOfWorkers; i++ {
		go pc.spawnWorker()
	}
}

func (pc *PortCollectorService) spawnWorker() {
	defer func() {
		workerFinishedChan <- struct{}{}
	}()

	for port := range pc.portChan {
		if err := pc.db.Store(context.Background(), port.Key, port.Data); err != nil {
			logger.Error(context.Background(), err, "could not store to db", port.Key)
		}
	}
}

func (pc *PortCollectorService) GracefulShutdown() <-chan interface{} {
	close(pc.portChan)

	go waitWorkersToFinish()

	return shutdownChan
}

func (pc *PortCollectorService) WaitToFinish() <-chan interface{} {
	go waitWorkersToFinish()

	return shutdownChan
}

func waitWorkersToFinish() {
	for i := 0; i < noOfWorkers; i++ {
		<-workerFinishedChan
	}

	shutdownChan <- struct{}{}
}
