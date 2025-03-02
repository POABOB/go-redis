package command

import (
	"go-redis/interface/cluster_database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"strings"
)

var Commands = &Command{
	commands: make(map[string]cluster_database.CommandFunc),
}

type Command struct {
	commands map[string]cluster_database.CommandFunc
}

func (c *Command) GetCommands() map[string]cluster_database.CommandFunc {
	return c.commands
}

// RegisterCommands returns a router that dispatches commands to peers
func RegisterCommands() {
	defaultCommands := []string{
		"EXISTS",
		"TYPE",
		"RENAME",
		"SET",
		"SETNX",
		"GET",
		"GETSET",
		"PING",
	}
	// TODO more...
	for _, command := range defaultCommands {
		RegisterDefaultCommand(command)
	}
}

func RegisterDefaultCommand(name string) {
	RegisterCommand(name, defaultFunc)
}

// RegisterCommand add command handler into cluster
func RegisterCommand(name string, commandFunc cluster_database.CommandFunc) {
	name = strings.ToLower(name)
	Commands.commands[name] = commandFunc
}

func init() {
	RegisterCommands()
}

// defaultFunc use consistent hash to select a peer
func defaultFunc(cluster cluster_database.ClusterDatabase, conn resp.Connection, args database.CommandLine) resp.Reply {
	key := string(args[1])
	peer := cluster.GetPeerNode(key)
	return cluster.RelayToPeer(peer, conn, args)
}
