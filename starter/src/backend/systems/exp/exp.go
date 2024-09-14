package exp

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

//kibu:service
type Service struct{}

//kibu:worker activity
type Activity struct{}

//kibu:worker workflow
type Workflow struct{}

//kibu:endpoint
func (s *Service) Index(ctx context.Context, req any) (res any, err error) {
	return
}

//kibu:workflow
func (w *Workflow) Start(ctx workflow.Context, req any) (res any, err error) {
	return
}

//kibu:activity
func (a *Activity) DoSomeWork(ctx context.Context, req any) (res any, err error) {
	return
}
