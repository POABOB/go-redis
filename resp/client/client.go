package client

import (
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/lib/sync/wait"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

const (
	chanSize = 256
	maxWait  = 3 * time.Second
)

// Client is a pipeline mode redis client
type Client struct {
	connection  net.Conn
	pendingReqs chan *request // wait to send
	waitingReqs chan *request // waiting response
	ticker      *time.Ticker
	address     string

	working *sync.WaitGroup // its counter presents unfinished requests(pending and waiting)
}

// request is a message sends to redis server
type request struct {
	id        uint64
	args      [][]byte
	reply     resp.Reply
	heartbeat bool
	waiting   *wait.Wait
	err       error
}

// MakeClient creates a new client
func MakeClient(address string) (*Client, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Client{
		address:     address,
		connection:  connection,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
		working:     &sync.WaitGroup{},
	}, nil
}

// Start starts asynchronous goroutines
func (client *Client) Start() {
	client.ticker = time.NewTicker(10 * time.Second)
	go client.handleWrite()
	go func() {
		err := client.handleRead()
		if err != nil {
			logger.Error(err)
		}
	}()
	go client.heartbeat()
}

// handleWrite sends requests
func (client *Client) handleWrite() {
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

// doRequest sends a request and handles errors
func (client *Client) doRequest(req *request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	re := reply.MakeMultiBulkReply(req.args)
	bytes := re.ToBytes()
	_, err := client.connection.Write(bytes)
	i := 0
	for err != nil && i < 3 {
		err = client.handleConnectionError(err)
		if err == nil {
			_, err = client.connection.Write(bytes)
		}
		i++
	}
	if err == nil {
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

// handleConnectionError handles connection error and reconnect
func (client *Client) handleConnectionError(err error) error {
	err1 := client.connection.Close()
	if err1 != nil {
		var opErr *net.OpError
		if errors.As(err1, &opErr) {
			if opErr.Err.Error() != "use of closed network connection" {
				return err1
			}
		}
	}
	conn, err1 := net.Dial("tcp", client.address)
	if err1 != nil {
		logger.Error(err1)
		return err1
	}
	client.connection = conn
	go func() {
		_ = client.handleRead()
	}()
	return nil
}

// handleRead reads all responses from the connection
func (client *Client) handleRead() error {
	ch := parser.ParseStream(client.connection)
	for payload := range ch {
		if payload.Error != nil {
			client.finishRequest(reply.MakeStandardErrorReply(payload.Error.Error()))
			continue
		}
		client.finishRequest(payload.Data)
	}
	return nil
}

// finishRequest receives a response
func (client *Client) finishRequest(reply resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()
	request := <-client.waitingReqs
	if request == nil {
		return
	}
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
}

// heartbeat uses ticker to send a heartbeat request
func (client *Client) heartbeat() {
	for range client.ticker.C {
		client.doHeartbeat()
	}
}

// doHeartbeat sends a heartbeat request
func (client *Client) doHeartbeat() {
	request := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	request.waiting.WaitWithTimeout(maxWait)
}

// Send sends a request to redis server
func (client *Client) Send(args [][]byte) resp.Reply {
	request := &request{
		args:      args,
		heartbeat: false,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	timeout := request.waiting.WaitWithTimeout(maxWait)
	if timeout {
		return reply.MakeStandardErrorReply("server time out")
	}
	if request.err != nil {
		return reply.MakeStandardErrorReply("request failed")
	}
	return request.reply
}

// Close stops asynchronous goroutines and close connection
func (client *Client) Close() {
	client.ticker.Stop()
	// stop new request
	close(client.pendingReqs)

	// wait stop process
	client.working.Wait()

	// clean
	_ = client.connection.Close()
	close(client.waitingReqs)
}
