package database

import (
	"go-redis/interface/resp"
	"strings"
)

var commandTable = make(map[string]*command)

// ExecFunc is a function that executes a commands in a database
type ExecFunc func(dict *DictEntity, args [][]byte) resp.Reply

// ExecSysFunc is a function that executes a commands in a connection
type ExecSysFunc func(dict *resp.Connection, args [][]byte) resp.Reply

type command struct {
	connExecutor ExecSysFunc // the function to execute the command when connection
	executor     ExecFunc    // the function to execute the command
	arity        int         // the number of arguments required by the command
}

// RegisterCommand registers a new commands
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	commandTable[name] = &command{executor: executor, arity: arity}
}

// RegisterSysCommand registers a new system commands
func RegisterSysCommand(name string, connExecutor ExecSysFunc, arity int) {
	name = strings.ToLower(name)
	commandTable[name] = &command{connExecutor: connExecutor, arity: arity}
}
