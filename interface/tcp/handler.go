package tcp

import (
	"context"
	"net"
)

// Handler defines the methods that any handler should implement tcp server
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
