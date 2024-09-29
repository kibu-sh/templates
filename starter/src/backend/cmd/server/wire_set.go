package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"github.com/kibu-sh/kibu/pkg/wireset"
	"go.temporal.io/sdk/worker"
	"kibu.sh/starter/src/backend/database/models"
	"kibu.sh/starter/src/backend/systems/billingv1"
	"kibu.sh/starter/src/backend/systems/billingv1/activities"
	"kibu.sh/starter/src/backend/systems/billingv1/services"
	"kibu.sh/starter/src/backend/systems/billingv1/workflows"
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
	Billingv1 billingv1.ServiceController
}

func BuildServiceSet(s *ServiceSet) (handlers []*httpx.Handler) {
	handlers = append(handlers, s.Billingv1.Build()...)
	return
}

func NewWorkerOptions() worker.Options {
	return worker.Options{
		MaxConcurrentActivityExecutionSize: 0,
	}
}

var wireSet = wire.NewSet(
	wireset.DefaultSet,

	// provided by system setup
	BuildWorkerSet,
	BuildServiceSet,
	NewWorkerOptions,

	// inject controllers
	wire.Struct(new(WorkerSet), "*"),
	wire.Struct(new(ServiceSet), "*"),

	// import billing wire set
	billingv1.WireSet,
	services.NewService,
	activities.NewActivities,
	workflows.NewCustomerSubscriptionsWorkflowFactory,

	// import database wire set
	models.NewTxnProvider,
	models.NewQuerier,
	models.NewConnPool,
	models.LoadConfig,
)
