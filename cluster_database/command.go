package cluster_database

import (
	"go-redis/interface/resp"
	"strings"
)

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
func RegisterCommand(name string, commandFunc CommandFunc) {
	name = strings.ToLower(name)
	commands[name] = commandFunc
}

// defaultFunc use consistent hash to select a peer
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.GetNode(key)
	return cluster.relayToPeer(peer, conn, args)
}

func init() {
	RegisterCommands()
}
