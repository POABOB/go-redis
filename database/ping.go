package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// Ping is used to reply ping command
func execPing(_ *Database, _ [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// RegisterPing registers ping command
func init() {
	RegisterCommand("ping", execPing, 1)
}
