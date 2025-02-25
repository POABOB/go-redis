package cluster_database

import (
	"context"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/client"
	"go-redis/resp/reply"
	"strconv"
)

// getPeerClient returns a peer client from the connection pool
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnections[peer]
	if !ok {
		return nil, errors.New("peer not found")
	}
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	clientConnection, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("type mismatch")
	}
	return clientConnection, nil
}

// returnPeerClient returns a peer client to the connection pool
func (cluster *ClusterDatabase) returnPeerClient(peer string, clientConnection *client.Client) error {
	pool, ok := cluster.peerConnections[peer]
	if !ok {
		return errors.New("peer not found")
	}
	return pool.ReturnObject(context.Background(), clientConnection)
}

// relayToPeer relays a command to a peer or the local database
func (cluster *ClusterDatabase) relayToPeer(peer string, connection resp.Connection, args CommandLine) resp.Reply {
	if peer == cluster.self {
		return cluster.database.Exec(connection, args)
	}
	clientConnection, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, clientConnection)
	}()
	clientConnection.Send(utils.ToCommandLine("SELECT", strconv.Itoa(connection.GetDBIndex())))
	return clientConnection.Send(args)
}

func (cluster *ClusterDatabase) broadcast(connection resp.Connection, args CommandLine) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, peer := range cluster.nodes {
		results[peer] = cluster.relayToPeer(peer, connection, args)
	}
	return results
}
