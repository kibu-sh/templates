package barv1

import (
	"cuelang.org/go/pkg/time"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
)

////go:generate go run github.com/kibu-sh/kibu/cmd/kibu-gen@latest

var _ Activities = (*activities)(nil)
var _ ActivitiesProxy = (*activitiesProxy)(nil)

type ActivitiesProxy interface {
	RunTransaction(ctx workflow.Context, req RunTransactionParams, mods ...temporal.ActivityOptionFunc) (res RunTransactionResult, err error)
	RunTransactionAsync(ctx workflow.Context, req RunTransactionParams, mods ...temporal.ActivityOptionFunc) temporal.Future[RunTransactionResult]
}

type activitiesProxy struct{}

func (a *activitiesProxy) RunTransaction(ctx workflow.Context, req RunTransactionParams, mods ...temporal.ActivityOptionFunc) (res RunTransactionResult, err error) {
	//TODO implement me
	panic("implement me")
}

func (*RunTransactionParams) activityOptions() temporal.ActivityOptionsBuilder {
	return temporal.NewActivityOptionsBuilder().
		WithStartToCloseTimeout(time.Second * 15)
}

// RunTransactionAsync should copy all docs from the other files implementation
func (a *activitiesProxy) RunTransactionAsync(ctx workflow.Context, req RunTransactionParams, mods ...temporal.ActivityOptionFunc) temporal.Future[RunTransactionResult] {
	ctx = workflow.WithActivityOptions(ctx, req.activityOptions().
		WithTaskQueue("custom_queue").
		WithOptionFuncs(mods...).
		Build())

	workflow.ExecuteActivity(ctx, "examplev1.activities.RunTransaction", req)
	return
}

var _ Workflows = (*workflows)(nil)

type WorkflowsProxy interface {
	StartSomeLongProcess(ctx workflow.Context, req LongWorkflowParams, mods ...temporal.WorkflowOptionFunc) (res Response, err error)
	StartSomeLongProcessAsync(ctx workflow.Context, req LongWorkflowParams, mods ...temporal.WorkflowOptionFunc) temporal.Future[Response]
}

type workflowsProxy struct{}

func (a *workflowsProxy) StartSomeLongProcessAsync(ctx workflow.Context, req LongWorkflowParams, mods ...temporal.WorkflowOptionFunc) temporal.Future[Response] {
	// consider making generic
	// prevents accidental task queue overrides from the outside
	ctx = workflow.WithChildOptions(ctx, req.workflowOptions().
		WithOptionFuncs(mods...).
		WithTaskQueue("custom_queue").
		AsChildOptions())

	workflow.ExecuteChildWorkflow(ctx, "", req)
	return nil
}

func (l LongWorkflowParams) workflowOptions() temporal.WorkflowOptionsBuilder {
	return temporal.NewWorkflowOptionsBuilder()
}
