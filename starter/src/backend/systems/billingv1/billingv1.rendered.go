package billingv1

import (
	"context"
	"github.com/google/wire"
	"github.com/kibu-sh/kibu/pkg/transport"
	"github.com/kibu-sh/kibu/pkg/transport/httpx"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"time"

	"github.com/kibu-sh/kibu/pkg/transport/temporal"
)

// compile time check to ensure implementations are correct
var _ Service = (*service)(nil)
var _ Activities = (*activities)(nil)
var _ ActivitiesProxy = (*activitiesProxy)(nil)
var _ CustomerSubscriptionsWorkflow = (*customerSubscriptionsWorkflow)(nil)
var _ CustomerSubscriptionsWorkflowChildRun = (*customerSubscriptionsChildRun)(nil)
var _ CustomerSubscriptionsWorkflowChildClient = (*customerSubscriptionsChildClient)(nil)

const (
	packageName                                        = "billingv1"
	customerSubscriptionsWorkflowName                  = "billingv1.customerSubscriptions"
	customerSubscriptionsWorkflowAttemptPaymentName    = "billingv1.customerSubscriptions.AttemptPayment"
	customerSubscriptionsWorkflowGetAccountDetailsName = "billingv1.customerSubscriptions.GetAccountDetails"
	customerSubscriptionsWorkflowCancelBillingName     = "billingv1.customerSubscriptions.CancelBillingRequest"
	customerSubscriptionsWorkflowSetDiscountName       = "billingv1.customerSubscriptions.SetDiscount"
	activitiesChargePaymentMethodName                  = "billingv1.activities.ChargePaymentMethod"
)

func NewSetDiscountSignalChannel(ctx workflow.Context) temporal.SignalChannel[SetDiscountRequest] {
	return temporal.NewSignalChannel[SetDiscountRequest](ctx, customerSubscriptionsWorkflowSetDiscountName)
}

func NewCancelBillingSignalChannel(ctx workflow.Context) temporal.SignalChannel[CancelBillingRequest] {
	return temporal.NewSignalChannel[CancelBillingRequest](ctx, customerSubscriptionsWorkflowCancelBillingName)
}

type CustomerSubscriptionsWorkflowRun interface {
	WorkflowID() string
	RunID() string
	Get(ctx context.Context) (CustomerSubscriptionsResponse, error)

	CancelBilling(ctx context.Context, req CancelBillingRequest) error
	SetDiscount(ctx context.Context, req SetDiscountRequest) error

	GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest) (GetAccountDetailsResponse, error)

	AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.UpdateOptionFunc) (AttemptPaymentResponse, error)
	AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.UpdateOptionFunc) (temporal.UpdateHandle[AttemptPaymentResponse], error)
}

type CustomerSubscriptionsWorkflowChildRun interface {
	WorkflowID() string
	IsReady() bool
	Underlying() workflow.ChildWorkflowFuture
	Get(ctx workflow.Context) (CustomerSubscriptionsResponse, error)
	CancelBilling(ctx workflow.Context, req CancelBillingRequest) error
	SetDiscount(ctx workflow.Context, req SetDiscountRequest) error

	WaitStart(ctx workflow.Context) (*workflow.Execution, error)
	SelectStart(sel workflow.Selector, fn func(CustomerSubscriptionsWorkflowChildRun)) workflow.Selector
	Select(sel workflow.Selector, fn func(CustomerSubscriptionsWorkflowChildRun)) workflow.Selector
}

type CustomerSubscriptionsExternalRun interface {
	WorkflowID() string
	RunID() string

	RequestCancellation(ctx workflow.Context) error

	CancelBilling(ctx workflow.Context, req CancelBillingRequest) error
	CancelBillingAsync(ctx workflow.Context, req CancelBillingRequest) workflow.Future

	SetDiscount(ctx workflow.Context, req SetDiscountRequest) error
	SetDiscountAsync(ctx workflow.Context, req SetDiscountRequest) workflow.Future
}

