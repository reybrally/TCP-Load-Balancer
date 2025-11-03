package balancer

import (
	"testing"

	"github.com/reybrally/TCP-Load-Balancer/internal/adapter/balancer"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
)

func TestRoundRobinSelection(t *testing.T) {
	rb := balancer.New()
	backends := []*model.Backend{
		model.NewBackend("b1", "localhost", 3001, 1),
		model.NewBackend("b2", "localhost", 3002, 1),
		model.NewBackend("b3", "localhost", 3003, 1),
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"First call", "b1"},
		{"Second call", "b2"},
		{"Third call", "b3"},
		{"Fourth call (cycle)", "b1"},
		{"Fifth call", "b2"},
	}

	for _, test := range tests {
		selected, err := rb.SelectBackend(backends)
		if err != nil || selected.GetID() != test.expected {
			t.Errorf("%s: expected %s, got %v, err: %v", test.name, test.expected, selected.GetID(), err)
		}
	}
}

func TestEmptyBackendsList(t *testing.T) {
	rb := balancer.New()
	_, err := rb.SelectBackend([]*model.Backend{})
	if err == nil {
		t.Error("Expected error for empty backends list, got nil")
	}
}

func TestSingleBackend(t *testing.T) {
	rb := balancer.New()
	backends := []*model.Backend{
		model.NewBackend("b1", "localhost", 3001, 1),
	}

	for i := 0; i < 5; i++ {
		selected, err := rb.SelectBackend(backends)
		if err != nil || selected.GetID() != "b1" {
			t.Errorf("Iteration %d: expected b1, got %v", i, selected.GetID())
		}
	}
}
