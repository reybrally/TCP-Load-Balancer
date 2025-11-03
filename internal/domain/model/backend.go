package model

import (
	"fmt"
	"sync"
)

type Backend struct {
	ID                string
	Address           string
	Port              int
	Weight            int
	IsHealthy         bool
	ActiveConnections int
	mu                sync.RWMutex
}

func NewBackend(id, address string, port, weight int) *Backend {
	return &Backend{
		ID:        id,
		Address:   address,
		Port:      port,
		Weight:    weight,
		IsHealthy: true,
	}
}

func (b *Backend) GetAddress() string {
	return fmt.Sprintf("%s:%d", b.Address, b.Port)
}

func (b *Backend) IncreaseConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ActiveConnections++
}

func (b *Backend) DecreaseConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ActiveConnections > 0 {
		b.ActiveConnections--
	}
}

func (b *Backend) GetActiveConnections() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ActiveConnections
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.IsHealthy = healthy
}

func (b *Backend) GetHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.IsHealthy
}
