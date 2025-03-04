package database

import (
	"go-redis/data_struct/dict"
	"go-redis/interface/database"
	dictInterface "go-redis/interface/dict"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"
	"time"
)

type DictEntity struct {
	index      int // the index of the database
	dict       dictInterface.Dict
	addAofFunc func(database.CommandLine)
}

// MakeDatabase creates a new database
func MakeDatabase() *DictEntity {
	return &DictEntity{index: 0, dict: dict.MakeShardedDict(), addAofFunc: func(commandLine database.CommandLine) {}}
}

func (dict *DictEntity) Exec(c resp.Connection, commandLine database.CommandLine) resp.Reply {
	commandName := strings.ToLower(string(commandLine[0]))
	command, ok := commandTable[commandName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown commands '" + commandName + "'")
	}
	if !dict.validateArity(command.arity, commandLine) {
		return reply.MakeArgsNumErrorReply(commandName)
	}
	connFn := command.connExecutor
	fn := command.executor
	if connFn != nil {
		return connFn(&c, commandLine[1:])
	}
	return fn(dict, commandLine[1:]) // Set key value -> key value
}

// validateArity checks if the arity of the commands is valid.
// -{nums} means at least nums, {nums} means exactly nums.
func (dict *DictEntity) validateArity(arity int, commandArgs [][]byte) bool {
	if arity < 0 {
		return len(commandArgs) >= -arity
	}
	return len(commandArgs) == arity
}

// GetEntity returns the entity for the given key
func (dict *DictEntity) GetEntity(key string) (*database.DataEntity, bool) {
	value, exists := dict.dict.Get(key)
	if !exists {
		return nil, false
	}
	return value.(*database.DataEntity), true
}

// SetEntity sets the entity for the given key and returns the number of entities set
func (dict *DictEntity) SetEntity(key string, entity *database.DataEntity) int {
	return dict.dict.Set(key, entity)
}

// SetEntityIfAbsent sets the entity for the given key, if the key does not exist
func (dict *DictEntity) SetEntityIfAbsent(key string, entity *database.DataEntity) int {
	return dict.dict.SetIfAbsent(key, entity)
}

// SetEntityIfExists sets the entity for the given key, if the key exists
func (dict *DictEntity) SetEntityIfExists(key string, entity *database.DataEntity) int {
	return dict.dict.SetIfExists(key, entity)
}

// DeleteEntity deletes the entity for the given key and returns the number of entities deleted
func (dict *DictEntity) DeleteEntity(key string) int {
	return dict.dict.Delete(key)
}

// DeleteEntities deletes the entities for the given keys
func (dict *DictEntity) DeleteEntities(keys ...string) int {
	deletedCount := 0
	for _, key := range keys {
		// dict.Delete will delete the key if it exists
		deletedCount += dict.DeleteEntity(key)
	}
	return deletedCount
}

// GetAndDeleteEntity gets the entity for the given key and deletes it
func (dict *DictEntity) GetAndDeleteEntity(key string) (*database.DataEntity, bool) {
	value, exists := dict.dict.GetAndDelete(key)
	return value.(*database.DataEntity), exists
}

// ForEach iterates over all the entities in the database
func (dict *DictEntity) ForEach(cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {
	dict.dict.ForEach(func(key string, value interface{}) bool {
		return cb(key, value.(*database.DataEntity), nil)
	})
}

// Flush flushes the database
func (dict *DictEntity) Flush() {
	dict.dict.Clear()
}
