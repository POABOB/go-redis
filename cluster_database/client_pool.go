package cluster_database

import (
	"context"
	"errors"
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/resp/client"
)

type ConnectionFactory struct {
	Peer string
}

func (c *ConnectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	clientConnection, err := client.MakeClient(c.Peer)
	if err != nil {
		return nil, err
	}
	clientConnection.Start() // start pipeline
	return pool.NewPooledObject(clientConnection), nil
}

func (c *ConnectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	clientConnection, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("type mismatch")
	}
	clientConnection.Close()
	return nil
}

func (c *ConnectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (c *ConnectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (c *ConnectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
