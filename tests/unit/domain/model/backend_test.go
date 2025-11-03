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
		t.Errorf("Expected 1 active connection after decrease, got %d", backend.GetActiveConnections())
	}

	backend.DecreaseConnections()
	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 active connections, got %d", backend.GetActiveConnections())
	}

	backend.DecreaseConnections()
	if backend.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 active connections (should not go negative), got %d", backend.GetActiveConnections())
	}
}

func TestConcurrentConnectionUpdates(t *testing.T) {
	backend := model.NewBackend("test", "localhost", 3001, 1)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			backend.IncreaseConnections()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if backend.GetActiveConnections() != 100 {
		t.Errorf("Expected 100 active connections, got %d", backend.GetActiveConnections())
	}
}

func TestBackendGetAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		port     int
		expected string
	}{
		{"Standard localhost", "localhost", 3001, "localhost:3001"},
		{"IP address", "192.168.1.1", 8080, "192.168.1.1:8080"},
		{"High port", "example.com", 65535, "example.com:65535"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := model.NewBackend("test", test.address, test.port, 1)
			if backend.GetAddress() != test.expected {
				t.Errorf("Expected address '%s', got '%s'", test.expected, backend.GetAddress())
			}
		})
	}
}
