package foov1

import (
	billingv1 "barv1"
	"context"
	"go.temporal.io/sdk/workflow"
)

type activities struct {
	Billingv1 billingv1.WorkflowsClient
}

type workflows struct {
	Billingv1 billingv1.WorkflowsProxy
}

func (a *activities) TryFromActivity(ctx context.Context) {
	handle, err := a.Billingv1.CustomerBilling().GetHandle(ctx, billingv1.GetHandleOpts{
		WorkflowID: "",
		RunID:      "",
	})

	_, err = handle.GetAccountDetails(ctx, billingv1.GetAccountDetailsRequest{})
	if err != nil {
		return
	}

	err = handle.SetDiscount(ctx, billingv1.SetDiscountSignal{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	if err != nil {
		return
	}

	_, err = handle.AttemptPayment(ctx, billingv1.AttemptPaymentRequest{
		Fail: true,
	})

	err = handle.CancelBilling(ctx, billingv1.CancelBillingSignal{})
	if err != nil {
		return
	}

	return
}

func (wf *workflows) TryFromWorkflow(ctx workflow.Context) {
	handle := wf.Billingv1.
		CustomerBilling().
		ExecuteAsync(ctx, billingv1.CustomerBillingRequest{})

	err := handle.SetDiscount(ctx, billingv1.SetDiscountSignal{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	if err != nil {
		return
	}

	_, _ = handle.Get(ctx)

	externalHandle := wf.Billingv1.CustomerBilling().External(billingv1.GetHandleOpts{
		WorkflowID: "external-id",
		RunID:      "maybe-run-id",
	})

	_ = externalHandle.SetDiscount(ctx, billingv1.SetDiscountSignal{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	_ = externalHandle.RequestCancellation(ctx)
}
