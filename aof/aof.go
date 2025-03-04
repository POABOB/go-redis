package aof

import (
	"context"
	"go-redis/config"
	databaseInterface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	aofQueueSize  = 1 << 20 // 1MB
	bufferSize    = 1 << 16 // 64KB buffer
	fsyncInterval = time.Second
	retryTimes    = 3

	FsyncAlways   = "always"   // FsyncAlways do fsync for every command
	FsyncEverySec = "everysec" // FsyncEverySec do fsync every second
	FsyncNo       = "no"       // FsyncNo lets operating system decides when to do fsync
)

type payload struct {
	commandLine databaseInterface.CommandLine
	dbIndex     int
	wg          *sync.WaitGroup
}

// Listener will be called-back after receiving an aof payload with a listener we can forward the updates to slave nodes etc.
type Listener interface {
	Callback([]databaseInterface.CommandLine)
}

// AofHandler receive messages from channel and write to AOF file
type AofHandler struct {
	database   databaseInterface.DatabaseEngine
	currentDB  int
	buffer     []byte // reuse commandLine buffer
	bufferLock sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	aofChan     chan *payload  // aofChan is the channel to receive aof payload(listenCmd will send payload to this channel)
	aofFile     *os.File       // aofFile is the file handler of aof file
	aofFilename string         // aofFilename is the path of aof file
	aofFsync    string         // aofFsync is the strategy of fsync
	aofFinished sync.WaitGroup // aofFinished is used to wait aof rewrite is finished

	aofRewriter  *AofRewriter
	pausingMutex sync.Mutex

	listeners map[Listener]struct{} // listeners is the map of listeners for other nodes
}

// NewAofHandler returns a new instance of AofHandler and open aof file
func NewAofHandler(database databaseInterface.DatabaseEngine) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.database = database
	handler.aofRewriter = NewAofRewriter(database)

	// Redis aofFsync default is "everysec"
	switch config.Properties.AppendFsync {
	case "always":
		handler.aofFsync = FsyncAlways
	case "no":
		handler.aofFsync = FsyncNo
	default:
		handler.aofFsync = FsyncEverySec
	}

	handler.LoadAof()
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	// run a goroutine to handle aof
	handler.ctx, handler.cancel = context.WithCancel(context.Background())
	handler.aofFinished.Add(1)
	go handler.HandleAof()
	if handler.aofFsync == FsyncEverySec {
		go handler.periodicFsync()
	}
	return handler, nil
}

// AddAof adds aof payload to aofChan
func (handler *AofHandler) AddAof(databaseIndex int, commandLine databaseInterface.CommandLine) {
	if !config.Properties.AppendOnly && handler.aofChan == nil {
		return
	}
	handler.aofChan <- &payload{commandLine: commandLine, dbIndex: databaseIndex}
}

// HandleAof handles aof payload
func (handler *AofHandler) HandleAof() {
	// wait for aof rewrite is finished
	defer handler.aofFinished.Done()

	handler.currentDB = 0
	for {
		select {
		case p := <-handler.aofChan:
			handler.processAofPayload(p)
		case <-handler.ctx.Done(): // close aof
			return
		}
	}
}

// processAofPayload
func (handler *AofHandler) processAofPayload(p *payload) {
	// this lock is used to prevent concurrent write
	handler.pausingMutex.Lock()
	defer handler.pausingMutex.Unlock()

	if p.dbIndex != handler.currentDB {
		selectCommand := utils.ToCommandLine("SELECT", strconv.Itoa(p.dbIndex))
		handler.bufferedWrite(selectCommand)
		handler.currentDB = p.dbIndex
	}
	handler.bufferedWrite(p.commandLine)
	if handler.aofFsync == FsyncAlways || len(handler.buffer) >= bufferSize {
		handler.flushBuffer()
	}
}

// LoadAof load aof when redis start
func (handler *AofHandler) LoadAof() {
	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer func() {
		_ = file.Close() // prevent memory leak
	}()
	ch := parser.ParseStream(file)
	fakeConnection := &connection.Connection{}
	fakeConnection.SetPassword(config.Properties.RequirePass)
	for payload := range ch {
		if payload.Error != nil {
			if payload.Error == io.EOF {
				break
			}
			logger.Error(payload.Error)
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		multiBulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("need multi bulk")
			continue
		}
		databaseReply := handler.database.Exec(fakeConnection, multiBulkReply.Args)
		if reply.IsErrorReply(databaseReply) {
			logger.Error("exec aof error", string(databaseReply.ToBytes()))
			continue
		}
	}
}

