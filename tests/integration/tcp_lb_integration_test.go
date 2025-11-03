package integration

import (
	"context"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/balancer"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/repository"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
	"testing"
)

func TestLoadBalancerSelectionOrder(t *testing.T) {
	repo := repository.New()
	log := logger.New("test")
	defer log.Sync()

	backends := []*model.Backend{
		model.NewBackend("b1", "localhost", 3001, 1),
		model.NewBackend("b2", "localhost", 3002, 1),
		model.NewBackend("b3", "localhost", 3003, 1),
	}

	for _, b := range backends {
		ctx := context.Background()
		repo.Add(ctx, b)
	}

	lb := balancer.New()
	ctx := context.Background()

	allBackends := repo.GetAll(ctx)

	expected := []string{"b1", "b2", "b3", "b1", "b2"}
	for i, exp := range expected {
		selected, err := lb.SelectBackend(allBackends)
		if err != nil {
			t.Fatalf("Failed to select backend: %v", err)
		}
		if selected.GetID() != exp {
			t.Errorf("Call %d: expected %s, got %s", i, exp, selected.GetID())
		}
	}
}

func TestRepositoryAddAndGet(t *testing.T) {
	repo := repository.New()
	ctx := context.Background()

	backend := model.NewBackend("test", "localhost", 8080, 1)
	err := repo.Add(ctx, backend)

	if err != nil {
		t.Fatalf("Failed to add backend: %v", err)
	}

	all := repo.GetAll(ctx)
	if len(all) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(all))
	}

	if all[0].GetID() != "test" {
		t.Errorf("Expected backend ID 'test', got %s", all[0].GetID())
	}
}
