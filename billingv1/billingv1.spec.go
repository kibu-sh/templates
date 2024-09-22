package billingv1

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

type ChargePaymentMethodRequest struct {
	Fail bool `json:"fail"`
}

type ChargePaymentMethodResponse struct {
	Success bool `json:"success"`
}

type CustomerSubscriptionsRequest struct{}
type CustomerSubscriptionsResponse struct{}

type SetDiscountRequest struct {
	DiscountCode string `json:"discount_code"`
}

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
