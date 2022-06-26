package portdomainsvc

import (
	"context"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
	"github.com/mikarios/jsonstreamer/internal/services/jsonstreamersvc"
)

type portDomainService struct {
	jsonStreamingSvc *jsonstreamersvc.JSONStreamer[*portmodel.PortData]
}

func New(jsonStreamingSvc *jsonstreamersvc.JSONStreamer[*portmodel.PortData]) *portDomainService {
	return &portDomainService{jsonStreamingSvc: jsonStreamingSvc}
}

func (pd *portDomainService) Start(ctx context.Context) {
	stream := pd.jsonStreamingSvc.Watch()

	go pd.jsonStreamingSvc.Start()

	go func() {
		for item := range stream {
			if item.Err != nil {
				logger.Error(ctx, item.Err, "problem getting item from stream")
				continue
			}

			logger.Warning(ctx, portmodel.Port{
				Key:  item.Key,
				Data: item.Data,
			})
		}
	}()
}
