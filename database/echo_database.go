package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

type EchoDatabase struct {
}

// NewEchoDatabase returns an new instance of EchoDatabase
func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

// Exec executes the echo command
func (e *EchoDatabase) Exec(_ resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e *EchoDatabase) Close() {
}

func (e *EchoDatabase) AfterClientClose(_ resp.Connection) {
}
