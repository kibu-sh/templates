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
	run := w.BarV1.CustomerBillingAsync(ctx, barv1.CustomerBillingParams{}, withCustomStartDelay)

	run.Get(ctx)
	return
}

func withCustomStartDelay(builder temporal.WorkflowOptionsBuilder) temporal.WorkflowOptionsBuilder {
	return builder.
		WithEnableEagerStart(true).
		WithStartDelay(10 * time.Second)
}
