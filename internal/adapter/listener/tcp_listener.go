package listener

import (
	"context"
	"fmt"
	"net"

	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type TCPListener struct {
	listener net.Listener
	logger   *logger.Logger
}

func New(host string, port int, logger *logger.Logger) (*TCPListener, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	return &TCPListener{
		listener: listener,
		logger:   logger,
	}, nil
}

func (tl *TCPListener) Listen(ctx context.Context, handler port.ConnectionHandler) error {
	tl.logger.Infof("Listening on %v", tl.listener.Addr())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := tl.listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				tl.logger.Warnf("Error accepting connection: %v", err)
				continue
			}

			go func() {
				if err := handler.Handle(ctx, conn); err != nil {
					tl.logger.Debugf("Error handling connection: %v", err)
				}
			}()
		}
	}
}

func (tl *TCPListener) Close() error {
	return tl.listener.Close()
}
