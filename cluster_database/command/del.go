package command

import (
	"go-redis/interface/cluster_database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// Del deletes the specified keys to the cluster
func Del(cluster cluster_database.ClusterDatabase, conn resp.Connection, args database.CommandLine) resp.Reply {
	var deleted int64 = 0
	results := cluster.Broadcast(conn, utils.ToCommandLine("FLUSHDB"))
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
