package database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

// execDel executes the del command.
// DEL key [key ...]
func execDel(db *Database, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deletedCount := db.DeleteEntities(keys...)
	if deletedCount > 0 {
		db.addAofFunc(utils.ToCommandLine2("DEL", keys...))
	}
	return reply.MakeIntReply(int64(deletedCount))
}

// execExists executes the exists command.
// EXISTS key [key ...]
func execExists(db *Database, args [][]byte) resp.Reply {
	var existsCount int64
	for _, v := range args {
		_, exists := db.GetEntity(string(v))
		if exists {
			existsCount++
		}
	}
	return reply.MakeIntReply(existsCount)
}

// execFlushDB executes the flushdb command.
// FLUSHDB
func execFlushDB(db *Database, args [][]byte) resp.Reply {
	db.Flush()
	db.addAofFunc(utils.ToCommandLine3("FLUSHDB", args...))
	return reply.MakeOkReply()
}

// execType executes the type command.
// TYPE key
func execType(db *Database, args [][]byte) resp.Reply {
	entity, exists := db.GetEntity(string(args[0]))
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

// execRename executes the rename command.
// RENAME key new_key
func execRename(db *Database, args [][]byte) resp.Reply {
	entity, exists := db.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeStandardErrorReply("no such key")
	}
	db.SetEntity(string(args[1]), entity)
	db.DeleteEntity(string(args[0]))
	db.addAofFunc(utils.ToCommandLine3("RENAME", args...))
	return reply.MakeOkReply()
}

// execRenameNx executes the renamenx command.
// RENAMENX key new_key
func execRenameNx(db *Database, args [][]byte) resp.Reply {
	if _, exists := db.GetEntity(string(args[1])); exists {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(string(args[0]))
	if !exists {
		return reply.MakeStandardErrorReply("no such key")
	}
	db.SetEntity(string(args[1]), entity)
	db.DeleteEntity(string(args[0]))
	db.addAofFunc(utils.ToCommandLine3("RENAMENX", args...))
	return reply.MakeIntReply(1)
}

// execKeys executes the keys command.
// KEYS pattern
func execKeys(db *Database, args [][]byte) resp.Reply {
	pattern, err := wildcard.CompilePattern(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	result := make([][]byte, 0)
	db.dict.ForEach(func(key string, value interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("FLUSHDB", execFlushDB, -1)
	RegisterCommand("TYPE", execType, 2)
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenameNx, 3)
	RegisterCommand("KEYS", execKeys, 2)
}
