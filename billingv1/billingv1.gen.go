package billingv1

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTaskQueueName                               = "billingv1"
	serviceWatchCustomerAccountName                    = "billingv1.WatchCustomerAccount"
	customerSubscriptionsWorkflowName                  = "billingv1.customerSubscriptions"
	customerSubscriptionsWorkflowAttemptPaymentName    = "billingv1.customerSubscriptions.AttemptPayment"
	customerSubscriptionsWorkflowGetAccountDetailsName = "billingv1.customerSubscriptions.GetAccountDetails"
	customerSubscriptionsWorkflowCancelBillingName     = "billingv1.customerSubscriptions.CancelBilling"
	customerSubscriptionsWorkflowSetDiscountName       = "billingv1.customerSubscriptions.SetDiscount"
	activitiesChargePaymentMethodName                  = "billingv1.activities.ChargePaymentMethod"
)

type WorkflowOptionFunc = temporal.WorkflowOptionFunc
type ActivityOptionFunc = temporal.ActivityOptionFunc
type UpdateOptionFunc = temporal.UpdateOptionFunc

type GetHandleOpts struct {
	WorkflowID string
	RunID      string
}

type UpdateHandle[T any] interface {
	UpdateID() string
	WorkflowID() string
	RunID() string
	Get(ctx context.Context) (T, error)
}

var _ UpdateHandle[any] = (*updateHandle[any])(nil)

type updateHandle[T any] struct {
	handle client.WorkflowUpdateHandle
}

func (u updateHandle[T]) UpdateID() string {
	return u.handle.UpdateID()
}

func (u updateHandle[T]) WorkflowID() string {
	return u.handle.WorkflowID()
}

func (u updateHandle[T]) RunID() string {
	return u.handle.RunID()
}

func (u updateHandle[T]) Get(ctx context.Context) (T, error) {
	var result T
	err := u.handle.Get(ctx, &result)
	return result, err
}

func NewUpdateHandle[T any](handle client.WorkflowUpdateHandle) UpdateHandle[T] {
	return &updateHandle[T]{handle}
}

type SignalHandle interface {
	Get(workflow.Context) error
	IsReady() bool
}

type signalHandle[T any] struct {
	handle workflow.Future
}

func (s signalHandle[T]) Get(ctx workflow.Context) error {
	return s.handle.Get(ctx, nil)
}

func (s signalHandle[T]) IsReady() bool {
	return s.handle.IsReady()
}

func NewSignalHandle[T any](handle workflow.Future) SignalHandle {
	return &signalHandle[T]{handle}
}

type AttemptPaymentFuture = UpdateHandle[AttemptPaymentResponse]
type SetDiscountFuture = SignalHandle

type CustomerBillingWorkflowRun interface {
	ID() string
	RunID() string
	Get(ctx context.Context) (CustomerBillingResponse, error)

	CancelBilling(ctx context.Context, req CancelBillingSignal) error
	SetDiscount(ctx context.Context, req SetDiscountSignal) error

	GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest) (GetAccountDetailsResult, error)

	AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (AttemptPaymentResponse, error)
	AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (AttemptPaymentFuture, error)
}

type CustomerBillingWorkflowClient interface {
	GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerBillingWorkflowRun, error)
	Execute(ctx context.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) (CustomerBillingWorkflowRun, error)
	ExecuteWithSetDiscount(ctx context.Context, req *CustomerBillingRequest, sig SetDiscountSignal, mods ...WorkflowOptionFunc) (CustomerBillingWorkflowRun, error)
}

type CustomerBillingWorkflowChildRun interface {
	ID() string
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
	Execute(ctx workflow.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) (CustomerBillingResponse, error)
	ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) CustomerBillingWorkflowChildRun
	External(ref GetHandleOpts) CustomerBillingExternalRun
}

type WorkflowsProxy interface {
	CustomerBilling() CustomerBillingWorkflowChildClient
}

type WorkflowsClient interface {
	CustomerBilling() CustomerBillingWorkflowClient
}

type workflowsClient struct {
	client client.Client
}

func (w *workflowsClient) CustomerBilling() CustomerBillingWorkflowClient {
	return &customerSubscriptionsClient{
		client: w.client,
	}
}

type workflowsProxy struct{}

