package cluster_database

import (
	"go-redis/interface/resp"
)

// Ping is used to reply ping command.
func Ping(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	return cluster.database.Exec(conn, args)
}

func init() {
	RegisterCommand("PING", Ping)
}