type CustomerSubscriptionsWorkflowClient interface {
	GetHandle(ctx context.Context, ref temporal.GetHandleOpts) (CustomerSubscriptionsWorkflowRun, error)
	Execute(ctx context.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error)
	ExecuteWithSetDiscount(ctx context.Context, req *CustomerSubscriptionsRequest, sig SetDiscountRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error)
}

type CustomerSubscriptionsWorkflowChildClient interface {
	Execute(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsResponse, error)
	ExecuteAsync(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) CustomerSubscriptionsWorkflowChildRun
	External(ref temporal.GetHandleOpts) CustomerSubscriptionsExternalRun
}

// ActivitiesProxy is a workflow interface for Activities
type ActivitiesProxy interface {
	ChargePaymentMethod(ctx workflow.Context, req ChargePaymentMethodRequest, mods ...temporal.ActivityOptionFunc) (res ChargePaymentMethodResponse, err error)
	ChargePaymentMethodAsync(ctx workflow.Context, req ChargePaymentMethodRequest, mods ...temporal.ActivityOptionFunc) temporal.Future[ChargePaymentMethodResponse]
}

type WorkflowsProxy interface {
	CustomerSubscriptions() CustomerSubscriptionsWorkflowChildClient
}

type WorkflowsClient interface {
	CustomerSubscriptions() CustomerSubscriptionsWorkflowClient
}

type activitiesProxy struct{}

func (a *activitiesProxy) ChargePaymentMethod(
	ctx workflow.Context,
	req ChargePaymentMethodRequest,
	mods ...temporal.ActivityOptionFunc,
) (res ChargePaymentMethodResponse, err error) {
	return a.ChargePaymentMethodAsync(ctx, req, mods...).Get(ctx)
}

// ChargePaymentMethodAsync proxies to Activities.ChargePaymentMethod
func (a *activitiesProxy) ChargePaymentMethodAsync(
	ctx workflow.Context,
	req ChargePaymentMethodRequest,
	mods ...temporal.ActivityOptionFunc,
) temporal.Future[ChargePaymentMethodResponse] {
	options := temporal.NewActivityOptionsBuilder().
		WithStartToCloseTimeout(time.Second * 30).
		WithTaskQueue(packageName).
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		Build()

	ctx = workflow.WithActivityOptions(ctx, options)
	future := workflow.ExecuteActivity(ctx, activitiesChargePaymentMethodName, req)
	return temporal.NewFuture[ChargePaymentMethodResponse](future)
}

type workflowsClient struct {
	client client.Client
}

func (w *workflowsClient) CustomerSubscriptions() CustomerSubscriptionsWorkflowClient {
	return &customerSubscriptionsClient{
		client: w.client,
	}
}

type workflowsProxy struct{}

func (w *workflowsProxy) CustomerSubscriptions() CustomerSubscriptionsWorkflowChildClient {
	return &customerSubscriptionsChildClient{}
}

type customerSubscriptionsClient struct {
	client client.Client
}

func (c *customerSubscriptionsClient) Execute(ctx context.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		WithTaskQueue(packageName).
		AsStartOptions()

	we, err := c.client.ExecuteWorkflow(ctx, options, customerSubscriptionsWorkflowName, req)
	if err != nil {
		return nil, err
	}

	return &customerSubscriptionsWorkflowRun{
		client:      c.client,
		workflowRun: we,
	}, nil
}

func (c *customerSubscriptionsClient) GetHandle(ctx context.Context, ref temporal.GetHandleOpts) (CustomerSubscriptionsWorkflowRun, error) {
	return &customerSubscriptionsWorkflowRun{
		client:      c.client,
		workflowRun: c.client.GetWorkflow(ctx, ref.WorkflowID, ref.RunID),
	}, nil
}

