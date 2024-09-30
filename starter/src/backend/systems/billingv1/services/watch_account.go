package services

import (
	"context"
	. "kibu.sh/starter/src/backend/systems/billingv1"
)

var _ Service = (*service)(nil)

// service implements billingv1.Service
type service struct {
	Workflows WorkflowsClient
}

func (s *service) WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
	run, err := s.Workflows.CustomerSubscriptionsWorkflow().Execute(ctx, CustomerSubscriptionsRequest{})
	if err != nil {
		return
	}

	accountRes, err := run.GetAccountDetails(ctx, GetAccountDetailsRequest{})
	if err != nil {
		return
	}

	res.Status = accountRes.Status
	return
}

// NewService creates an instance of Service
//
//kibu:provider
func NewService(workflows WorkflowsClient) Service {
	return &service{
		Workflows: workflows,
	}
}
