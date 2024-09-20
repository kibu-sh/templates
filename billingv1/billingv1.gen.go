package billingv1

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
)

////go:generate go run github.com/kibu-sh/kibu/cmd/kibu-gen@latest

const (
	barv1ServiceWatchBillingBillingName               = "barv1.WatchBillingBilling"
	barv1CustomerBillingWorkflowName                  = "barv1.customerBillingWorkflow"
	barv1CustomerBillingWorkflowAttemptPaymentName    = "barv1.customerBillingWorkflow.AttemptPayment"
	barv1CustomerBillingWorkflowGetAccountDetailsName = "barv1.customerBillingWorkflow.GetAccountDetails"
	barv1CustomerBillingWorkflowCancelBillingName     = "barv1.customerBillingWorkflow.CancelBilling"
	barv1CustomerBillingWorkflowSetDiscountName       = "barv1.customerBillingWorkflow.SetDiscount"
	barv1ActivitiesChargePaymentMethodName            = "barv1.activities.ChargePaymentMethod"
)

type GetHandleOpts struct {
	WorkflowID string
	RunID      string
}

type CustomerBillingWorkflowRun interface {
	ID() string
	RunID() string

	Get(ctx context.Context) (CustomerBillingResponse, error)

	AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.WorkflowOptionFunc) (res AttemptPaymentResponse, err error)

	GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest, mods ...temporal.WorkflowOptionFunc) (res GetAccountDetailsResult, err error)

	CancelBilling(ctx context.Context, req CancelBillingSignal, mods ...temporal.WorkflowOptionFunc) error
	SetDiscount(ctx context.Context, req SetDiscountSignal, mods ...temporal.WorkflowOptionFunc) error
}

type CustomerBillingWorkflowClient interface {
	Execute(ctx context.Context, req CustomerBillingRequest) (run CustomerBillingWorkflowRun, err error)
	GetHandle(ctx context.Context, ref GetHandleOpts) (run CustomerBillingWorkflowRun, err error)
	SignalWithStartSetDiscount(ctx context.Context, req SetDiscountSignal) (run CustomerBillingWorkflowRun, err error)
}

type CustomerBillingWorkflowChildRun interface {
	ID() string
	RunID() string
	IsReady() bool
	Underlying() workflow.ChildWorkflowFuture
	Get(ctx workflow.Context) (CustomerBillingResponse, error)
	CancelBilling(ctx workflow.Context, req CancelBillingSignal) error
	SetDiscount(ctx workflow.Context, req SetDiscountSignal) error
}

type CustomerBillingExternalRun interface {
	ID() string
	RunID() string

	RequestCancellation(ctx workflow.Context) error

	CancelBilling(ctx workflow.Context, req CancelBillingSignal) error
	CancelBillingAsync(ctx workflow.Context, req CancelBillingSignal) workflow.Future

	SetDiscount(ctx workflow.Context, req SetDiscountSignal) error
	SetDiscountAsync(ctx workflow.Context, req SetDiscountSignal) workflow.Future
}

type CustomerBillingWorkflowChildClient interface {
	Execute(ctx workflow.Context, req CustomerBillingRequest) (res CustomerBillingResponse, err error)
	ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest) CustomerBillingWorkflowChildRun
	External(ref GetHandleOpts) CustomerBillingExternalRun
}

type WorkflowsProxy interface {
	CustomerBilling() CustomerBillingWorkflowChildClient
}

type WorkflowsClient interface {
	CustomerBilling() CustomerBillingWorkflowClient
}

var _ WorkflowsClient = (*workflowsClient)(nil)
var _ WorkflowsProxy = (*workflowsProxy)(nil)
var _ CustomerBillingWorkflowClient = (*customerBillingWorkflowClient)(nil)
var _ CustomerBillingWorkflowChildClient = (*customerBillingWorkflowChildClient)(nil)
var _ CustomerBillingExternalRun = (*customerBillingExternalRun)(nil)
var _ CustomerBillingWorkflowRun = (*customerBillingWorkflowRun)(nil)
var _ CustomerBillingWorkflowChildRun = (*customerBillingWorkflowChildRun)(nil)

type workflowsClient struct{}

func (w workflowsClient) CustomerBilling() CustomerBillingWorkflowClient {
	//TODO implement me
	panic("implement me")
}

type workflowsProxy struct{}

func (w workflowsProxy) CustomerBilling() CustomerBillingWorkflowChildClient {
	//TODO implement me
	panic("implement me")
}

type customerBillingWorkflowClient struct{}

func (c customerBillingWorkflowClient) Execute(ctx context.Context, req CustomerBillingRequest) (run CustomerBillingWorkflowRun, err error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowClient) GetHandle(ctx context.Context, ref GetHandleOpts) (run CustomerBillingWorkflowRun, err error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowClient) SignalWithStartSetDiscount(ctx context.Context, req SetDiscountSignal) (run CustomerBillingWorkflowRun, err error) {
	//TODO implement me
	panic("implement me")
}

type customerBillingWorkflowChildClient struct{}

func (c customerBillingWorkflowChildClient) Execute(ctx workflow.Context, req CustomerBillingRequest) (res CustomerBillingResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildClient) ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest) CustomerBillingWorkflowChildRun {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildClient) External(ref GetHandleOpts) CustomerBillingExternalRun {
	//TODO implement me
	panic("implement me")
}

type customerBillingExternalRun struct{}

func (c customerBillingExternalRun) ID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) RunID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) RequestCancellation(ctx workflow.Context) error {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) CancelBilling(ctx workflow.Context, req CancelBillingSignal) error {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) CancelBillingAsync(ctx workflow.Context, req CancelBillingSignal) workflow.Future {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingExternalRun) SetDiscountAsync(ctx workflow.Context, req SetDiscountSignal) workflow.Future {
	//TODO implement me
	panic("implement me")
}

type customerBillingWorkflowRun struct{}

func (c customerBillingWorkflowRun) ID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) RunID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) Get(ctx context.Context) (CustomerBillingResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.WorkflowOptionFunc) (res AttemptPaymentResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest, mods ...temporal.WorkflowOptionFunc) (res GetAccountDetailsResult, err error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) CancelBilling(ctx context.Context, req CancelBillingSignal, mods ...temporal.WorkflowOptionFunc) error {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowRun) SetDiscount(ctx context.Context, req SetDiscountSignal, mods ...temporal.WorkflowOptionFunc) error {
	//TODO implement me
	panic("implement me")
}

type customerBillingWorkflowChildRun struct{}

func (c customerBillingWorkflowChildRun) ID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) RunID() string {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) IsReady() bool {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) Underlying() workflow.ChildWorkflowFuture {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) Get(ctx workflow.Context) (CustomerBillingResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) CancelBilling(ctx workflow.Context, req CancelBillingSignal) error {
	//TODO implement me
	panic("implement me")
}

func (c customerBillingWorkflowChildRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	//TODO implement me
	panic("implement me")
}
