package billingv1

import (
	"context"
	. "github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTaskQueueName                               = "billingv1"
	serviceWatchCustomerAccountName                    = "billingv1.WatchCustomerAccount"
	customerSubscriptionsWorkflowName                  = "billingv1.customerSubscriptions"
	customerSubscriptionsWorkflowAttemptPaymentName    = "billingv1.customerSubscriptions.AttemptPayment"
	customerSubscriptionsWorkflowGetAccountDetailsName = "billingv1.customerSubscriptions.GetAccountDetails"
	customerSubscriptionsWorkflowCancelBillingName     = "billingv1.customerSubscriptions.CancelBillingRequest"
	customerSubscriptionsWorkflowSetDiscountName       = "billingv1.customerSubscriptions.SetDiscount"
	activitiesChargePaymentMethodName                  = "billingv1.activities.ChargePaymentMethod"
)

func NewSetDiscountSignalChannel(ctx workflow.Context) SignalChannel[SetDiscountRequest] {
	return NewSignalChannel[SetDiscountRequest](ctx, customerSubscriptionsWorkflowSetDiscountName)
}

func NewCancelBillingSignalChannel(ctx workflow.Context) SignalChannel[CancelBillingRequest] {
	return NewSignalChannel[CancelBillingRequest](ctx, customerSubscriptionsWorkflowCancelBillingName)
}

type CustomerSubscriptionsWorkflowRun interface {
	ID() string
	RunID() string
	Get(ctx context.Context) (CustomerSubscriptionsResponse, error)

	CancelBilling(ctx context.Context, req CancelBillingRequest) error
	SetDiscount(ctx context.Context, req SetDiscountRequest) error

	GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest) (GetAccountDetailsResult, error)

	AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (AttemptPaymentResponse, error)
	AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (UpdateHandle[AttemptPaymentResponse], error)
}

type CustomerSubscriptionsWorkflowClient interface {
	GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerSubscriptionsWorkflowRun, error)
	Execute(ctx context.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error)
	ExecuteWithSetDiscount(ctx context.Context, req *CustomerSubscriptionsRequest, sig SetDiscountRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error)
}

type CustomerSubscriptionsWorkflowChildRun interface {
	ID() string
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
	ID() string
	RunID() string

	RequestCancellation(ctx workflow.Context) error

	CancelBilling(ctx workflow.Context, req CancelBillingRequest) error
	CancelBillingAsync(ctx workflow.Context, req CancelBillingRequest) workflow.Future

	SetDiscount(ctx workflow.Context, req SetDiscountRequest) error
	SetDiscountAsync(ctx workflow.Context, req SetDiscountRequest) workflow.Future
}

type CustomerSubscriptionsWorkflowChildClient interface {
	Execute(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsResponse, error)
	ExecuteAsync(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) CustomerSubscriptionsWorkflowChildRun
	External(ref GetHandleOpts) CustomerSubscriptionsExternalRun
}

type WorkflowsProxy interface {
	CustomerSubscriptions() CustomerSubscriptionsWorkflowChildClient
}

type WorkflowsClient interface {
	CustomerSubscriptions() CustomerSubscriptionsWorkflowClient
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

func (c *customerSubscriptionsClient) Execute(ctx context.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error) {
	options := NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		WithTaskQueue(defaultTaskQueueName).
		AsStartOptions()

	we, err := c.client.ExecuteWorkflow(ctx, options, customerSubscriptionsWorkflowName, req)
	if err != nil {
		return nil, err
	}

	return &customerSubscriptionsRun{
		client:      c.client,
		workflowRun: we,
	}, nil
}

func (c *customerSubscriptionsClient) GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerSubscriptionsWorkflowRun, error) {
	return &customerSubscriptionsRun{
		client:      c.client,
		workflowRun: c.client.GetWorkflow(ctx, ref.WorkflowID, ref.RunID),
	}, nil
}

func (c *customerSubscriptionsClient) ExecuteWithSetDiscount(ctx context.Context, req *CustomerSubscriptionsRequest, sig SetDiscountRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsWorkflowRun, error) {
	options := NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(sig).
		WithOptions(mods...).
		WithTaskQueue(defaultTaskQueueName).
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

	return &customerSubscriptionsRun{
		client:      c.client,
		workflowRun: run,
	}, nil
}

