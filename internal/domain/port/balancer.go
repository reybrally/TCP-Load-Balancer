package port

import (
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

type LoadBalancer interface {
	SelectBackend(backends []*model.Backend) (*model.Backend, error)
}
