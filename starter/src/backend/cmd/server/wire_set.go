package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/kibu-sh/kibu/pkg/wireset"
	"go.temporal.io/sdk/worker"
	"kibu.sh/starter/src/backend/database/models"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

var wireSet = wire.NewSet(
	wireset.DefaultSet,
	billingv1.WireSet,
	models.NewTxnProvider,
	models.NewQuerier,
	models.NewConnPool,
	BuildWorkerSet,
	wire.Struct(new(WorkerSet), "*"),
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

func BuildServiceSet(s *ServiceSet, r transport.EndpointRegistry) []worker.Worker {
	s.Billingv1.Register(r)
	return []worker.Worker{}
}