func (w *workflowsProxy) CustomerBilling() CustomerBillingWorkflowChildClient {
	return &customerSubscriptionsChildClient{}
}

type customerSubscriptionsClient struct {
	client client.Client
}

func (c *customerSubscriptionsClient) Execute(ctx context.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) (CustomerBillingWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
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

func (c *customerSubscriptionsClient) GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerBillingWorkflowRun, error) {
	return &customerSubscriptionsRun{
		client:      c.client,
		workflowRun: c.client.GetWorkflow(ctx, ref.WorkflowID, ref.RunID),
	}, nil
}

func (c *customerSubscriptionsClient) ExecuteWithSetDiscount(ctx context.Context, req *CustomerBillingRequest, sig SetDiscountSignal, mods ...WorkflowOptionFunc) (CustomerBillingWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
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

func (c *customerSubscriptionsChildClient) Execute(ctx workflow.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) (CustomerBillingResponse, error) {
	return c.ExecuteAsync(ctx, req, mods...).Get(ctx)
}

func (c *customerSubscriptionsChildClient) ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest, mods ...WorkflowOptionFunc) CustomerBillingWorkflowChildRun {
	options := temporal.NewWorkflowOptionsBuilder().
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

func (c *customerSubscriptionsChildClient) External(ref GetHandleOpts) CustomerBillingExternalRun {
	return &customerBillingExternalRun{
		workflowID: ref.WorkflowID,
		runID:      ref.RunID,
	}
}

type customerBillingExternalRun struct {
	workflowID string
	runID      string
}

func (c *customerBillingExternalRun) ID() string {
	return c.workflowID
}

func (c *customerBillingExternalRun) RunID() string {
	return c.runID
}

func (c *customerBillingExternalRun) RequestCancellation(ctx workflow.Context) error {
	return workflow.RequestCancelExternalWorkflow(ctx, c.workflowID, c.runID).Get(ctx, nil)
}

func (c *customerBillingExternalRun) CancelBilling(ctx workflow.Context, req CancelBillingSignal) error {
	return c.CancelBillingAsync(ctx, req).Get(ctx, nil)
}

func (c *customerBillingExternalRun) CancelBillingAsync(ctx workflow.Context, req CancelBillingSignal) workflow.Future {
	return workflow.SignalExternalWorkflow(ctx, c.workflowID, c.runID, customerSubscriptionsWorkflowCancelBillingName, req)
}

func (c *customerBillingExternalRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	return c.SetDiscountAsync(ctx, req).Get(ctx, nil)
}

func (c *customerBillingExternalRun) SetDiscountAsync(ctx workflow.Context, req SetDiscountSignal) workflow.Future {
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

func (c *customerSubscriptionsRun) Get(ctx context.Context) (CustomerBillingResponse, error) {
	var result CustomerBillingResponse
	err := c.workflowRun.Get(ctx, &result)
	return result, err
}

func (c *customerSubscriptionsRun) AttemptPaymentAsync(ctx context.Context, req AttemptPaymentRequest, mods ...UpdateOptionFunc) (AttemptPaymentFuture, error) {
	options := temporal.NewUpdateOptionsBuilder().
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

func (c *customerSubscriptionsRun) CancelBilling(ctx context.Context, req CancelBillingSignal) error {
	return c.client.SignalWorkflow(ctx, c.ID(), c.RunID(), customerSubscriptionsWorkflowCancelBillingName, req)
}

func (c *customerSubscriptionsRun) SetDiscount(ctx context.Context, req SetDiscountSignal) error {
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

func (c *customerSubscriptionsChildRun) Get(ctx workflow.Context) (CustomerBillingResponse, error) {
	var result CustomerBillingResponse
	err := c.childFuture.Get(ctx, &result)
	return result, err
}

func (c *customerSubscriptionsChildRun) CancelBilling(ctx workflow.Context, req CancelBillingSignal) error {
	return c.childFuture.SignalChildWorkflow(ctx, customerSubscriptionsWorkflowCancelBillingName, req).Get(ctx, nil)
}

func (c *customerSubscriptionsChildRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	return c.childFuture.SignalChildWorkflow(ctx, customerSubscriptionsWorkflowSetDiscountName, req).Get(ctx, nil)
}
