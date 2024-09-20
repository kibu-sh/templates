package barv1

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

type AccountStatus string

const (
	AccountStatusUnsubscribed   AccountStatus = "trial"
	AccountStatusSubscribed     AccountStatus = "subscribed"
	AccountStatusPaymentFailed  AccountStatus = "payment_failed"
	AccountStatusPaymentPending AccountStatus = "payment_pending"
	AccountStatusCanceled       AccountStatus = "canceled"
)

type WatchAccountRequest struct{}
type WatchAccountResponse struct {
	Status AccountStatus
}

//kibu:service
type Service interface {
	// WatchAccount establishes a websocket connection to the billing service
	WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error)
}

type ChargePaymentMethodRequest struct {
	Fail bool `json:"fail"`
}

type ChargePaymentMethodResponse struct {
	Success bool `json:"success"`
}

// Activities attempt to synchronize the workflow state with an external payment gateway
//
//kibu:worker activity task_queue=custom_queue
type Activities interface {
	// ChargePaymentMethod performs work against another transactional system
	ChargePaymentMethod(ctx context.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error)
}

type CustomerBillingRequest struct{}
type CustomerBillingResponse struct{}
type CancelBillingRequest struct{}

type AttemptPaymentRequest struct {
	Fail bool `json:"fail"`
}

type AttemptPaymentResponse struct {
	Success bool `json:"success"`
}

type GetAccountDetailsRequest struct{}
type GetAccountDetailsResult struct {
	Status AccountStatus
}

// CustomerBillingWorkflow controls the customer's billing process
// can be inferred by the naming convention
//
//kibu:workflow task_queue=billing
type customer interface {
	// execute initiates a long-running workflow for the customers account
	// complains if interface doesn't implement execute
	//
	//kibu:workflow:execute
	execute(ctx workflow.Context, req CustomerBillingRequest) (res CustomerBillingResponse, err error)

	// AttemptPayment attempts to charge the customers payment method
	// the account status will reflect the outcome of the attempt
	//kibu:workflow:update
	attemptPayment(ctx workflow.Context, req AttemptPaymentRequest) (res AttemptPaymentResponse, err error)

	// GetAccountDetails returns the current account status
	// should not mutate state, doesn't have context
	// should not call activities (helps prevent accidental activity calls)
	//kibu:workflow:query
	getAccountDetails(req GetAccountDetailsRequest) (res GetAccountDetailsResult, err error)

	// CancelBilling sends a signal to cancel the customer's billing process
	// this will end the workflow
	//kibu:workflow:signal
	cancelBilling(ctx workflow.Context, req CancelBillingRequest) (err error)
}

// CustomerBillingWorkflow is a workflow that is referenced by the CustomerBilling workflow
// can be inferred by the naming convention
// if we don't recognize the prefix and there's no directive, get the line number and complain
