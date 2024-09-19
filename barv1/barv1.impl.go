package barv1

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

// check that these exist
// if not create them
type activities struct {
	//*MyActivityDeps
}

func (s *activities) RunTransaction(ctx context.Context, req Request) (res Response, err error) {
	return
}

type workflows struct{}

func (s *workflows) StartSomeLongProcess(ctx workflow.Context, req Request) (res Response, err error) {
	return
}

func myOverrides(builder OptionsBuilder[Request]) OptionsBuilder[Request] {

}
