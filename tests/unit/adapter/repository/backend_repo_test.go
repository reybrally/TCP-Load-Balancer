package model

import (
	"testing"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

func TestBackendCreation(t *testing.T) {
	backend := model.NewBackend("test-id", "192.168.1.1", 8080, 2)

	if backend.GetID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", backend.GetID())
	}

	if backend.GetAddress() != "192.168.1.1:8080" {
		t.Errorf("Expected address '192.168.1.1:8080', got %s", backend.GetAddress())
	}

	if backend.Weight != 2 {
		t.Errorf("Expected weight 2, got %d", backend.Weight)
	}
}

func TestBackendHealthStatus(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	if !backend.GetHealthy() {
		t.Error("Expected backend to be healthy by default")
	}

	backend.SetHealthy(false)
	if backend.GetHealthy() {
		t.Error("Expected backend to be unhealthy after SetHealthy(false)")
	}

	backend.SetHealthy(true)
	if !backend.GetHealthy() {
		t.Error("Expected backend to be healthy after SetHealthy(true)")
	}
}

func TestConnectionCount(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 active connections, got %d", backend.GetActiveConnections())
	}

	backend.IncreaseConnections()
	if backend.GetActiveConnections() != 1 {
		t.Errorf("Expected 1 active connection, got %d", backend.GetActiveConnections())
	}

	backend.IncreaseConnections()
	if backend.GetActiveConnections() != 2 {
		t.Errorf("Expected 2 active connections, got %d", backend.GetActiveConnections())
	}

	backend.DecreaseConnections()
	if backend.GetActiveConnections() != 1 {
		t.Errorf("Expected 1 active connection, got %d", backend.GetActiveConnections())
	}
}
