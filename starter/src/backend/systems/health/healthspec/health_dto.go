package healthspec

import "context"

type IndexRequest struct{}
type IndexResponse struct {
	Ok bool `json:"ok"`
}

type Service interface {
	Index(ctx context.Context, req IndexRequest) (IndexResponse, error)
}
