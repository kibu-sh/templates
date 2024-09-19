package foov1

import (
	"barv1"
	"cuelang.org/go/pkg/time"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
)

//kibu:workflows
type Workflows interface {
	ScheduledWork(ctx workflow.Context)
}

type workflows struct {
	BarV1 barv1.WorkflowsProxy
}

func (w *workflows) ScheduledWork(ctx workflow.Context) {
	run := w.BarV1.StartSomeLongProcessAsync(ctx, barv1.LongWorkflowParams{}, func(builder temporal.WorkflowOptionsBuilder) temporal.WorkflowOptionsBuilder {
		return builder.
			WithEnableEagerStart(true).
			WithStartDelay(10 * time.Second)
	})

	run.IsReady()

	return
}
