package exp

import (
	"context"
	"fmt"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
	"kibu.sh/starter/src/backend/database/models"
)

//kibu:service
type Service struct {
	Workflow Workflow__Client
}

//kibu:worker activity
type Activity struct {
	TxnProvider models.TxnProvider
}

//kibu:worker workflow
type Workflow struct {
	Activity Activity__Proxy
}

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

func (r Request) StartChildWorkflowOptions(opts temporal.ChildWorkflowOptions) temporal.ChildWorkflowOptions {
	return opts
}

//kibu:endpoint
func (s *Service) Index(ctx context.Context, req Request) (res Response, err error) {
	run, err := s.Workflow.Start(ctx, req)
	if err != nil {
		return
	}

	res, err = run.Get(ctx)
	if err != nil {
		return
	}

	return
}

//kibu:workflow
func (w *Workflow) Start(ctx workflow.Context, req Request) (res Response, err error) {
	ctx = temporal.WithDefaultActivityOptions(ctx)
	return w.Activity.DoSomeWork(ctx, req).Get(ctx)
}

//kibu:activity
func (a *Activity) DoSomeWork(ctx context.Context, req Request) (res Response, err error) {
	ctx, txn, err := a.TxnProvider(ctx)
	if err != nil {
		return
	}
	defer txn.RollbackOnErr(&err)

	_, err = txn.Querier().CheckConn(ctx)
	if err != nil {
		return
	}

	res.Message = fmt.Sprintf("Hello %s", req.Name)

	return
}
