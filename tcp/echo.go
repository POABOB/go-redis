package tcp

import (
	"bufio"
	"context"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type EchoClient struct {
	Connection net.Conn
	Waiting    wait.Wait
}

// Close closes the connection while timeout
func (client *EchoClient) Close() error {
	client.Waiting.WaitWithTimeout(10 * 1000 * time.Millisecond)
	// no need to return error while the connection is closed
	_ = client.Connection.Close()
	return nil
}

type EchoHandler struct {
	activeConnections sync.Map
	closing           atomic.Boolean
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// in case the server is closed
	if handler.closing.Get() {
		_ = conn.Close()
	}
	client := &EchoClient{
		Connection: conn,
	}
	handler.activeConnections.Store(client, struct{}{})
	// prevent memory leak
	defer handler.activeConnections.Delete(client)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection closing...")
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(message)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("Closing server...")
	handler.closing.Set(true)
	handler.activeConnections.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		// return true to continue the iteration or it will stop at the first iteration
		return true
	})
	return nil
}
