package cluster_database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// Rename is used to rename a key
func Rename(cluster *ClusterDatabase, conn resp.Connection, args CommandLine) resp.Reply {
	if len(args) != 3 {
		return reply.MakeArgsNumErrorReply(string(args[0]))
	}

	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.peerPicker.GetNode(src)
	destPeer := cluster.peerPicker.GetNode(dest)

	if srcPeer == destPeer {
		return cluster.relayToPeer(srcPeer, conn, args)
	}

	// delete the key of source peer, and set the key to the destination peer
	sourceReply := cluster.relayToPeer(srcPeer, conn, utils.ToCommandLine("GETDEL", src))
	bulkReply, ok := sourceReply.(*reply.BulkReply)
	if !ok || bulkReply.Arg == nil {
		return reply.MakeStandardErrorReply("no such key")
	}
	setReply := cluster.relayToPeer(destPeer, conn, utils.ToCommandLine("SET", dest, string(bulkReply.Arg)))
	if reply.IsErrorReply(setReply) {
		return setReply
	}
	return reply.MakeOkReply()
}

func init() {
	RegisterCommand("RENAME", Rename)
	RegisterCommand("RENAMENX", Rename)
}
