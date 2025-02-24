package database

import (
	"go-redis/interface/resp"
	"time"
)

type CommandLine = [][]byte

// Database is the interface for redis style storage engine
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(client resp.Connection)
}

// DatabaseEngine is the embedding storage engine exposing more methods for complex application
type DatabaseEngine interface {
	Database
	ForEach(dbIndex int, cb func(key string, data *DataEntity, expiration *time.Time) bool)
}

type DataEntity struct {
	Data interface{}
}
