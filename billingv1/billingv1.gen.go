package billingv1

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

// Constants for workflow and activity names.
const (
	barv1ServiceWatchBillingBillingName               = "barv1.WatchBillingBilling"
	barv1CustomerBillingWorkflowName                  = "barv1.customerBillingWorkflow"
	barv1CustomerBillingWorkflowAttemptPaymentName    = "barv1.customerBillingWorkflow.AttemptPayment"
	barv1CustomerBillingWorkflowGetAccountDetailsName = "barv1.customerBillingWorkflow.GetAccountDetails"
	barv1CustomerBillingWorkflowCancelBillingName     = "barv1.customerBillingWorkflow.CancelBilling"
	barv1CustomerBillingWorkflowSetDiscountName       = "barv1.customerBillingWorkflow.SetDiscount"
	barv1ActivitiesChargePaymentMethodName            = "barv1.activities.ChargePaymentMethod"
)

type WorkflowOptionFunc = temporal.WorkflowOptionFunc
type ActivityOptionFunc = temporal.ActivityOptionFunc
type UpdateOptionFunc = temporal.UpdateOptionFunc

// GetHandleOpts for getting workflow handles.
type GetHandleOpts struct {
	WorkflowID string
	RunID      string
}

// Interfaces and structs implementing them.

type CustomerBillingWorkflowRun interface {
	ID() string
	RunID() string

	Get(ctx context.Context) (CustomerBillingResponse, error)

	AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...WorkflowOptionFunc) (AttemptPaymentResponse, error)

	GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest, mods ...WorkflowOptionFunc) (GetAccountDetailsResult, error)

	CancelBilling(ctx context.Context, req CancelBillingSignal, mods ...WorkflowOptionFunc) error
	SetDiscount(ctx context.Context, req SetDiscountSignal, mods ...WorkflowOptionFunc) error
}

type CustomerBillingWorkflowClient interface {
	Execute(ctx context.Context, req CustomerBillingRequest) (CustomerBillingWorkflowRun, error)
	GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerBillingWorkflowRun, error)
	SignalWithStartSetDiscount(ctx context.Context, req SetDiscountSignal) (CustomerBillingWorkflowRun, error)
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
	Execute(ctx workflow.Context, req CustomerBillingRequest) (CustomerBillingResponse, error)
	ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest) CustomerBillingWorkflowChildRun
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
	return &customerBillingWorkflowClient{
		client: w.client,
	}
}

// Implementing WorkflowsProxy
type workflowsProxy struct{}

func (w *workflowsProxy) CustomerBilling() CustomerBillingWorkflowChildClient {
	return &customerBillingWorkflowChildClient{}
}

// Implementing CustomerBillingWorkflowClient
type customerBillingWorkflowClient struct {
	client client.Client
}

func (c *customerBillingWorkflowClient) Execute(ctx context.Context, req CustomerBillingRequest) (CustomerBillingWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
		WithTaskQueue("default").
		WithProvidersWhenSupported().
		AsStartOptions()

	we, err := c.client.ExecuteWorkflow(ctx, options, barv1CustomerBillingWorkflowName, req)
	if err != nil {
		return nil, err
	}

	return &customerBillingWorkflowRun{
		client:      c.client,
		workflowRun: we,
	}, nil
}

func (c *customerBillingWorkflowClient) GetHandle(ctx context.Context, ref GetHandleOpts) (CustomerBillingWorkflowRun, error) {
	we := c.client.GetWorkflow(ctx, ref.WorkflowID, ref.RunID)
	return &customerBillingWorkflowRun{
		client:      c.client,
		workflowRun: we,
	}, nil
}

func (c *customerBillingWorkflowClient) SignalWithStartSetDiscount(ctx context.Context, req SetDiscountSignal) (CustomerBillingWorkflowRun, error) {
	options := temporal.NewWorkflowOptionsBuilder().
		WithTaskQueue("default").
		WithProvidersWhenSupported().
		AsStartOptions()

	we, err := c.client.SignalWithStartWorkflow(ctx, options.ID, barv1CustomerBillingWorkflowSetDiscountName, req, options, barv1CustomerBillingWorkflowName, req)
	if err != nil {
		return nil, err
	}

	return &customerBillingWorkflowRun{
		client:      c.client,
		workflowRun: we,
	}, nil
}

// Implementing CustomerBillingWorkflowChildClient
type customerBillingWorkflowChildClient struct{}

func (c *customerBillingWorkflowChildClient) Execute(ctx workflow.Context, req CustomerBillingRequest) (CustomerBillingResponse, error) {
	future := c.ExecuteAsync(ctx, req)
	return future.Get(ctx)
}

