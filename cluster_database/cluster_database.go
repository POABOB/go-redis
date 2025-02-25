package cluster_database

import (
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/aof"
	"go-redis/config"
	"go-redis/database"
	databaseInterface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistent_hash"
)

// commandFunc is a function that executes a command
type commandFunc func(cluster *ClusterDatabase, conn resp.Connection, commandAndArgs [][]byte) resp.Reply

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

//
//func (cluster_database *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
//	defer func() {
//		if err := recover(); err != nil {
//			logger.Error(err)
//			result = reply.MakeUnknownErrorReply()
//		}
//	}()
//
//}
//
//func (cluster_database *ClusterDatabase) Close() {
//	cluster_database.database.Close()
//}
//
//func (cluster_database *ClusterDatabase) AfterClientClose(conn resp.Connection) {
//	cluster_database.database.AfterClientClose(conn)
//}
//
//func (cluster_database *ClusterDatabase) ForEach(dbIndex int, cb func(key string, data *databaseInterface.DataEntity, expiration *time.Time) bool) {
//}
