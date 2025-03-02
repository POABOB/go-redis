package database

import (
	"go-redis/data_struct/dict"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"
	"time"
)

type Database struct {
	index      int // the index of the database
	dict       dict.Dict
	addAofFunc func(database.CommandLine)
}

// ExecFunc is a function that executes a command in a database
type ExecFunc func(db *Database, args [][]byte) resp.Reply

// MakeDatabase creates a new database
func MakeDatabase() *Database {
	return &Database{index: 0, dict: dict.MakeShardedDict(), addAofFunc: func(commandLine database.CommandLine) {}}
}

func (db *Database) Exec(_ resp.Connection, commandLine database.CommandLine) resp.Reply {
	commandName := strings.ToLower(string(commandLine[0]))
	cmd, ok := commandTable[commandName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown command '" + commandName + "'")
	}
	if !db.validateArity(cmd.arity, commandLine) {
		return reply.MakeArgsNumErrorReply(commandName)
	}
	fn := cmd.executor
	// Set key value -> key value
	return fn(db, commandLine[1:])
}

// validateArity checks if the arity of the command is valid.
// -{nums} means at least nums, {nums} means exactly nums.
func (db *Database) validateArity(arity int, commandArgs [][]byte) bool {
	if arity < 0 {
		return len(commandArgs) >= -arity
	}
	return len(commandArgs) == arity
}

// GetEntity returns the entity for the given key
func (db *Database) GetEntity(key string) (*database.DataEntity, bool) {
	value, exists := db.dict.Get(key)
	if !exists {
		return nil, false
	}
	return value.(*database.DataEntity), true
}

// SetEntity sets the entity for the given key and returns the number of entities set
func (db *Database) SetEntity(key string, entity *database.DataEntity) int {
	return db.dict.Set(key, entity)
}

// SetEntityIfAbsent sets the entity for the given key, if the key does not exist
func (db *Database) SetEntityIfAbsent(key string, entity *database.DataEntity) int {
	return db.dict.SetIfAbsent(key, entity)
}

// SetEntityIfExists sets the entity for the given key, if the key exists
func (db *Database) SetEntityIfExists(key string, entity *database.DataEntity) int {
	return db.dict.SetIfExists(key, entity)
}

// DeleteEntity deletes the entity for the given key and returns the number of entities deleted
func (db *Database) DeleteEntity(key string) int {
	return db.dict.Delete(key)
}

// DeleteEntities deletes the entities for the given keys
func (db *Database) DeleteEntities(keys ...string) int {
	deletedCount := 0
	for _, key := range keys {
		// db.Delete will delete the key if it exists
		deletedCount += db.DeleteEntity(key)
	}
	return deletedCount
}

// GetAndDeleteEntity gets the entity for the given key and deletes it
func (db *Database) GetAndDeleteEntity(key string) (*database.DataEntity, bool) {
	value, exists := db.dict.GetAndDelete(key)
	return value.(*database.DataEntity), exists
}

// ForEach iterates over all the entities in the database
func (db *Database) ForEach(cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {
	db.dict.ForEach(func(key string, value interface{}) bool {
		return cb(key, value.(*database.DataEntity), nil)
	})
}

// Flush flushes the database
func (db *Database) Flush() {
	db.dict.Clear()
}
