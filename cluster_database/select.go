package cluster_database

import (
	"go-redis/interface/resp"
)

// Select is used to select a database
func Select(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	return cluster.database.Exec(conn, args)
}

func init() {
	RegisterCommand("SELECT", Select)
}
