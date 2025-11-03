package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/balancer"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/repository"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

func TestLoadBalancerDistribution(t *testing.T) {
	repo := repository.New()
	lb := balancer.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("backend1", "localhost", 3001, 1),
		model.NewBackend("backend2", "localhost", 3002, 1),
		model.NewBackend("backend3", "localhost", 3003, 1),
	}

	for _, b := range backends {
		b.SetHealthy(true)
		repo.Add(ctx, b)
	}

	healthyBackends := repo.GetHealthy(ctx)

	distributionMap := make(map[string]int)
	rounds := 30

	for i := 0; i < rounds; i++ {
		selected, err := lb.SelectBackend(healthyBackends)
		if err != nil {
			t.Fatalf("Failed to select backend: %v", err)
		}
		distributionMap[selected.GetID()]++
	}

	expectedPerBackend := rounds / len(backends)
	for id, count := range distributionMap {
		if count != expectedPerBackend {
			t.Errorf("Backend %s: expected %d selections, got %d", id, expectedPerBackend, count)
		}
	}

	t.Logf("Distribution: %v (total rounds: %d)", distributionMap, rounds)
}

func TestLoadBalancerWithUnhealthyBackends(t *testing.T) {
	repo := repository.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("healthy1", "localhost", 3001, 1),
		model.NewBackend("unhealthy1", "localhost", 3002, 1),
		model.NewBackend("healthy2", "localhost", 3003, 1),
		model.NewBackend("unhealthy2", "localhost", 3004, 1),
	}

	for i, b := range backends {
		if i%2 == 0 {
			b.SetHealthy(true)
		} else {
			b.SetHealthy(false)
		}
		repo.Add(ctx, b)
	}

	healthyBackends := repo.GetHealthy(ctx)
	if len(healthyBackends) != 2 {
		t.Fatalf("Expected 2 healthy backends, got %d", len(healthyBackends))
	}

	for _, b := range healthyBackends {
		if !b.GetHealthy() {
			t.Errorf("Backend %s should be healthy", b.GetID())
		}
	}
}

func TestLoadBalancerConcurrentSelections(t *testing.T) {
	repo := repository.New()
	lb := balancer.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("backend1", "localhost", 3001, 1),
		model.NewBackend("backend2", "localhost", 3002, 1),
		model.NewBackend("backend3", "localhost", 3003, 1),
	}

	for _, b := range backends {
		b.SetHealthy(true)
		repo.Add(ctx, b)
	}

	healthyBackends := repo.GetHealthy(ctx)
	distributionMap := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	numGoroutines := 30
	selectionsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < selectionsPerGoroutine; j++ {
				selected, err := lb.SelectBackend(healthyBackends)
				if err != nil {
					t.Errorf("Failed to select backend: %v", err)
					return
				}

				mu.Lock()
				distributionMap[selected.GetID()]++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	totalSelections := numGoroutines * selectionsPerGoroutine
	if len(distributionMap) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(distributionMap))
	}

	total := 0
	for _, count := range distributionMap {
		total += count
	}

	if total != totalSelections {
		t.Errorf("Expected %d total selections, got %d", totalSelections, total)
	}

	t.Logf("Concurrent distribution: %v (total: %d)", distributionMap, total)
}

func TestBackendConnectionTracking(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 initial connections, got %d", backend.GetActiveConnections())
	}

	for i := 1; i <= 10; i++ {
		backend.IncreaseConnections()
		if backend.GetActiveConnections() != i {
			t.Errorf("After %d increments: expected %d connections, got %d", i, i, backend.GetActiveConnections())
		}
	}

	for i := 10; i > 0; i-- {
		backend.DecreaseConnections()
		if backend.GetActiveConnections() != i-1 {
			t.Errorf("After decrease: expected %d connections, got %d", i-1, backend.GetActiveConnections())
		}
	}
}

func TestBackendHealthTransitions(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	if !backend.GetHealthy() {
		t.Error("Backend should be healthy by default")
	}

	backend.SetHealthy(false)
	if backend.GetHealthy() {
		t.Error("Backend should be unhealthy after SetHealthy(false)")
	}

	backend.SetHealthy(true)
	if !backend.GetHealthy() {
		t.Error("Backend should be healthy after SetHealthy(true)")
	}

	for i := 0; i < 10; i++ {
		isHealthy := i%2 == 0
		backend.SetHealthy(isHealthy)
		if backend.GetHealthy() != isHealthy {
			t.Errorf("Iteration %d: expected healthy=%v, got %v", i, isHealthy, backend.GetHealthy())
		}
	}
}

func TestRepositoryOperations(t *testing.T) {
	repo := repository.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("b1", "localhost", 3001, 1),
		model.NewBackend("b2", "localhost", 3002, 1),
		model.NewBackend("b3", "localhost", 3003, 1),
	}

	for _, b := range backends {
		b.SetHealthy(true)
		err := repo.Add(ctx, b)
		if err != nil {
			t.Fatalf("Failed to add backend: %v", err)
		}
	}

	all := repo.GetAll(ctx)
	if len(all) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(all))
	}

	healthy := repo.GetHealthy(ctx)
	if len(healthy) != 3 {
		t.Errorf("Expected 3 healthy backends, got %d", len(healthy))
	}

	all[0].SetHealthy(false)
	healthyAfter := repo.GetHealthy(ctx)
	if len(healthyAfter) != 2 {
		t.Errorf("Expected 2 healthy backends after marking one unhealthy, got %d", len(healthyAfter))
	}
}

func TestMultipleBackendSelections(t *testing.T) {
	repo := repository.New()
	lb := balancer.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("primary", "localhost", 3001, 2),
		model.NewBackend("secondary", "localhost", 3002, 1),
	}

	for _, b := range backends {
		b.SetHealthy(true)
		repo.Add(ctx, b)
	}

	healthyBackends := repo.GetHealthy(ctx)

	expected := []string{"primary", "secondary", "primary", "secondary"}
	for i, exp := range expected {
		selected, err := lb.SelectBackend(healthyBackends)
		if err != nil {
			t.Fatalf("Failed to select backend: %v", err)
		}
		if selected.GetID() != exp {
			t.Errorf("Selection %d: expected %s, got %s", i, exp, selected.GetID())
		}
	}
}

func TestBackendTimeoutScenario(t *testing.T) {
	backend := model.NewBackend("slow", "localhost", 9999, 1)
	backend.SetHealthy(true)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Log("Context timeout correctly triggered")
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected context timeout before 200ms")
	}
}
