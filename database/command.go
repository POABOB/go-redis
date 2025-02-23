package database

import "strings"

var commandTable = make(map[string]*command)

type command struct {
	executor ExecFunc // the function to execute the command
	arity    int      // the number of arguments required by the command
}

// RegisterCommand registers a new command
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	commandTable[name] = &command{executor: executor, arity: arity}
}
