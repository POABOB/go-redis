package database

import (
	databaseInterface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

func init() {
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("FLUSHDB", execFlushDB, -1)
	RegisterCommand("TYPE", execType, 2)
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenameNx, 3)
	RegisterCommand("KEYS", execKeys, 2)
}

// execDel executes the del commands.
// DEL key [key ...]
func execDel(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deletedCount := dictEntity.DeleteEntities(keys...)
	if deletedCount > 0 {
		dictEntity.addAofFunc(utils.ToCommandLine2("DEL", keys...))
	}
	return reply.MakeIntReply(int64(deletedCount))
}

// execExists executes the exists commands.
// EXISTS key [key ...]
func execExists(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	var existsCount int64
	for _, v := range args {
		_, exists := dictEntity.GetEntity(string(v))
		if exists {
			existsCount++
		}
	}
	return reply.MakeIntReply(existsCount)
}

// execFlushDB executes the flushdb commands.
// FLUSHDB
func execFlushDB(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	dictEntity.Flush()
	dictEntity.addAofFunc(utils.ToCommandLine3("FLUSHDB", args...))
	return reply.MakeOkReply()
}

// execType executes the type commands.
// TYPE key
func execType(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	entity, exists := dictEntity.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeStatusReply("none")
	}

	switch entity.Data.(type) {
	case [][]byte:
		reply.MakeStatusReply("string")
		// TODO add more types
	}
	return reply.MakeUnknownErrorReply()
}

// execRename executes the rename commands.
// RENAME key new_key
func execRename(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	entity, exists := dictEntity.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeStandardErrorReply("no such key")
	}
	dictEntity.SetEntity(string(args[1]), entity)
	dictEntity.DeleteEntity(string(args[0]))
	dictEntity.addAofFunc(utils.ToCommandLine3("RENAME", args...))
	return reply.MakeOkReply()
}

// execRenameNx executes the renamenx commands.
// RENAMENX key new_key
func execRenameNx(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	if _, exists := dictEntity.GetEntity(string(args[1])); exists {
		return reply.MakeIntReply(0)
	}

	entity, exists := dictEntity.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeStandardErrorReply("no such key")
	}
	dictEntity.SetEntity(string(args[1]), entity)
	dictEntity.DeleteEntity(string(args[0]))
	dictEntity.addAofFunc(utils.ToCommandLine3("RENAMENX", args...))
	return reply.MakeIntReply(1)
}

// execKeys executes the keys commands.
// KEYS pattern
func execKeys(dictEntity *DictEntity, args databaseInterface.CommandLine) resp.Reply {
	pattern, err := wildcard.CompilePattern(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	result := make([][]byte, 0)
	dictEntity.dict.ForEach(func(key string, value interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}
