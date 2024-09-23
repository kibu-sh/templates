package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"github.com/kibu-sh/kibu/pkg/wireset"
	"github.com/samber/lo"
	"go.temporal.io/sdk/worker"
	"kibu.sh/starter/src/backend/database/models"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

type WorkerSet struct {
	Billingv1 billingv1.Worker
}

func BuildWorkerSet(w *WorkerSet) []worker.Worker {
	return []worker.Worker{
		w.Billingv1.Build(),
	}
}

type ServiceSet struct {
	Billingv1 billingv1.ServiceController
}

func BuildServiceSet(s *ServiceSet, r transport.EndpointRegistry) []*httpx.Handler {
	s.Billingv1.Register(r)

	return lo.Map(r.GetByType(transport.EndpointTypeHTTP), func(ep transport.EndpointInfo, _ int) *httpx.Handler {
		meta := ep.Metadata.(transport.HTTPMetadata)
		return httpx.NewHandler(meta.Path, ep.Handler).WithMethods(meta.Method)
	})
}

func NewWorkerOptions() worker.Options {
	return worker.Options{
		MaxConcurrentActivityExecutionSize: 0,
	}
}

var wireSet = wire.NewSet(
	wireset.DefaultSet,
	billingv1.WireSet,
	models.NewTxnProvider,
	models.NewQuerier,
	models.NewConnPool,
	models.LoadConfig,
	BuildWorkerSet,
	BuildServiceSet,
	NewWorkerOptions,
	transport.NewEndpointRegistry,
	wire.Struct(new(WorkerSet), "*"),
	wire.Struct(new(ServiceSet), "*"),
)
