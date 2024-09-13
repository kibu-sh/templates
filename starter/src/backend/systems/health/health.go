package health

import (
	"context"
	"kibu.sh/starter/src/backend/systems/health/healthspec"
)

// compile time check to ensure Service implements healthspec.Service
var _ healthspec.Service = (*Service)(nil)

//kibu:service
type Service struct{}

//kibu:worker activity
type Activity struct{}

//kibu:endpoint path=/internal/health method=GET
func (s *Service) Index(ctx context.Context, req healthspec.IndexRequest) (res healthspec.IndexResponse, err error) {
	res.Ok = true
	return
}
