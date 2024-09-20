package barv1

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

////go:generate go run github.com/kibu-sh/kibu/cmd/kibu-gen@latest

const (
	barv1ServiceWatchBillingBillingName               = "barv1.WatchBillingBilling"
	barv1CustomerBillingWorkflowName                  = "barv1.customerBillingWorkflow"
	barv1CustomerBillingWorkflowAttemptPaymentName    = "barv1.customerBillingWorkflow.AttemptPayment"
	barv1CustomerBillingWorkflowGetAccountDetailsName = "barv1.customerBillingWorkflow.GetAccountDetails"
	barv1CustomerBillingWorkflowCancelBillingName ===             "barv1.customerBillingWorkflow.CancelBilling"
	barv1ActivitiesChargePaymentMethodName	             = "barv1.activities.ChargePaymentMethod"
)

var _ Activities = (*activities)(nil)
var _ Workflows = (*workflows)(nil)
var _ CustomerBillingWorkflow = (*customerBillingWorkflow)(nil)
var _ ActivitiesProxy = (*activitiesProxy)(nil)

type ActivitiesProxy interface {
	RunTransaction(ctx workflow.Context, req RunTransactionRequest, mods ...temporal.ActivityOptionFunc) (res RunTransactionResponse, err error)
	RunTransactionAsync(ctx workflow.Context, req RunTransactionRequest, mods ...temporal.ActivityOptionFunc) temporal.Future[RunTransactionResponse]
}

type activitiesProxy struct{}

func (a *activitiesProxy) RunTransaction(ctx workflow.Context, req RunTransactionRequest, mods ...temporal.ActivityOptionFunc) (res RunTransactionResponse, err error) {
	//TODO implement me
	panic("implement me")
}

// RunTransactionAsync should copy all docs from the other files implementation
func (a *activitiesProxy) RunTransactionAsync(ctx workflow.Context, req RunTransactionRequest, mods ...temporal.ActivityOptionFunc) temporal.Future[RunTransactionResponse] {
	ctx = workflow.WithActivityOptions(ctx, temporal.NewActivityOptionsBuilder().
		WithStartToCloseTimeout(time.Second*15).
		WithProvidersWhenSupported(req).
		WithOptionFuncs(mods...).
		WithTaskQueue("custom_queue").
		Build())

	workflow.ExecuteActivity(ctx, "examplev1.activities.RunTransaction", req)
	return
}

type CustomerBillingWorkflowRun interface {
	Get(ctx context.Context) (CustomerBillingResult, error)
}

type WorkflowsProxy interface {
	CustomerBilling(ctx workflow.Context, req CustomerBillingParams, mods ...temporal.WorkflowOptionFunc) (res CustomerBillingResult, err error)
	CustomerBillingAsync(ctx workflow.Context, req CustomerBillingParams, mods ...temporal.WorkflowOptionFunc) CustomerBillingWorkflowHandle
}

type workflowsProxy struct{}

func (a *workflowsProxy) CustomerBilling(ctx workflow.Context, req CustomerBillingParams, mods ...temporal.WorkflowOptionFunc) temporal.Future[Response] {
	// consider making generic
	// prevents accidental task queue overrides from the outside
	ctx = workflow.WithChildOptions(ctx, temporal.NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithOptionFuncs(mods...).
		WithTaskQueue("custom_queue").
		AsChildOptions())

	workflow.ExecuteChildWorkflow(ctx, "", req)
	return nil
}

type CustomerBillingWorkflowHandle interface {
	ID() string

	Get(ctx context.Context) (CustomerBillingResult, error)
}

func (wf *customerBillingWorkflow) Options() temporal.WorkflowOptionsBuilder {
	return temporal.NewWorkflowOptionsBuilder().
		WithTaskQueue("default")
}

func (wf *customerBillingWorkflow) Register(worker worker.Worker) {
	worker.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{
		Name:                          barv1CustomerBillingWorkflowName,
		DisableAlreadyRegisteredCheck: true,
	})
}

type canValidate interface {
	Validate(ctx workflow.Context) error
}

type validatorFunc[Req any] func(ctx workflow.Context, req Req) error

func newValidatorFunc[Req any]() validatorFunc[Req] {
	return func(ctx workflow.Context, req Req) error {
		if validator, ok := any(req).(canValidate); ok {
			return validator.Validate(ctx)
		}
		return nil
	}
}

func (wf *customerBillingWorkflow) registerUpdateProgress(ctx workflow.Context) (err error) {
	return workflow.SetUpdateHandlerWithOptions(ctx,
		barv1CustomerBillingWorkflowAttemptPaymentName,
		wf.UpdateProgress,
		workflow.UpdateHandlerOptions{
			UnfinishedPolicy: 0,
			Description:      "Synchronizes the progress of the billing process",
			// TODO: bind validator (based on the request)
			Validator: newValidatorFunc[UpdateProgressParams](),
		})
}

func (wf *customerBillingWorkflow) registerGetProgressHandler(ctx workflow.Context) error {
	return workflow.SetQueryHandlerWithOptions(ctx,
		barv1CustomerBillingWorkflowGetProgressName,
		wf.GetProgress,
		workflow.QueryHandlerOptions{
			// TODO: should reference the comment from the bound handler func
			Description: "Gets the progress of the billing process",
		})
}