func (c *customerBillingWorkflowChildClient) ExecuteAsync(ctx workflow.Context, req CustomerBillingRequest) CustomerBillingWorkflowChildRun {
	options := temporal.NewWorkflowOptionsBuilder().
		WithTaskQueue("default").
		WithProvidersWhenSupported().
		AsChildOptions()

	ctx = workflow.WithChildOptions(ctx, options)
	childFuture := workflow.ExecuteChildWorkflow(ctx, barv1CustomerBillingWorkflowName, req)

	return &customerBillingWorkflowChildRun{
		childFuture: childFuture,
	}
}

func (c *customerBillingWorkflowChildClient) External(ref GetHandleOpts) CustomerBillingExternalRun {
	return &customerBillingExternalRun{
		workflowID: ref.WorkflowID,
		runID:      ref.RunID,
	}
}

// Implementing CustomerBillingExternalRun
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
	return workflow.SignalExternalWorkflow(ctx, c.workflowID, c.runID, barv1CustomerBillingWorkflowCancelBillingName, req)
}

func (c *customerBillingExternalRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	return c.SetDiscountAsync(ctx, req).Get(ctx, nil)
}

func (c *customerBillingExternalRun) SetDiscountAsync(ctx workflow.Context, req SetDiscountSignal) workflow.Future {
	return workflow.SignalExternalWorkflow(ctx, c.workflowID, c.runID, barv1CustomerBillingWorkflowSetDiscountName, req)
}

// Implementing CustomerBillingWorkflowRun
type customerBillingWorkflowRun struct {
	client      client.Client
	workflowRun client.WorkflowRun
}

func (c *customerBillingWorkflowRun) ID() string {
	return c.workflowRun.GetID()
}

func (c *customerBillingWorkflowRun) RunID() string {
	return c.workflowRun.GetRunID()
}

func (c *customerBillingWorkflowRun) Get(ctx context.Context) (CustomerBillingResponse, error) {
	var result CustomerBillingResponse
	err := c.workflowRun.Get(ctx, &result)
	return result, err
}

func (c *customerBillingWorkflowRun) AttemptPayment(ctx context.Context, req AttemptPaymentRequest, mods ...WorkflowOptionFunc) (AttemptPaymentResponse, error) {
	options := temporal.NewUpdateOptionsBuilder().
		WithProvidersWhenSupported(req).
		WithWorkflowID(c.ID()).
		WithRunID(c.RunID()).
		Build()

	updateHandle, err := c.client.UpdateWorkflow(ctx, options)
	if err != nil {
		return AttemptPaymentResponse{}, err
	}

	var result AttemptPaymentResponse
	err = updateHandle.Get(ctx, &result)
	return result, err
}

func (c *customerBillingWorkflowRun) GetAccountDetails(ctx context.Context, req GetAccountDetailsRequest, mods ...WorkflowOptionFunc) (GetAccountDetailsResult, error) {
	queryResponse, err := c.client.QueryWorkflow(ctx, c.ID(), c.RunID(), barv1CustomerBillingWorkflowGetAccountDetailsName, req)
	if err != nil {
		return GetAccountDetailsResult{}, err
	}

	var result GetAccountDetailsResult
	err = queryResponse.Get(&result)
	return result, err
}

func (c *customerBillingWorkflowRun) CancelBilling(ctx context.Context, req CancelBillingSignal, mods ...WorkflowOptionFunc) error {
	return c.client.SignalWorkflow(ctx, c.ID(), c.RunID(), barv1CustomerBillingWorkflowCancelBillingName, req)
}

func (c *customerBillingWorkflowRun) SetDiscount(ctx context.Context, req SetDiscountSignal, mods ...WorkflowOptionFunc) error {
	return c.client.SignalWorkflow(ctx, c.ID(), c.RunID(), barv1CustomerBillingWorkflowSetDiscountName, req)
}

// Implementing CustomerBillingWorkflowChildRun
type customerBillingWorkflowChildRun struct {
	workflowId  string
	runId       string
	childFuture workflow.ChildWorkflowFuture
}

func (c *customerBillingWorkflowChildRun) ID() string {
	return c.workflowId
}

func (c *customerBillingWorkflowChildRun) IsReady() bool {
	return c.childFuture.IsReady()
}

func (c *customerBillingWorkflowChildRun) Underlying() workflow.ChildWorkflowFuture {
	return c.childFuture
}

func (c *customerBillingWorkflowChildRun) Get(ctx workflow.Context) (CustomerBillingResponse, error) {
	var result CustomerBillingResponse
	err := c.childFuture.Get(ctx, &result)
	return result, err
}

func (c *customerBillingWorkflowChildRun) CancelBilling(ctx workflow.Context, req CancelBillingSignal) error {
	return c.childFuture.SignalChildWorkflow(ctx, barv1CustomerBillingWorkflowCancelBillingName, req).Get(ctx, nil)
}

func (c *customerBillingWorkflowChildRun) SetDiscount(ctx workflow.Context, req SetDiscountSignal) error {
	return c.childFuture.SignalChildWorkflow(ctx, barv1CustomerBillingWorkflowSetDiscountName, req).Get(ctx, nil)
}
