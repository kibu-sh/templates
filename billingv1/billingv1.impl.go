package billingv1

import (
	"errors"
	"go.temporal.io/sdk/workflow"
)

// come back to this line
////go:generate go run github.com/kibu-sh/kibu/cmd/kibu@latest build ./

// activities synchronize the workflow state with an external payment gateway
//
//kibu:worker:activity task_queue=billingv1
type activities struct{}

// ChargePaymentMethod performs work against another transactional system
//
//kibu:activity
func (a *activities) ChargePaymentMethod(ctx workflow.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error) {
	res.Success = !req.Fail

	if !res.Success {
		err = errors.New("payment failed")
	}
	return
}

// customerSubscriptionsWorkflow represents a single long-running workflow for a customer
//
//kibu:worker:workflow task_queue=billingv1
type customerSubscriptionsWorkflow struct {
	accountStatus AccountStatus
	discountCode  string
}

// Execute initiates a long-running workflow for the customers account
//
//kibu:workflow:execute
func (wf *customerSubscriptionsWorkflow) Execute(ctx workflow.Context, req CustomerBillingRequest) (res CustomerBillingResponse, err error) {
	// set initial account status
	wf.accountStatus = AccountStatusSubscribed

	if err = wf.registerUpdateProgress(ctx); err != nil {
		return
	}

	if err = wf.registerGetProgressHandler(ctx); err != nil {
		return
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			channel := workflow.GetSignalChannelWithOptions(ctx,
				barv1CustomerBillingWorkflowCancelBillingName,
				workflow.SignalChannelOptions{
					// TODO: get from struct comment
					Description: "Sets the progress of the billing process",
				})

			var signal CancelBillingSignal
			channel.Receive(ctx, &signal)
		}
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			channel := workflow.GetSignalChannelWithOptions(ctx,
				barv1CustomerBillingWorkflowSetDiscountName,
				workflow.SignalChannelOptions{
					// TODO: get from struct comment
					Description: "Sets the progress of the billing process",
				})

			var signal SetDiscountSignal
			channel.Receive(ctx, &signal)
			_ = wf.SetDiscount(signal)
		}
	})

	ctx.Done().Receive(ctx, nil)
	return
}

// GetAccountDetails returns the current account status
// should not mutate state, doesn't have context
// should not call activities (helps prevent accidental activity calls)
//
//kibu:workflow:query
func (wf *customerSubscriptionsWorkflow) GetAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResult, err error) {
	res.Status = wf.accountStatus
	return
}

// CancelBilling sends a signal to cancel the customer's billing process
// this will end the workflow
//
//kibu:workflow:signal
func (wf *customerSubscriptionsWorkflow) CancelBilling(ctx workflow.Context, req CancelBillingRequest) (err error) {
	wf.accountStatus = AccountStatusCanceled
	return
}

// AttemptPayment attempts to charge the customers payment method
// the account status will reflect the outcome of the attempt
//
//kibu:workflow:update
func (wf *customerSubscriptionsWorkflow) AttemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error) {
	wf.accountStatus = AccountStatusPaymentPending
	// TODO: process transaction here
	return
}

func (wf *customerSubscriptionsWorkflow) SetDiscount(req SetDiscountSignal) error {
	wf.discountCode = req.DiscountCode
	return nil
}
