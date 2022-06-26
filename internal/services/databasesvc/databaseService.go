package databasesvc

import (
	"context"

	"github.com/olivere/elastic/v7"

	"github.com/mikarios/jsonstreamer/internal/config"
	"github.com/mikarios/jsonstreamer/internal/elasticclient"
)

var shutDownChannel = make(chan interface{})

type dbService struct {
	esClient *elastic.Client
}

func New() *dbService {
	return &dbService{esClient: elasticclient.GetInstance()}
}

func (db *dbService) Store(ctx context.Context, key string, data interface{}) error {
	cfg := config.GetInstance()
	_, err := db.
		esClient.
		Index().
		Index(cfg.Elastic.Indices.Ports.Index).
		Id(key).
		BodyJson(data).
		Do(ctx)

	return err
}

func (db *dbService) GracefulShutdown() <-chan interface{} {
	go func() {
		shutDownChannel <- struct{}{}
	}()

	return shutDownChannel
}
