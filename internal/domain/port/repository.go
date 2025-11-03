package port

import (
	"context"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

type BackendRepository interface {
	GetAll(ctx context.Context) []*model.Backend

	GetHealthy(ctx context.Context) []*model.Backend

	Add(ctx context.Context, backend *model.Backend) error

	Remove(ctx context.Context, id string) error

	Update(ctx context.Context, backend *model.Backend) error
}
