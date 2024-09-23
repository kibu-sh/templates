package billingv1

import (
	"context"
	"errors"
	"go.temporal.io/sdk/workflow"
)

// service implements Service
type service struct {
	Workflows WorkflowsClient
}

// WatchAccount implements Service.WatchAccount
func (s *service) WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
	run, err := s.Workflows.CustomerSubscriptions().Execute(ctx, CustomerSubscriptionsRequest{})
	if err != nil {
		return
	}

	accountRes, err := run.GetAccountDetails(ctx, GetAccountDetailsRequest{})
	if err != nil {
		return
	}

	res.Status = accountRes.Status
	return
}

// activities implements Activities
type activities struct{}

// ChargePaymentMethod implements Activities.ChargePaymentMethod
func (a *activities) ChargePaymentMethod(ctx context.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error) {
	res.Success = !req.Fail

	if !res.Success {
		err = errors.New("payment failed")
	}
	return
}

// customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
type customerSubscriptionsWorkflow struct {
	activities    ActivitiesProxy
	accountStatus AccountStatus
	discountCode  string
}

// Execute implements CustomerSubscriptionsWorkflow.Execute
func (wf *customerSubscriptionsWorkflow) Execute(
	ctx workflow.Context,
	req CustomerSubscriptionsRequest,
) (res CustomerSubscriptionsResponse, err error) {
	// set initial account status
	wf.accountStatus = AccountStatusSubscribed

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			signal, _ := NewCancelBillingSignalChannel(ctx).Receive(ctx)
			_ = wf.CancelBilling(ctx, signal)
		}
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			NewCancelBillingSignalChannel(ctx).
				Select(workflow.NewSelector(ctx), nil).
				Select(ctx)
		}
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			signal, _ := NewSetDiscountSignalChannel(ctx).Receive(ctx)
			_ = wf.SetDiscount(ctx, signal)
		}
	})

	ctx.Done().Receive(ctx, nil)
	return
}

// GetAccountDetails implements CustomerSubscriptionsWorkflow.GetAccountDetails
func (wf *customerSubscriptionsWorkflow) GetAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResponse, err error) {
	res.Status = wf.accountStatus
	return
}

// AttemptPayment implements CustomerSubscriptionsWorkflow.AttemptPayment
func (wf *customerSubscriptionsWorkflow) AttemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error) {
	wf.accountStatus = AccountStatusPaymentPending

	_, err = wf.activities.ChargePaymentMethod(ctx, ChargePaymentMethodRequest{
		Fail: req.Fail,
	})

	if err != nil {
		wf.accountStatus = AccountStatusPaymentFailed
		return
	}

	wf.accountStatus = AccountStatusSubscribed
	return
}

// SetDiscount implements CustomerSubscriptionsWorkflow.SetDiscount
func (wf *customerSubscriptionsWorkflow) SetDiscount(ctx workflow.Context, req SetDiscountRequest) error {
	wf.discountCode = req.DiscountCode
	return nil
}

// CancelBilling implements CustomerSubscriptionsWorkflow.CancelBilling
func (wf *customerSubscriptionsWorkflow) CancelBilling(ctx workflow.Context, req CancelBillingRequest) (err error) {
	wf.accountStatus = AccountStatusCanceled
	return
}

// NewService creates an instance of Service
//
//kibu:provider
func NewService(workflows WorkflowsClient) Service {
	return &service{
		Workflows: workflows,
	}
}

// NewActivities creates an instance of Activities
//
//kibu:provider
func NewActivities() Activities {
	return &activities{}
}

// NewCustomerSubscriptionsWorkflow creates an instance of CustomerSubscriptionsWorkflow
//
//kibu:provider
func NewCustomerSubscriptionsWorkflow(activities ActivitiesProxy) CustomerSubscriptionsWorkflow {
	return &customerSubscriptionsWorkflow{
		activities: activities,
	}
}