func (c *customerSubscriptionsClient) ExecuteWithSetDiscount(ctx context.Context, req *CustomerSubscriptionsRequest, sig SetDiscountRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(sig).
		WithOptions(mods...).
		WithTaskQueue(packageName).
		AsStartOptions()

	run, err := c.client.SignalWithStartWorkflow(ctx,
		options.ID,
		customerSubscriptionsWorkflowSetDiscountName,
		sig,
		options,
		customerSubscriptionsWorkflowName,
		req)

	if err != nil {
		return nil, err
	}

	return &customerSubscriptionsWorkflowRun{
		client:      c.client,
		workflowRun: run,
	}, nil
}

type customerSubscriptionsChildClient struct{}

func (c *customerSubscriptionsChildClient) Execute(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) (CustomerSubscriptionsResponse, error) {
	return c.ExecuteAsync(ctx, req, mods...).Get(ctx)
}

func (c *customerSubscriptionsChildClient) ExecuteAsync(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...temporal.WorkflowOptionFunc) CustomerSubscriptionsWorkflowChildRun {
	options := temporal.NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		WithTaskQueue(packageName).
		AsChildOptions()

	ctx = workflow.WithChildOptions(ctx, options)
	childFuture := workflow.ExecuteChildWorkflow(ctx, customerSubscriptionsWorkflowName, req)

	return &customerSubscriptionsChildRun{
		childFuture: childFuture,
	}
}

func (c *customerSubscriptionsChildClient) External(ref temporal.GetHandleOpts) CustomerSubscriptionsExternalRun {
	return &customerSubscriptionsExternalRun{
		workflowID: ref.WorkflowID,
		runID:      ref.RunID,
	}
}

type customerSubscriptionsExternalRun struct {
	workflowID string
	runID      string
}

func (c *customerSubscriptionsExternalRun) WorkflowID() string {
	return c.workflowID
}

func (c *customerSubscriptionsExternalRun) RunID() string {
	return c.runID
}

func (c *customerSubscriptionsExternalRun) RequestCancellation(ctx workflow.Context) error {
	return workflow.RequestCancelExternalWorkflow(ctx, c.workflowID, c.runID).Get(ctx, nil)
}

func (c *customerSubscriptionsExternalRun) CancelBilling(ctx workflow.Context, req CancelBillingRequest) error {
	return c.CancelBillingAsync(ctx, req).Get(ctx, nil)
}

func (c *customerSubscriptionsExternalRun) CancelBillingAsync(ctx workflow.Context, req CancelBillingRequest) workflow.Future {
	return workflow.SignalExternalWorkflow(ctx, c.workflowID, c.runID, customerSubscriptionsWorkflowCancelBillingName, req)
}

func (c *customerSubscriptionsExternalRun) SetDiscount(ctx workflow.Context, req SetDiscountRequest) error {
	return c.SetDiscountAsync(ctx, req).Get(ctx, nil)
}

func (c *customerSubscriptionsExternalRun) SetDiscountAsync(ctx workflow.Context, req SetDiscountRequest) workflow.Future {
	return workflow.SignalExternalWorkflow(ctx, c.workflowID, c.runID, customerSubscriptionsWorkflowSetDiscountName, req)
}

type customerSubscriptionsWorkflowRun struct {
	client      client.Client
	workflowRun client.WorkflowRun
}

func (c *customerSubscriptionsWorkflowRun) WorkflowID() string {
	return c.workflowRun.GetID()
}

func (c *customerSubscriptionsWorkflowRun) RunID() string {
	return c.workflowRun.GetRunID()
}

func (c *customerSubscriptionsWorkflowRun) Get(ctx context.Context) (CustomerSubscriptionsResponse, error) {
	var result CustomerSubscriptionsResponse
	err := c.workflowRun.Get(ctx, &result)
	return result, err
}

