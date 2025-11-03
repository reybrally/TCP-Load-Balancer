package port

import (
	"context"
	"net"
)

type ConnectionHandler interface {
	Handle(ctx context.Context, conn net.Conn) error
}

type TCPListener interface {
	Listen(ctx context.Context, handler ConnectionHandler) error

	Close() error
}
