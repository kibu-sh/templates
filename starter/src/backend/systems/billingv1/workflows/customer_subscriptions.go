package workflows

import (
	"go.temporal.io/sdk/workflow"
	. "kibu.sh/starter/src/backend/systems/billingv1"
)

// ensure that customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
var _ CustomerSubscriptionsWorkflow = (*customerSubscriptionsWorkflow)(nil)

// customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
type customerSubscriptionsWorkflow struct {
	Activities    ActivitiesProxy
	accountStatus AccountStatus
	discountCode  string
	input         *CustomerSubscriptionsWorkflowInput
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
			signal, _ := wf.input.CancelBillingChannel.Receive(ctx)
			_ = wf.CancelBilling(ctx, signal)
		}
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			wf.input.CancelBillingChannel.
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

	_, err = wf.Activities.ChargePaymentMethod(ctx, ChargePaymentMethodRequest{
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

// NewCustomerSubscriptionsWorkflowFactory returns a factory function for CustomerSubscriptionsWorkflow
//
//kibu:provider
func NewCustomerSubscriptionsWorkflowFactory(activities ActivitiesProxy) CustomerSubscriptionsWorkflowFactory {
	return func(_ *CustomerSubscriptionsWorkflowInput) (CustomerSubscriptionsWorkflow, error) {
		return &customerSubscriptionsWorkflow{
			Activities: activities,
		}, nil
	}
}
