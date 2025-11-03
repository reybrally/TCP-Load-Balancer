package usecase

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type HandleConnectionUseCase struct {
	balancer   port.LoadBalancer
	repository port.BackendRepository
	metrics    port.MetricsCollector
	logger     *logger.Logger
}

func New(balancer port.LoadBalancer, repository port.BackendRepository, metrics port.MetricsCollector, logger *logger.Logger) *HandleConnectionUseCase {
	return &HandleConnectionUseCase{
		balancer:   balancer,
		repository: repository,
		metrics:    metrics,
		logger:     logger,
	}
}

func (hc *HandleConnectionUseCase) Handle(ctx context.Context, clientConn net.Conn) error {
	defer clientConn.Close()

	startTime := time.Now()

	hc.logger.Debugf("New connection from %s", clientConn.RemoteAddr().String())

	healthyBackends := hc.repository.GetHealthy(ctx)
	if len(healthyBackends) == 0 {
		hc.logger.Warnf("No healthy backends available for client %s", clientConn.RemoteAddr().String())
		hc.metrics.IncConnectionErrors("all", "no_healthy_backends")
		clientConn.Write([]byte("No backends available\n"))
		return nil
	}

	backend, err := hc.balancer.SelectBackend(healthyBackends)
	if err != nil {
		hc.logger.Errorf("Failed to select backend: %v", err)
		hc.metrics.IncConnectionErrors("all", "backend_selection_failed")
		return err
	}

	backendAddr := backend.GetAddress()

	hc.logger.Debugf("Routing connection from %s to backend %s", clientConn.RemoteAddr().String(), backendAddr)

	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		hc.logger.Errorf("Failed to connect to backend %s: %v", backendAddr, err)
		hc.metrics.IncConnectionErrors(backendAddr, "connection_failed")
		clientConn.Write([]byte("Backend unavailable\n"))
		return err
	}
	defer backendConn.Close()

	hc.metrics.IncConnectionsTotal(backendAddr)
	hc.metrics.IncConnectionsActive(backendAddr)
	defer hc.metrics.DecConnectionsActive(backendAddr)

	backend.IncreaseConnections()
	defer backend.DecreaseConnections()

	err = hc.proxyConnections(clientConn, backendConn)

	duration := time.Since(startTime).Seconds()
	hc.metrics.ObserveConnectionDuration(backendAddr, duration)

	return err
}

func (hc *HandleConnectionUseCase) proxyConnections(clientConn, backendConn net.Conn) error {
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(backendConn, clientConn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		errChan <- err
	}()

	<-errChan

	return nil
}
