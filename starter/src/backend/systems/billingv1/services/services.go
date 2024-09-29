package services

import (
	"context"
	. "kibu.sh/starter/src/backend/systems/billingv1"
)

// service provides a contract for the billing system
//
//kibu:service
type service struct {
	Workflows WorkflowsClient
}

// WatchAccount establishes a connection and receives messages for account changes
//
//kibu:endpoint
func (s *service) WatchAccount(ctx context.Context, req WatchAccountRequest) (res WatchAccountResponse, err error) {
	run, err := s.Workflows.CustomerSubscriptions().Execute(ctx, CustomerSubscriptionsRequest{})
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
