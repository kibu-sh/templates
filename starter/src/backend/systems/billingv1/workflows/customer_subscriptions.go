package workflows

import (
	"go.temporal.io/sdk/workflow"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

// ensure that customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
var _ billingv1.CustomerSubscriptionsWorkflow = (*customerSubscriptionsWorkflow)(nil)

// customerSubscriptionsWorkflow implements CustomerSubscriptionsWorkflow
type customerSubscriptionsWorkflow struct {
	Activities    billingv1.ActivitiesProxy
	accountStatus billingv1.AccountStatus
	discountCode  string
	input         *billingv1.CustomerSubscriptionsWorkflowInput
}

// Execute implements CustomerSubscriptionsWorkflow.Execute
func (wf *customerSubscriptionsWorkflow) Execute(
	ctx workflow.Context,
	req billingv1.CustomerSubscriptionsRequest,
) (res billingv1.CustomerSubscriptionsResponse, err error) {
	// set initial account status
	wf.accountStatus = billingv1.AccountStatusSubscribed

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
			signal, _ := billingv1.NewSetDiscountSignalChannel(ctx).Receive(ctx)
			_ = wf.SetDiscount(ctx, signal)
		}
	})

	ctx.Done().Receive(ctx, nil)
	return
}

// GetAccountDetails implements CustomerSubscriptionsWorkflow.GetAccountDetails
func (wf *customerSubscriptionsWorkflow) GetAccountDetails(req billingv1.GetAccountDetailsRequest) (res billingv1.GetAccountDetailsResponse, err error) {
	res.Status = wf.accountStatus
	res.DiscountCode = wf.discountCode
	return
}

// AttemptPayment implements CustomerSubscriptionsWorkflow.AttemptPayment
func (wf *customerSubscriptionsWorkflow) AttemptPayment(ctx workflow.Context, req billingv1.AttemptPaymentRequest) (res billingv1.AttemptPaymentResponse, err error) {
	wf.accountStatus = billingv1.AccountStatusPaymentPending

	_, err = wf.Activities.ChargePaymentMethod(ctx, billingv1.ChargePaymentMethodRequest{
		Fail: req.Fail,
	})

	if err != nil {
		wf.accountStatus = billingv1.AccountStatusPaymentFailed
		return
	}

	wf.accountStatus = billingv1.AccountStatusSubscribed
	return
}

// SetDiscount implements CustomerSubscriptionsWorkflow.SetDiscount
func (wf *customerSubscriptionsWorkflow) SetDiscount(ctx workflow.Context, req billingv1.SetDiscountRequest) error {
	wf.discountCode = req.DiscountCode
	return nil
}

// CancelBilling implements CustomerSubscriptionsWorkflow.CancelBilling
func (wf *customerSubscriptionsWorkflow) CancelBilling(ctx workflow.Context, req billingv1.CancelBillingRequest) (err error) {
	wf.accountStatus = billingv1.AccountStatusCanceled
	return
}

// NewCustomerSubscriptionsWorkflowFactory returns a factory function for CustomerSubscriptionsWorkflow
//
//kibu:provider
func NewCustomerSubscriptionsWorkflowFactory(activities billingv1.ActivitiesProxy) billingv1.CustomerSubscriptionsWorkflowFactory {
	return func(input *billingv1.CustomerSubscriptionsWorkflowInput) (billingv1.CustomerSubscriptionsWorkflow, error) {
		return &customerSubscriptionsWorkflow{
			Activities: activities,
			input:      input,
		}, nil
	}
}
