package connection

import (
	"go-redis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

type Connection struct {
	connection   net.Conn   // connection instance
	waitingReply wait.Wait  // WaitGroup with timeout feature
	mutex        sync.Mutex // Mutex Lock
	selectedDB   int        // DB index
	password     string     // login pass
}

// NewConnection creates a new instance of Connection
func NewConnection(conn net.Conn) *Connection {
	return &Connection{connection: conn}
}

// RemoteAddress returns the remote address
func (c *Connection) RemoteAddress() net.Addr {
	return c.connection.RemoteAddr()
}

// Write writes data to the connection and increments the waitingReply counter
func (c *Connection) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	c.mutex.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mutex.Unlock()
	}()
	_, err := c.connection.Write(bytes)
	return err
}

// GetDBIndex returns the DB index
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB sets the DB index
func (c *Connection) SelectDB(dbIndex int) {
	c.selectedDB = dbIndex
}

// SetPassword stores password for authentication
func (c *Connection) SetPassword(password string) {
	c.password = password
}

// GetPassword get password for authentication
func (c *Connection) GetPassword() string {
	return c.password
}

// Close closes the connection while timeout
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * 1000 * time.Millisecond)
	// no need to return error while the connection is closed
	_ = c.connection.Close()
	c.password = ""
	return nil
}
