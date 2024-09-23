package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"github.com/kibu-sh/kibu/pkg/wireset"
	"go.temporal.io/sdk/worker"
	"kibu.sh/starter/src/backend/database/models"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

type WorkerSet struct {
	Billingv1 billingv1.WorkerController
}

func BuildWorkerSet(w *WorkerSet) []worker.Worker {
	return []worker.Worker{
		w.Billingv1.Build(),
	}
}

type ServiceSet struct {
	//Billingv1 billingv1.ServiceController
}

func BuildServiceSet(s *ServiceSet) []*httpx.Handler {
	return []*httpx.Handler{}
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
	wire.Struct(new(WorkerSet), "*"),
	wire.Struct(new(ServiceSet), "*"),
)
