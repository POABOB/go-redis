package database

import (
	databaseInterface "go-redis/interface/database"
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
	RegisterCommand("GETDEL", execGetDel, 2)
	RegisterCommand("STRLEN", execStrLen, 2)
}

// execGet executes the get commands.
// GET key
func execGet(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	entity, exists := dictEntity.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	// TODO if there are different types, entity.Data.([]byte) need to check
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// execSet executes the set commands.
// SET key value
func execSet(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	dictEntity.SetEntity(string(args[0]), &databaseInterface.DataEntity{Data: args[1]})
	dictEntity.addAofFunc(utils.ToCommandLine3("SET", args...))
	return reply.MakeOkReply()
}

// execSetNx executes the setnx commands.
// SETNX key value
func execSetNx(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	defer dictEntity.addAofFunc(utils.ToCommandLine3("SETNX", args...))
	return reply.MakeIntReply(int64(dictEntity.SetEntityIfAbsent(string(args[0]), &databaseInterface.DataEntity{Data: args[1]})))
}

// execGetSet executes the getset commands.
// GETSET key value
func execGetSet(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	oldEntity, exists := dictEntity.GetEntity(string(args[0]))
	dictEntity.SetEntity(string(args[0]), &databaseInterface.DataEntity{Data: args[1]})
	if !exists {
		return reply.MakeNullBulkReply()
	}
	dictEntity.addAofFunc(utils.ToCommandLine3("GETSET", args...))
	return reply.MakeBulkReply(oldEntity.Data.([]byte))
}

// execGetDel executes the getdel commands.
// GETDEL key
func execGetDel(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	oldEntity, exists := dictEntity.GetAndDeleteEntity(string(args[0]))
	if oldEntity == nil || !exists {
		return reply.MakeNullBulkReply()
	}
	dictEntity.addAofFunc(utils.ToCommandLine3("GETDEL", args...))
	return reply.MakeBulkReply(oldEntity.Data.([]byte))
}

// execStrLen executes the strlen commands.
// STRLEN key
func execStrLen(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	entity, exists := dictEntity.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeIntReply(int64(len(entity.Data.([]byte))))
}
