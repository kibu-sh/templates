package activities

import (
	"context"
	"errors"
	"kibu.sh/starter/src/backend/database/models"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

var _ billingv1.Activities = (*activities)(nil)

// activities implements billingv1.Activities
type activities struct {
	TxnProvider models.TxnProvider
}

// ChargePaymentMethod implements billingv1.Activities.ChargePaymentMethod
func (a *activities) ChargePaymentMethod(ctx context.Context, req billingv1.ChargePaymentMethodRequest) (res billingv1.ChargePaymentMethodResponse, err error) {
	res.Success = !req.Fail

	if !res.Success {
		err = errors.New("payment failed")
	}
	return
}

// NewActivities creates an instance of Activities
//
//kibu:provider
func NewActivities(provider models.TxnProvider) billingv1.Activities {
	return &activities{
		TxnProvider: provider,
	}
}
