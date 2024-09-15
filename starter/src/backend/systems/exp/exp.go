package exp

import (
	"context"
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

//kibu:endpoint
func (s *Service) Index(ctx context.Context, req any) (res any, err error) {
	run, err := s.Workflow.Start(ctx, "", req)
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
func (w *Workflow) Start(ctx workflow.Context, req any) (res any, err error) {
	return w.Activity.DoSomeWork(ctx, req).Get(ctx)
}

//kibu:activity
func (a *Activity) DoSomeWork(ctx context.Context, req any) (res any, err error) {
	ctx, txn, err := a.TxnProvider(ctx)
	if err != nil {
		return
	}
	defer txn.RollbackOnErr(&err)

	check, err := txn.Querier().CheckConn(ctx)
	if err != nil {
		return
	}

	return check, nil
}
