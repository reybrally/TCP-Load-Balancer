package health

import (
	"context"
	"net"
	"time"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type TCPChecker struct {
	timeout time.Duration
	metrics port.MetricsCollector
	logger  *logger.Logger
}

func New(timeout time.Duration, metrics port.MetricsCollector, logger *logger.Logger) *TCPChecker {
	return &TCPChecker{
		timeout: timeout,
		metrics: metrics,
		logger:  logger,
	}
}

func (tc *TCPChecker) Check(ctx context.Context, backend *model.Backend) bool {
	if backend == nil {
		return false
	}

	backendAddr := backend.GetAddress()

	checkCtx, cancel := context.WithTimeout(ctx, tc.timeout)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(checkCtx, "tcp", backendAddr)
	if err != nil {
		tc.logger.Debugf("Health check failed for %s: %v", backendAddr, err)
		tc.metrics.IncHealthChecksTotal(backendAddr, "failed")
		tc.metrics.SetBackendHealthStatus(backendAddr, false)
		return false
	}
	defer conn.Close()

	tc.logger.Debugf("Health check passed for %s", backendAddr)
	tc.metrics.IncHealthChecksTotal(backendAddr, "success")
	tc.metrics.SetBackendHealthStatus(backendAddr, true)
	return true
}
