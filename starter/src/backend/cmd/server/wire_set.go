package main

import (
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/wireset"
	"go.temporal.io/sdk/worker"
	"kibu.sh/starter/src/backend/gen/kibuwire"
)

func NewWorkerOptions() worker.Options {
	return worker.Options{
		MaxConcurrentActivityExecutionSize: 0,
	}
}

var wireSet = wire.NewSet(
	wireset.DefaultSet,
	kibuwire.SuperSet,

	// provided by system setup
	NewWorkerOptions,
)
