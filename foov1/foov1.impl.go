package foov1

import (
	"context"
	"github.com/kibu-sh/kibu/pkg/transport/temporal"
	"go.temporal.io/sdk/workflow"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

type activities struct {
	Billingv1 billingv1.WorkflowsClient
}

type workflows struct {
	Billingv1 billingv1.WorkflowsProxy
}

func (a *activities) TryFromActivity(ctx context.Context) {
	handle, err := a.Billingv1.CustomerSubscriptions().GetHandle(ctx, temporal.GetHandleOpts{
		WorkflowID: "",
		RunID:      "",
	})
	if err != nil {
		return
	}

	a.Billingv1.
		CustomerSubscriptions().
		Execute(ctx, billingv1.billingv1spec{})

	_, err = handle.GetAccountDetails(ctx, billingv1.billingv1spec{})
	if err != nil {
		return
	}

	err = handle.SetDiscount(ctx, billingv1.billingv1spec{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	if err != nil {
		return
	}

	_, err = handle.AttemptPayment(ctx, billingv1.billingv1spec{
		Fail: true,
	})

	err = handle.CancelBilling(ctx, billingv1.billingv1spec{})
	if err != nil {
		return
	}

	return
}

func (wf *workflows) TryFromWorkflow(ctx workflow.Context) {
	handle := wf.Billingv1.
		CustomerSubscriptions().
		ExecuteAsync(ctx, billingv1.billingv1spec{})

	err := handle.SetDiscount(ctx, billingv1.billingv1spec{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	if err != nil {
		return
	}

	_, _ = handle.Get(ctx)

	externalHandle := wf.Billingv1.CustomerSubscriptions().External(temporal.GetHandleOpts{
		WorkflowID: "external-id",
		RunID:      "maybe-run-id",
	})

	_ = externalHandle.SetDiscount(ctx, billingv1.billingv1spec{
		DiscountCode: "temporal.replay.100.percent.off",
	})

	_ = externalHandle.RequestCancellation(ctx)
}
