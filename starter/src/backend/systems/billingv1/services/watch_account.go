package services

import (
	"context"
	"kibu.sh/starter/src/backend/systems/billingv1"
)

var _ billingv1.Service = (*service)(nil)

// service implements billingv1.Service
type service struct {
	Workflows billingv1.WorkflowsClient
}

func (s *service) WatchAccount(ctx context.Context, req billingv1.WatchAccountRequest) (res billingv1.WatchAccountResponse, err error) {
	run, err := s.Workflows.CustomerSubscriptionsWorkflow().Execute(ctx, billingv1.CustomerSubscriptionsRequest{})
	if err != nil {
		return
	}

	accountRes, err := run.GetAccountDetails(ctx, billingv1.GetAccountDetailsRequest{})
	if err != nil {
		return
	}

	res.Status = accountRes.Status
	return
}

// NewService creates an instance of Service
//
//kibu:provider
func NewService(workflows billingv1.WorkflowsClient) billingv1.Service {
	return &service{
		Workflows: workflows,
	}
}
