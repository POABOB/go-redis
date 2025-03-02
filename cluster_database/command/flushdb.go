package command

import (
	"go-redis/interface/cluster_database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// FlushDB is used to flush the cluster of all databases
func FlushDB(cluster cluster_database.ClusterDatabase, conn resp.Connection, _ database.CommandLine) resp.Reply {
	results := cluster.Broadcast(conn, utils.ToCommandLine("FLUSHDB"))
	for _, result := range results {
		if reply.IsErrorReply(result) {
			return reply.MakeStandardErrorReply("error: " + result.(resp.ErrorReply).Error())
		}
	}
	return reply.MakeOkReply()
}

func init() {
	RegisterCommand("FLUSHDB", FlushDB)
}
