package datebase

import "go-redis/interface/resp"

type CommandLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(client resp.Connection)
}

type DataEntity struct {
	Data interface{}
}
