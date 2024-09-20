package barv1

import (
	"go.temporal.io/sdk/workflow"
)

// customerBillingWorkflow represents a single long-running workflow for a customer
//
//kibu:workflow
type customerBillingWorkflow struct {
	accountStatus AccountStatus
}

// Execute initiates a long-running workflow for the customers account
//
//kibu:workflow:execute
func (wf *customerBillingWorkflow) Execute(ctx workflow.Context, req CustomerBillingRequest) (res CustomerBillingResponse, err error) {
	// set initial account status
	wf.accountStatus = AccountStatusSubscribed

	if err = wf.registerUpdateProgress(ctx); err != nil {
		return
	}

	if err = wf.registerGetProgressHandler(ctx); err != nil {
		return
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		channel := workflow.GetSignalChannelWithOptions(ctx,
			barv1CustomerBillingWorkflowSetProgressName,
			workflow.SignalChannelOptions{
				// TODO: get from struct comment
				Description: "Sets the progress of the billing process",
			})

		var signal SetProgressParams
		channel.Receive(ctx, &signal)
	})

	ctx.Done().Receive(ctx, nil)
	return
}

// GetAccountDetails returns the current account status
// should not mutate state, doesn't have context
// should not call activities (helps prevent accidental activity calls)
//
//kibu:workflow:query
func (wf *customerBillingWorkflow) GetAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResult, err error) {
	res.Status = wf.accountStatus
	return
}

// CancelBilling sends a signal to cancel the customer's billing process
// this will end the workflow
//
//kibu:workflow:signal
func (wf *customerBillingWorkflow) CancelBilling(ctx workflow.Context, req CancelBillingRequest) (err error) {
	wf.accountStatus = AccountStatusCanceled
	return
}

// AttemptPayment attempts to charge the customers payment method
// the account status will reflect the outcome of the attempt
//
//kibu:workflow:update
func (wf *customerBillingWorkflow) AttemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error) {
	wf.accountStatus = AccountStatusPaymentPending
	// TODO: process transaction here
	return
}