func (c *customerSubscriptionsWorkflowRun) AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.UpdateOptionFunc) (temporal.UpdateHandle[AttemptPaymentResponse], error) {
	options := temporal.NewUpdateOptionsBuilder().
		WithUpdateName(customerSubscriptionsWorkflowAttemptPaymentName).
		WithWorkflowID(c.WorkflowID()).
		WithRunID(c.RunID()).
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		WithArgs(req).
		Build()

	handle, err := c.client.UpdateWorkflow(ctx, options)
	if err != nil {
		return nil, err
	}

	return temporal.NewUpdateHandle[AttemptPaymentResponse](handle), nil
}

func (c *customerSubscriptionsWorkflowRun) AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...temporal.UpdateOptionFunc) (AttemptPaymentResponse, error) {
	handle, err := c.AttemptPaymentAsync(ctx, req, mods...)
	if err != nil {
		return AttemptPaymentResponse{}, err
	}
	return handle.Get(ctx)
}

func (c *customerSubscriptionsWorkflowRun) GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest) (GetAccountDetailsResponse, error) {
	queryResponse, err := c.client.QueryWorkflow(ctx, c.WorkflowID(), c.RunID(),
		customerSubscriptionsWorkflowGetAccountDetailsName, req)

	if err != nil {
		return GetAccountDetailsResponse{}, err
	}

	var result GetAccountDetailsResponse
	err = queryResponse.Get(&result)
	return result, err
}

func (c *customerSubscriptionsWorkflowRun) CancelBilling(ctx context.Context, req CancelBillingRequest) error {
	return c.client.SignalWorkflow(ctx, c.WorkflowID(), c.RunID(),
		customerSubscriptionsWorkflowCancelBillingName, req)
}

func (c *customerSubscriptionsWorkflowRun) SetDiscount(ctx context.Context, req SetDiscountRequest) error {
	return c.client.SignalWorkflow(ctx, c.WorkflowID(), c.RunID(),
		customerSubscriptionsWorkflowSetDiscountName, req)
}

type customerSubscriptionsChildRun struct {
	workflowId  string
	childFuture workflow.ChildWorkflowFuture
}

func (c *customerSubscriptionsChildRun) WorkflowID() string {
	return c.workflowId
}

func (c *customerSubscriptionsChildRun) IsReady() bool {
	return c.childFuture.IsReady()
}

func (c *customerSubscriptionsChildRun) Underlying() workflow.ChildWorkflowFuture {
	return c.childFuture
}

func (c *customerSubscriptionsChildRun) Get(ctx workflow.Context) (CustomerSubscriptionsResponse, error) {
	var result CustomerSubscriptionsResponse
	err := c.childFuture.Get(ctx, &result)
	return result, err
}

func (c *customerSubscriptionsChildRun) CancelBilling(ctx workflow.Context, req CancelBillingRequest) error {
	return c.childFuture.SignalChildWorkflow(ctx,
		customerSubscriptionsWorkflowCancelBillingName, req).Get(ctx, nil)
}

func (c *customerSubscriptionsChildRun) SetDiscount(ctx workflow.Context, req SetDiscountRequest) error {
	return c.childFuture.SignalChildWorkflow(ctx,
		customerSubscriptionsWorkflowSetDiscountName, req).Get(ctx, nil)
}

func (c *customerSubscriptionsChildRun) Select(sel workflow.Selector, fn func(CustomerSubscriptionsWorkflowChildRun)) workflow.Selector {
	return sel.AddFuture(c.childFuture, func(workflow.Future) {
		if fn != nil {
			fn(c)
		}
	})
}

func (c *customerSubscriptionsChildRun) SelectStart(sel workflow.Selector, fn func(CustomerSubscriptionsWorkflowChildRun)) workflow.Selector {
	return sel.AddFuture(c.childFuture.GetChildWorkflowExecution(), func(workflow.Future) {
		if fn != nil {
			fn(c)
		}
	})
}

