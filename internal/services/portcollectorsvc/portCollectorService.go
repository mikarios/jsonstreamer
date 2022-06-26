package portcollectorsvc

import (
	"context"
	"errors"
	"runtime"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
)

var (
	workerFinishedChan = make(chan interface{})
	shutdownChan       = make(chan interface{})
	noOfWorkers        int
	ErrEmptyDB         = errors.New("db is nil")
)

// In principle, I wouldn't use a DB here IF there was some processing to be done to the ports. I would implement a
// similar way of communication as the one between jsonStreamer and portCollector.
// Since there's no processing done, and I don't think I will have enough time for this, I will simply inject a DB here.
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

func (pc *PortCollectorService) NewPort(port *portmodel.Port) {
	defer func() {
		if r := recover(); r != nil {
			err, _ := r.(error)
			logger.Error(context.Background(), err, "could not send port to channel. Is system shutting down?")
		}
	}()

	pc.portChan <- port
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
		if pc.db == nil {
			logger.Error(context.Background(), ErrEmptyDB, "could not store", port.Key, *port.Data)
			continue
		}

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

func waitWorkersToFinish() {
	for i := 0; i < noOfWorkers; i++ {
		<-workerFinishedChan
	}

	shutdownChan <- struct{}{}
}
