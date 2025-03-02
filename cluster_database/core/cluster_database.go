package core

import (
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/aof"
	command2 "go-redis/cluster_database/command"
	"go-redis/config"
	"go-redis/database"
	databaseInterface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistent_hash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strings"
	"time"
)

type ClusterDatabase struct {
	database   databaseInterface.DatabaseEngine
	aofHandler *aof.AofHandler
	self       string

	nodes           []string
	peerPicker      *consistent_hash.NodeMap
	peerConnections map[string]*pool.ObjectPool
}

// NewClusterDatabase returns a new ClusterDatabase
func NewClusterDatabase() *ClusterDatabase {
	clusterDatabase := &ClusterDatabase{
		self:            config.Properties.Self,
		database:        database.NewStandaloneDatabase(),
		peerPicker:      consistent_hash.NewNodeMap(nil),
		peerConnections: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	clusterDatabase.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		clusterDatabase.peerConnections[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &ConnectionFactory{
			Peer: peer,
		})
	}
	clusterDatabase.nodes = nodes
	return clusterDatabase
}

func (cluster *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			result = reply.MakeUnknownErrorReply()
		}
	}()
	command := strings.ToLower(string(args[0]))
	commands := command2.Commands.GetCommands()
	if commandFunc, ok := commands[command]; ok {
		result = commandFunc(cluster, client, args)
	} else {
		result = reply.MakeStandardErrorReply("not support command: " + string(args[0]))
	}
	return
}

func (cluster *ClusterDatabase) Close() {
	cluster.database.Close()
}

func (cluster *ClusterDatabase) AfterClientClose(conn resp.Connection) {
	cluster.database.AfterClientClose(conn)
}

// TODO
func (cluster *ClusterDatabase) ForEach(_ int, _ func(key string, data *databaseInterface.DataEntity, expiration *time.Time) bool) {
}

func (cluster *ClusterDatabase) GetPeerNode(key string) string {
	return cluster.peerPicker.GetNode(key)
}

func (cluster *ClusterDatabase) GetDatabase() databaseInterface.DatabaseEngine {
	return cluster.database
}