// bufferedWrite writes commandLine to aof buffer
func (handler *AofHandler) bufferedWrite(commandLine databaseInterface.CommandLine) {
	handler.bufferLock.Lock()
	defer handler.bufferLock.Unlock()
	data := reply.MakeMultiBulkReply(commandLine).ToBytes()
	handler.buffer = append(handler.buffer, data...)
}

// flushBuffer flushes aof buffer to disk
func (handler *AofHandler) flushBuffer() {
	handler.bufferLock.Lock()
	defer handler.bufferLock.Unlock()
	if len(handler.buffer) == 0 {
		return
	}

	// retry
	var err error
	for i := 0; i < retryTimes; i++ {
		_, err = handler.aofFile.Write(handler.buffer)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		logger.Error("AOF write error after retries: ", err)
		return
	}

	handler.buffer = handler.buffer[:0] // Clear buffer
	if handler.aofFsync != FsyncNo {
		handler.safeSync()
	}
	fileInfo, _ := handler.aofFile.Stat()
	currentAofSize := fileInfo.Size()
	handler.checkAofRewrite(currentAofSize)
}

// periodicFsync flushes aof buffer to disk in a period
func (handler *AofHandler) periodicFsync() {
	ticker := time.NewTicker(fsyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-handler.ctx.Done():
			return
		case <-ticker.C:
			handler.flushBuffer()
		}
	}
}

// safeSync flushes aof buffer to disk with retry attempts
func (handler *AofHandler) safeSync() {
	for i := 0; i < retryTimes; i++ {
		if err := handler.aofFile.Sync(); err != nil {
			logger.Error("AOF fsync failed, retrying... Attempt", i+1, "Error:", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return
	}
	logger.Error("AOF fsync failed after multiple attempts. Data consistency may be at risk.")
}

// ScheduleRewrite schedule aof rewrite
func (handler *AofHandler) ScheduleRewrite() {
	// this lock is used to prevent concurrent rewrite
	if !handler.aofRewriter.TryLock() {
		return
	}

	handler.aofFinished.Add(1)
	go func() {
		defer handler.aofFinished.Done()
		handler.pausingMutex.Lock()
		defer handler.pausingMutex.Unlock()
		// Before starting rewrite, fsync the current AOF file to ensure data consistency
		handler.safeSync()

		err := handler.aofRewriter.TriggerRewrite()
		if err != nil {
			logger.Error("AOF rewrite failed:", err)
			return
		}
		// Close old aofFile and open the new one
		_ = handler.aofFile.Close()
		aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			panic(err)
		}
		handler.aofFile = aofFile
		// Reselect current db after the rewrite
		data := reply.MakeMultiBulkReply(utils.ToCommandLine("SELECT", strconv.Itoa(handler.currentDB))).ToBytes()
		_, err = handler.aofFile.Write(data)
		if err != nil {
			panic(err)
		}
	}()
}

// checkAofRewrite check aof rewrite rule when aof file size need rewrite
func (handler *AofHandler) checkAofRewrite(fileSize int64) {
	autoAofRewriteMinSize, _ := utils.ParseSize(config.Properties.AutoAofRewriteMinSize)
	atoAofRewritePercentage := config.Properties.AutoAofRewritePercentage
	lastRewriteSize := handler.aofRewriter.lastRewriteSize
	if fileSize >= autoAofRewriteMinSize && (fileSize-lastRewriteSize)/lastRewriteSize >= atoAofRewritePercentage {
		handler.ScheduleRewrite()
	}
}

// Close closes aof
func (handler *AofHandler) Close() {
	handler.cancel()            // trigger ctx.Done()，make all goroutines exit
	handler.aofFinished.Wait()  // wait all goroutines exit
	handler.safeSync()          // safe flush
	_ = handler.aofFile.Close() // close aof file
}
