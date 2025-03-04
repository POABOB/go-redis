package database

import (
	"go-redis/config"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func init() {
	RegisterSysCommand("AUTH", execAuth, -2)
}

// Auth validate client's password
func execAuth(c *resp.Connection, args database.CommandLine) resp.Reply {
	if len(args) != 1 {
		return reply.MakeStandardErrorReply("ERR wrong number of arguments for 'auth' command")
	}
	if config.Properties.RequirePass == "" {
		return reply.MakeStandardErrorReply("ERR Client sent AUTH, but no password is set")
	}
	password := string(args[0])
	(*c).SetPassword(password)
	if config.Properties.RequirePass != password {
		return reply.MakeStandardErrorReply("ERR invalid password")
	}
	return &reply.OkReply{}
}

func isAuthenticated(c resp.Connection) bool {
	if config.Properties.RequirePass == "" {
		return true
	}
	return c.GetPassword() == config.Properties.RequirePass
}
