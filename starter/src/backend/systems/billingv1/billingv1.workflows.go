package billingv1

import (
	"go.temporal.io/sdk/workflow"
)

// customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
type customerSubscriptionsWorkflow struct {
	Activities    ActivitiesProxy
	input         *customerSubscriptionsWorkflowInput
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
	return func(input *customerSubscriptionsWorkflowInput) (CustomerSubscriptionsWorkflow, error) {
		return &customerSubscriptionsWorkflow{
			Activities: activities,
		}, nil
	}
}
