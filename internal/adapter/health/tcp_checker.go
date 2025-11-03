package health

import (
	"context"
	"net"
	"time"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/model"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type TCPChecker struct {
	timeout time.Duration
	logger  *logger.Logger
}

func New(timeout time.Duration, logger *logger.Logger) *TCPChecker {
	return &TCPChecker{
		timeout: timeout,
		logger:  logger,
	}
}

func (tc *TCPChecker) Check(ctx context.Context, backend *model.Backend) bool {
	if backend == nil {
		return false
	}

	checkCtx, cancel := context.WithTimeout(ctx, tc.timeout)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(checkCtx, "tcp", backend.GetAddress())
	if err != nil {
		tc.logger.Debugf("Health check failed for %s: %v", backend.GetAddress(), err)
		return false
	}
	defer conn.Close()

	tc.logger.Debugf("Health check passed for %s", backend.GetAddress())
	return true
}
