package cluster_database

import (
	"go-redis/interface/database"
	"go-redis/interface/resp"
)

// ClusterDatabase is a cluster database
type ClusterDatabase interface {
	database.DatabaseEngine
	RelayToPeer(peer string, connection resp.Connection, args database.CommandLine) resp.Reply
	Broadcast(connection resp.Connection, args database.CommandLine) map[string]resp.Reply
	GetPeerNode(key string) string
	GetDatabase() database.DatabaseEngine
}

// CommandFunc is a function that executes a command
type CommandFunc func(cluster ClusterDatabase, conn resp.Connection, args database.CommandLine) resp.Reply
