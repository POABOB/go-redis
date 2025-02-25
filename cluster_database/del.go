package cluster_database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// Del deletes the specified keys to the cluster
func Del(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	var deleted int64 = 0
	results := cluster.broadcast(conn, utils.ToCommandLine("FLUSHDB"))
	for _, result := range results {
		if reply.IsErrorReply(result) {
			return reply.MakeStandardErrorReply("error: " + result.(resp.ErrorReply).Error())
		}
		if intReply, ok := result.(*reply.IntReply); ok {
			deleted += intReply.Code
		}
	}
	return reply.MakeIntReply(deleted)
}

func init() {
	RegisterCommand("DEL", Del)
}
