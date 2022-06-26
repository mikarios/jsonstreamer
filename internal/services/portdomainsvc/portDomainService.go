package portdomainsvc

import (
	"context"

	"github.com/mikarios/golib/logger"

	"github.com/mikarios/jsonstreamer/internal/models/portmodel"
	"github.com/mikarios/jsonstreamer/internal/services/jsonstreamersvc"
	"github.com/mikarios/jsonstreamer/internal/services/portcollectorsvc"
)

type portDomainService struct {
	jsonStreamingSvc     *jsonstreamersvc.JSONStreamer[*portmodel.PortData]
	portCollectorService *portcollectorsvc.PortCollectorService
}

func New(
	jsonStreamingSvc *jsonstreamersvc.JSONStreamer[*portmodel.PortData],
	portCollectorService *portcollectorsvc.PortCollectorService,
) *portDomainService {
	return &portDomainService{jsonStreamingSvc: jsonStreamingSvc, portCollectorService: portCollectorService}
}

func (pd *portDomainService) Start(ctx context.Context) {
	stream := pd.jsonStreamingSvc.Watch()

	go pd.jsonStreamingSvc.Start()
	go pd.portCollectorService.Start()

	go func() {
		for item := range stream {
			if item.Err != nil {
				logger.Error(ctx, item.Err, "problem getting item from stream")
				continue
			}

			pd.portCollectorService.NewPort(&portmodel.Port{Key: item.Key, Data: item.Data})
		}
	}()
}
