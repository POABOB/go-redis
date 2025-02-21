package handler

import (
	"context"
	"errors"
	datebaseInterface "go-redis/interface/datebase"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

type RespHandler struct {
	activeConnections sync.Map
	database          datebaseInterface.Database
	closing           atomic.Boolean
}

// MakeHandler creates a new handler
func MakeHandler() *RespHandler {
	var database datebaseInterface.Database
	// TODO Implement database
	return &RespHandler{database: database}
}

// Handle creates a new connection with the client and serves it
func (handler *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	// in case the server is closed
	if handler.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConnection(conn)
	handler.activeConnections.Store(client, struct{}{})
	// prevent memory leak
	defer handler.activeConnections.Delete(client)

	ch := parser.ParseStream(conn)
	for payload := range ch {
		// Error
		if payload.Error != nil {
			// EOF or closed connection
			if errors.Is(payload.Error, io.EOF) || errors.Is(payload.Error, io.ErrUnexpectedEOF) ||
				strings.Contains(payload.Error.Error(), "use of closed network connection") {
				handler.closeOneClient(client)
				logger.Info("Connection closed: " + conn.RemoteAddr().String())
				return
			}
			// protocol error
			errorReply := reply.MakeStandardErrorReply(payload.Error.Error())
			err := client.Write(errorReply.ToBytes())
			if err != nil {
				handler.closeOneClient(client)
				logger.Info("Connection closed: " + conn.RemoteAddr().String())
				return
			}
			continue
		}

		// Exec
		if payload.Data == nil {
			continue
		}
		multiBulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}

		result := handler.database.Exec(client, multiBulkReply.Args)
		if result == nil {
			unknownErrorReply := reply.MakeUnknownErrorReply()
			_ = client.Write(unknownErrorReply.ToBytes())
			continue
		}
		_ = client.Write(result.ToBytes())
	}
}

// Close closes the server and all active connections
func (handler *RespHandler) Close() error {
	logger.Info("Closing server...")
	handler.closing.Set(true)
	handler.activeConnections.Range(func(key, value interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		// return true to continue the iteration, or it will stop at the first iteration
		return true
	})
	handler.database.Close()
	return nil
}

// closeOneClient closes one client
func (handler *RespHandler) closeOneClient(client *connection.Connection) {
	_ = client.Close()
	handler.database.AfterClientClose(client)
	handler.activeConnections.Delete(client)
}
