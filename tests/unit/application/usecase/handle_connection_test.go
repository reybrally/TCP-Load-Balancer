package usecase

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/balancer"
	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/repository"
	"github.com/reybrally/TCP-Load-Balancer/internal/application/usecase"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type mockMetricsCollector struct{}

func (m *mockMetricsCollector) IncConnectionsTotal(backend string)                  {}
func (m *mockMetricsCollector) IncConnectionsActive(backend string)                 {}
func (m *mockMetricsCollector) DecConnectionsActive(backend string)                 {}
func (m *mockMetricsCollector) IncConnectionErrors(backend, reason string)          {}
func (m *mockMetricsCollector) ObserveConnectionDuration(backend string, d float64) {}
func (m *mockMetricsCollector) IncHealthChecksTotal(backend, status string)         {}
func (m *mockMetricsCollector) SetBackendHealthStatus(backend string, healthy bool) {}

func TestHandleConnectionWithNoHealthyBackends(t *testing.T) {
	repo := repository.New()
	lb := balancer.New()
	metrics := &mockMetricsCollector{}
	log := logger.New("test")
	defer log.Sync()

	uc := usecase.New(lb, repo, metrics, log)

	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	backend := model.NewBackend("test", "localhost", 9999, 1)
	backend.SetHealthy(false)
	repo.Add(ctx, backend)

	done := make(chan error, 1)
	go func() {
		done <- uc.Handle(ctx, server)
	}()

	buf := make([]byte, 1024)
	client.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _ := client.Read(buf)
	response := string(buf[:n])

	if response != "No backends available\n" {
		t.Errorf("Expected 'No backends available' message, got: %s", response)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Logf("Handle returned error (expected): %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Handle connection timed out")
	}
}

func TestHandleConnectionWithHealthyBackends(t *testing.T) {
	repo := repository.New()
	// УБИРАЕМ: lb := balancer.New()
	log := logger.New("test")
	defer log.Sync()

	ctx := context.Background()

	// Add healthy backend
	backend := model.NewBackend("test", "localhost", 3001, 1)
	backend.SetHealthy(true)
	repo.Add(ctx, backend)

	healthyBackends := repo.GetHealthy(ctx)
	if len(healthyBackends) != 1 {
		t.Errorf("Expected 1 healthy backend, got %d", len(healthyBackends))
	}

	if !healthyBackends[0].GetHealthy() {
		t.Error("Expected backend to be healthy")
	}
}

func TestHandleConnectionBackendSelection(t *testing.T) {
	repo := repository.New()
	lb := balancer.New()
	log := logger.New("test")
	defer log.Sync()

	ctx := context.Background()

	backendConfigs := []struct {
		id   string
		port int
	}{
		{"b1", 3001},
		{"b2", 3002},
		{"b3", 3003},
	}

	for _, cfg := range backendConfigs {
		b := model.NewBackend(cfg.id, "localhost", cfg.port, 1)
		b.SetHealthy(true)
		repo.Add(ctx, b)
	}

	healthyBackends := repo.GetHealthy(ctx)
	if len(healthyBackends) != 3 {
		t.Fatalf("Expected 3 healthy backends, got %d", len(healthyBackends))
	}

	expected := []string{"b1", "b2", "b3", "b1"}
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

func TestConnectionCountTracking(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 initial connections, got %d", backend.GetActiveConnections())
	}

	backend.IncreaseConnections()
	if backend.GetActiveConnections() != 1 {
		t.Errorf("Expected 1 connection after increase, got %d", backend.GetActiveConnections())
	}

	backend.DecreaseConnections()
	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 connections after decrease, got %d", backend.GetActiveConnections())
	}
}

func TestRepositoryFiltersHealthyBackends(t *testing.T) {
	repo := repository.New()
	ctx := context.Background()

	backends := []*model.Backend{
		model.NewBackend("healthy-1", "localhost", 3001, 1),
		model.NewBackend("unhealthy-1", "localhost", 3002, 1),
		model.NewBackend("healthy-2", "localhost", 3003, 1),
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
		t.Errorf("Expected 2 healthy backends, got %d", len(healthyBackends))
	}

	for _, b := range healthyBackends {
		if !b.GetHealthy() {
			t.Errorf("Expected all backends in GetHealthy() to be healthy, got unhealthy: %s", b.GetID())
		}
	}
}
