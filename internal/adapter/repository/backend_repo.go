package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

type BackendRepo struct {
	backends map[string]*model.Backend
	mu       sync.RWMutex
}

func New() *BackendRepo {
	return &BackendRepo{
		backends: make(map[string]*model.Backend),
	}
}

func (r *BackendRepo) GetAll(ctx context.Context) []*model.Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backends := make([]*model.Backend, 0, len(r.backends))
	for _, backend := range r.backends {
		backends = append(backends, backend)
	}
	return backends
}

func (r *BackendRepo) GetHealthy(ctx context.Context) []*model.Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backends := make([]*model.Backend, 0)
	for _, backend := range r.backends {
		if backend.GetHealthy() {
			backends = append(backends, backend)
		}
	}
	return backends
}

func (r *BackendRepo) Add(ctx context.Context, backend *model.Backend) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backends[backend.ID]; exists {
		return fmt.Errorf("backend with ID %s already exists", backend.ID)
	}

	r.backends[backend.ID] = backend
	return nil
}

func (r *BackendRepo) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backends[id]; !exists {
		return fmt.Errorf("backend with ID %s not found", id)
	}

	delete(r.backends, id)
	return nil
}

func (r *BackendRepo) Update(ctx context.Context, backend *model.Backend) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backends[backend.ID]; !exists {
		return fmt.Errorf("backend with ID %s not found", backend.ID)
	}

	r.backends[backend.ID] = backend
	return nil
}
