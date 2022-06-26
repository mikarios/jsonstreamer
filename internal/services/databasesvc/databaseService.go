package databasesvc

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/olivere/elastic/v7"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/elasticclient"
)

var (
	shutDownChannel = make(chan interface{})
	writingInDB     = make(map[string]interface{})
	writingInDBLock = &sync.Mutex{}
)

type DBService struct {
	esClient *elastic.Client
}

func New() *DBService {
	return &DBService{esClient: elasticclient.GetInstance()}
}

func (db *DBService) Store(ctx context.Context, key string, data interface{}) error {
	cfg := config.GetInstance()

	myID := uuid.New().String()

	writingInDBLock.Lock()
	writingInDB[myID] = struct{}{}
	writingInDBLock.Unlock()

	_, err := db.
		esClient.
		Index().
		Index(cfg.Elastic.Indices.Ports.Index).
		Id(key).
		BodyJson(data).
		Do(ctx)

	writingInDBLock.Lock()
	delete(writingInDB, myID)
	writingInDBLock.Unlock()

	return err
}

func (db *DBService) GracefulShutdown() <-chan interface{} {
	go func() {
		t := time.NewTicker(10 * time.Millisecond)

		// This is pointless with the current solution as this is synchronous
		for range t.C {
			writingInDBLock.Lock()

			if len(writingInDB) == 0 {
				writingInDB = nil
				shutDownChannel <- struct{}{}

				return
			} else {
				logger.Trace(context.Background(), "waiting for ", len(writingInDB), " db operations to finish")
			}

			writingInDBLock.Unlock()
		}
	}()

	return shutDownChannel
}
