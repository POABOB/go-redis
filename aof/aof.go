package aof

import (
	"context"
	"go-redis/config"
	"go-redis/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
	"os"
	"strconv"
	"sync"
)

type CommandLine = [][]byte

const (
	aofQueueSize = 1 << 20

	FsyncAlways   = "always"   // FsyncAlways do fsync for every command
	FsyncEverySec = "everysec" // FsyncEverySec do fsync every second
	FsyncNo       = "no"       // FsyncNo lets operating system decides when to do fsync
)

type payload struct {
	commandLine CommandLine
	dbIndex     int
	wg          *sync.WaitGroup
}

// Listener will be called-back after receiving a aof payload with a listener we can forward the updates to slave nodes etc.
type Listener interface {
	Callback([]CommandLine)
}

// AofHandler receive messages from channel and write to AOF file
type AofHandler struct {
	ctx         context.Context
	cancel      context.CancelFunc
	database    *database.DatabaseEngine
	aofChan     chan *payload // aofChan is the channel to receive aof payload(listenCmd will send payload to this channel)
	aofFile     *os.File      // aofFile is the file handler of aof file
	aofFilename string        // aofFilename is the path of aof file
	aofFsync    string        // aofFsync is the strategy of fsync
	aofFinished chan struct{} // aof goroutine will send msg to main goroutine through this channel when aof tasks finished and ready to shut down
	pausingAof  sync.Mutex    // pause aof for start/finish aof rewrite progress
	currentDB   int
	listeners   map[Listener]struct{}
	buffer      []CommandLine // reuse commandLine buffer
}

// NewAofHandler returns a new instance of AofHandler and open aof file
func NewAofHandler(database *database.DatabaseEngine) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.database = database
	handler.LoadAof()
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	// run a goroutine to handle aof
	go func() {
		handler.HandleAof()
	}()
	return handler, nil
}

// AddAof adds aof payload to aofChan
func (handler *AofHandler) AddAof(databaseIndex int, commandLine CommandLine) {
	if !config.Properties.AppendOnly && handler.aofChan == nil {
		return
	}
	handler.aofChan <- &payload{commandLine: commandLine, dbIndex: databaseIndex}
}

func (handler *AofHandler) HandleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { // switch db
			selectCommand := utils.ToCommandLine("SELECT", strconv.Itoa(p.dbIndex))
			if success := execWriteAofFile(handler, selectCommand); !success {
				continue
			}
			handler.currentDB = p.dbIndex
		}
		if success := execWriteAofFile(handler, p.commandLine); !success {
			continue
		}
	}
}

func (handler *AofHandler) LoadAof() {

}

// execWriteAofFile writes aof payload to aof file
func execWriteAofFile(handler *AofHandler, commandLine CommandLine) bool {
	data := reply.MakeMultiBulkReply(commandLine).ToBytes()
	_, err := handler.aofFile.Write(data)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}
