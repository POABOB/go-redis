package cluster_database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// FlushDB is used to flush the cluster of all databases
func FlushDB(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	results := cluster.broadcast(conn, utils.ToCommandLine("FLUSHDB"))
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