type customerSubscriptionsChildClient struct{}

func (c *customerSubscriptionsChildClient) Execute(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) (CustomerSubscriptionsResponse, error) {
	return c.ExecuteAsync(ctx, req, mods...).Get(ctx)
}

func (c *customerSubscriptionsChildClient) ExecuteAsync(ctx workflow.Context, req CustomerSubscriptionsRequest, mods ...WorkflowOptionFunc) CustomerSubscriptionsWorkflowChildRun {
	options := NewWorkflowOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithOptions(mods...).
		WithTaskQueue(defaultTaskQueueName).
		AsChildOptions()

	ctx = workflow.WithChildOptions(ctx, options)
	childFuture := workflow.ExecuteChildWorkflow(ctx, customerSubscriptionsWorkflowName, req)

	return &customerSubscriptionsChildRun{
		childFuture: childFuture,
	}
}

func (c *customerSubscriptionsChildClient) External(ref GetHandleOpts) CustomerSubscriptionsExternalRun {
	return &customerSubscriptionsExternalRun{
		workflowID: ref.WorkflowID,
		runID:      ref.RunID,
	}
}

type customerSubscriptionsExternalRun struct {
	workflowID string
	runID      string
}

func (c *customerSubscriptionsExternalRun) ID() string {
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

type customerSubscriptionsRun struct {
	client      client.Client
	workflowRun client.WorkflowRun
}

func (c *customerSubscriptionsRun) ID() string {
	return c.workflowRun.GetID()
}

func (c *customerSubscriptionsRun) RunID() string {
	return c.workflowRun.GetRunID()
}

func (c *customerSubscriptionsRun) Get(ctx context.Context) (CustomerSubscriptionsResponse, error) {
	var result CustomerSubscriptionsResponse
	err := c.workflowRun.Get(ctx, &result)
	return result, err
}

func (c *customerSubscriptionsRun) AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (UpdateHandle[AttemptPaymentResponse], error) {
	options := NewUpdateOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithWorkflowID(c.ID()).
		WithRunID(c.RunID()).
		WithOptions(mods...).
		WithArgs(req).
		Build()

	handle, err := c.client.UpdateWorkflow(ctx, options)
	if err != nil {
		return nil, err
	}

	return NewUpdateHandle[AttemptPaymentResponse](handle), nil
}

func (c *customerSubscriptionsRun) AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (AttemptPaymentResponse, error) {
	handle, err := c.AttemptPaymentAsync(ctx, req, mods...)
	if err != nil {
		return AttemptPaymentResponse{}, err
	}
	return handle.Get(ctx)
}

func (c *customerSubscriptionsRun) GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest) (GetAccountDetailsResult, error) {
	queryResponse, err := c.client.QueryWorkflow(ctx, c.ID(), c.RunID(),
		customerSubscriptionsWorkflowGetAccountDetailsName, req)
	if err != nil {
		return GetAccountDetailsResult{}, err
	}

	var result GetAccountDetailsResult
	err = queryResponse.Get(&result)
	return result, err
}

func (c *customerSubscriptionsRun) CancelBilling(ctx context.Context, req CancelBillingRequest) error {
	return c.client.SignalWorkflow(ctx, c.ID(), c.RunID(), customerSubscriptionsWorkflowCancelBillingName, req)
}

func (c *customerSubscriptionsRun) SetDiscount(ctx context.Context, req SetDiscountRequest) error {
	return c.client.SignalWorkflow(ctx, c.ID(), c.RunID(), customerSubscriptionsWorkflowSetDiscountName, req)
}

type customerSubscriptionsChildRun struct {
	workflowId  string
	childFuture workflow.ChildWorkflowFuture
}

func (c *customerSubscriptionsChildRun) ID() string {
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
	return c.childFuture.SignalChildWorkflow(ctx, customerSubscriptionsWorkflowCancelBillingName, req).Get(ctx, nil)
}

func (c *customerSubscriptionsChildRun) SetDiscount(ctx workflow.Context, req SetDiscountRequest) error {
	return c.childFuture.SignalChildWorkflow(ctx, customerSubscriptionsWorkflowSetDiscountName, req).Get(ctx, nil)
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
