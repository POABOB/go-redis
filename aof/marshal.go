package aof

import (
	databaseInterface "go-redis/interface/database"
	"go-redis/resp/reply"
)

var (
	setCommand = []byte("SET")
)

// EntityToCommand serialize data entity to redis command
func EntityToCommand(key string, entity *databaseInterface.DataEntity) *reply.MultiBulkReply {
	if entity == nil {
		return nil
	}
	var command *reply.MultiBulkReply
	switch val := entity.Data.(type) {
	case []byte:
		command = stringToCommand(key, val)
	}
	return command
}

// stringToCommand serialize string type data to redis command
func stringToCommand(key string, bytes []byte) *reply.MultiBulkReply {
	args := [][]byte{
		setCommand,
		[]byte(key),
		bytes,
	}
	return reply.MakeMultiBulkReply(args)
}
