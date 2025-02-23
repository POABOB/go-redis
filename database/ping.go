package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// init registers ping command.
func init() {
	RegisterCommand("ping", execPing, 1)
}

// Ping is used to reply ping command.
// PING
func execPing(_ *Database, _ [][]byte) resp.Reply {
	return reply.MakePongReply()
}
