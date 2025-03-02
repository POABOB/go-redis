package command

import (
	"go-redis/interface/cluster_database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
)

// Select is used to select a database
func Select(cluster cluster_database.ClusterDatabase, conn resp.Connection, args database.CommandLine) resp.Reply {
	return cluster.GetDatabase().Exec(conn, args)
}

func init() {
	RegisterCommand("SELECT", Select)
}
