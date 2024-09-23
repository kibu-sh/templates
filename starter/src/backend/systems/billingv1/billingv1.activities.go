package billingv1

import (
	"context"
	"errors"
	"kibu.sh/starter/src/backend/database/models"
)

// activities implements Activities
type activities struct {
	TxnProvider models.TxnProvider
}

// ChargePaymentMethod implements Activities.ChargePaymentMethod
func (a *activities) ChargePaymentMethod(ctx context.Context, req ChargePaymentMethodRequest) (res ChargePaymentMethodResponse, err error) {
	res.Success = !req.Fail

	if !res.Success {
		err = errors.New("payment failed")
	}
	return
}

// NewActivities creates an instance of Activities
//
//kibu:provider
func NewActivities(provider models.TxnProvider) Activities {
	return &activities{
		TxnProvider: provider,
	}
}
