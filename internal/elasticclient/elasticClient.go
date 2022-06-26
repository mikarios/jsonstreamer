package elasticclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/config"
)

var (
	once           sync.Once
	instance       *elastic.Client
	errInstanceNil = errors.New("instance is nil")
)

func GetInstance() *elastic.Client {
	if instance == nil {
		Init()
	}

	return instance
}

func Init() *elastic.Client {
	ctx := context.Background()
	cfg := config.GetInstance()

	once.Do(func() {
		logger.Debug(ctx, "Connecting to ElasticSearch")

		opts := []elastic.ClientOptionFunc{
			elastic.SetURL(cfg.Elastic.URLList...),
			elastic.SetSniff(false),
			elastic.SetHealthcheckInterval(5 * time.Minute),
			elastic.SetGzip(true),
			elastic.SetErrorLog(logger.NewLogger("Elastic", logger.LogLevels.ERROR)),
		}

		var err error

		// Since healthcheck is enabled by default NewClient also checks connection to elastic
		instance, err = elastic.NewClient(opts...)
		if err != nil {
			logger.Panic(ctx, err, "Cannot get elastic info")
		}

		if instance == nil {
			logger.Panic(ctx, errInstanceNil)
		}

		logger.Debug(ctx, "Connected to elastic successfully")
	})

	return instance
}
