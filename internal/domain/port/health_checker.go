package port

import (
	"context"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

type HealthChecker interface {
	Check(ctx context.Context, backend *model.Backend) bool
}
