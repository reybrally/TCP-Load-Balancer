package balancer

import (
	"fmt"
	"sync"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
)

type RoundRobin struct {
	current int
	mu      sync.Mutex
}

func New() port.LoadBalancer {
	return &RoundRobin{
		current: 0,
	}
}

func (rr *RoundRobin) SelectBackend(backends []*model.Backend) (*model.Backend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}

	rr.mu.Lock()
	defer rr.mu.Unlock()

	backend := backends[rr.current%len(backends)]

	rr.current++

	return backend, nil
}