func (c *customerSubscriptionsChildRun) WaitStart(ctx workflow.Context) (*workflow.Execution, error) {
	var exec workflow.Execution
	err := c.childFuture.GetChildWorkflowExecution().Get(ctx, &exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

type customerSubscriptionsWorkflowInput struct {
	Request              CustomerSubscriptionsRequest
	SetDiscountChannel   temporal.SignalChannel[SetDiscountRequest]
	CancelBillingChannel temporal.SignalChannel[SetDiscountRequest]
}

type CustomerSubscriptionsWorkflowFactory func(input *customerSubscriptionsWorkflowInput) (CustomerSubscriptionsWorkflow, error)

type CustomerSubscriptionsWorkflowController struct {
	Factory CustomerSubscriptionsWorkflowFactory
}

func (wk *CustomerSubscriptionsWorkflowController) Execute(ctx workflow.Context, req CustomerSubscriptionsRequest) (res CustomerSubscriptionsResponse, err error) {
	input := &customerSubscriptionsWorkflowInput{
		Request:              req,
		SetDiscountChannel:   NewSetDiscountSignalChannel(ctx),
		CancelBillingChannel: NewSetDiscountSignalChannel(ctx),
	}

	wf, err := wk.Factory(input)
	if err != nil {
		return
	}

	if err = workflow.SetQueryHandler(ctx,
		customerSubscriptionsWorkflowGetAccountDetailsName,
		wf.GetAccountDetails,
	); err != nil {
		return
	}

	if err = workflow.SetUpdateHandlerWithOptions(ctx,
		customerSubscriptionsWorkflowAttemptPaymentName,
		wf.AttemptPayment,
		workflow.UpdateHandlerOptions{
			Validator:        nil,
			UnfinishedPolicy: 0,
			Description:      "",
		},
	); err != nil {
		return
	}

	return wf.Execute(ctx, req)
}

func (wk *CustomerSubscriptionsWorkflowController) Build(registry worker.WorkflowRegistry) {
	registry.RegisterWorkflowWithOptions(wk.Execute, workflow.RegisterOptions{
		Name:                          customerSubscriptionsWorkflowName,
		DisableAlreadyRegisteredCheck: true,
	})
}

type ServiceController struct {
	Service Service
}

func (svc *ServiceController) Build() []*httpx.Handler {
	return []*httpx.Handler{
		httpx.NewHandler("/billingv1/WatchAccount",
			transport.NewEndpoint(svc.Service.WatchAccount)).
			WithMethods("POST"),
	}
}

type ActivitiesController struct {
	Activities Activities
}

func (act *ActivitiesController) Build(registry worker.ActivityRegistry) {
	registry.RegisterActivityWithOptions(act.Activities.ChargePaymentMethod, activity.RegisterOptions{
		Name:                          activitiesChargePaymentMethodName,
		DisableAlreadyRegisteredCheck: true,
	})
}

type WorkerController struct {
	Client                                  client.Client
	Options                                 worker.Options
	ActivitiesController                    ActivitiesController
	CustomerSubscriptionsWorkflowController CustomerSubscriptionsWorkflowController
}

func (wc *WorkerController) Build() worker.Worker {
	wk := worker.New(wc.Client, packageName, wc.Options)
	wc.ActivitiesController.Build(wk)
	wc.CustomerSubscriptionsWorkflowController.Build(wk)
	return wk
}

func NewActivitiesProxy() ActivitiesProxy {
	return &activitiesProxy{}
}

func NewWorkflowsProxy() WorkflowsProxy {
	return &workflowsProxy{}
}

func NewWorkflowsClient(client client.Client) WorkflowsClient {
	return &workflowsClient{
		client: client,
	}
}

var WireSet = wire.NewSet(
	NewService,
	NewActivities,
	NewActivitiesProxy,
	NewWorkflowsProxy,
	NewWorkflowsClient,
	NewCustomerSubscriptionsWorkflowFactory,
	wire.Struct(new(WorkerController), "*"),
	wire.Struct(new(ServiceController), "*"),
	wire.Struct(new(ActivitiesController), "*"),
	wire.Struct(new(CustomerSubscriptionsWorkflowController), "*"),
)
