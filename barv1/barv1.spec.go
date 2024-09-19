package barv1

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

type Request struct{}
type Response struct{}

//kibu:service
type Service interface {
	// MyRequestHandler lets you trigger some work from an HTTP endpoint
	//
	//kibu:endpoint method=POST
	MyRequestHandler(ctx context.Context, req Request) (res Response, err error)
}

type RunTransactionParams struct{}
type RunTransactionResult struct{}

//kibu:worker activity task_queue=custom_queue
type Activities interface {
	// RunTransaction performs work against another transactional system
	RunTransaction(ctx context.Context, req Request) (res Response, err error)
}

type LongWorkflowParams struct {
}

type LongWorkflowResult struct{}

//kibu:worker workflow task_queue=custom_queue
type Workflows interface {
	// StartSomeLongProcess starts a workflow
	StartSomeLongProcess(ctx workflow.Context, req Request) (res Response, err error)
}
