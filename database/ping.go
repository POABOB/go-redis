package database

import (
	databaseInterface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// init registers ping command.
func init() {
	RegisterCommand("ping", execPing, 1)
}

// Ping is used to reply ping commands.
// PING
func execPing(_ *DictEntity, _ databaseInterface.CommandLine) resp.Reply {
	return reply.MakePongReply()
}
