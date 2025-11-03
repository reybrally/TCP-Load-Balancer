package usecase

import (
	"context"
	"io"
	"net"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type HandleConnectionUseCase struct {
	balancer   port.LoadBalancer
	repository port.BackendRepository
	logger     *logger.Logger
}

func New(balancer port.LoadBalancer, repository port.BackendRepository, logger *logger.Logger) *HandleConnectionUseCase {
	return &HandleConnectionUseCase{
		balancer:   balancer,
		repository: repository,
		logger:     logger,
	}
}

func (hc *HandleConnectionUseCase) Handle(ctx context.Context, clientConn net.Conn) error {
	defer clientConn.Close()

	hc.logger.Debugf("New connection from %s", clientConn.RemoteAddr().String())

	healthyBackends := hc.repository.GetHealthy(ctx)
	if len(healthyBackends) == 0 {
		hc.logger.Warnf("No healthy backends available for client %s", clientConn.RemoteAddr().String())
		clientConn.Write([]byte("No backends available\n"))
		return nil
	}

	backend, err := hc.balancer.SelectBackend(healthyBackends)
	if err != nil {
		hc.logger.Errorf("Failed to select backend: %v", err)
		return err
	}

	hc.logger.Debugf("Routing connection from %s to backend %s", clientConn.RemoteAddr().String(), backend.GetAddress())

	backendConn, err := net.Dial("tcp", backend.GetAddress())
	if err != nil {
		hc.logger.Errorf("Failed to connect to backend %s: %v", backend.GetAddress(), err)
		clientConn.Write([]byte("Backend unavailable\n"))
		return err
	}
	defer backendConn.Close()

	backend.IncreaseConnections()
	defer backend.DecreaseConnections()

	return hc.proxyConnections(clientConn, backendConn)
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
