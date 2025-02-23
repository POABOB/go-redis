package database

import (
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// init registers all string commands.
func init() {
	RegisterCommand("GET", execGet, 2)
	RegisterCommand("SET", execSet, 3)
	RegisterCommand("SETNX", execSetNx, 3)
	RegisterCommand("GETSET", execGetSet, 3)
	RegisterCommand("STRLEN", execStrLen, 2)
}

// execGet executes the get command.
// GET key
func execGet(db *Database, args [][]byte) resp.Reply {
	entity, exists := db.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	// TODO if there are different types, entity.Data.([]byte) need to check
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// execSet executes the set command.
// SET key value
func execSet(db *Database, args [][]byte) resp.Reply {
	db.SetEntity(string(args[0]), &database.DataEntity{Data: args[1]})
	db.addAofFunc(utils.ToCommandLine3("SET", args...))
	return reply.MakeOkReply()
}

// execSetNx executes the setnx command.
// SETNX key value
func execSetNx(db *Database, args [][]byte) resp.Reply {
	defer db.addAofFunc(utils.ToCommandLine3("SETNX", args...))
	return reply.MakeIntReply(int64(db.SetEntityIfAbsent(string(args[0]), &database.DataEntity{Data: args[1]})))
}

// execGetSet executes the getset command.
// GETSET key value
func execGetSet(db *Database, args [][]byte) resp.Reply {
	oldEntity, exists := db.GetEntity(string(args[0]))
	db.SetEntity(string(args[0]), &database.DataEntity{Data: args[1]})
	if !exists {
		return reply.MakeNullBulkReply()
	}
	db.addAofFunc(utils.ToCommandLine3("GETSET", args...))
	return reply.MakeBulkReply(oldEntity.Data.([]byte))
}

// execStrLen executes the strlen command.
// STRLEN key
func execStrLen(db *Database, args [][]byte) resp.Reply {
	entity, exists := db.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeIntReply(int64(len(entity.Data.([]byte))))
}
