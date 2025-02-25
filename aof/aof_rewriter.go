package aof

import (
	"go-redis/config"
	databaseInterface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
	"os"
	"strconv"
	"sync"
	"time"
)

var commandPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

type AofRewriter struct {
	database        databaseInterface.DatabaseEngine
	lastRewriteSize int64
	tempFile        *os.File
	rewriteMutex    sync.Mutex
}

// NewAofRewriter creates a new AOF rewriter instance
func NewAofRewriter(database databaseInterface.DatabaseEngine) *AofRewriter {
	return &AofRewriter{database: database}
}

func (rewriter *AofRewriter) TriggerRewrite() error {
	rewriter.rewriteMutex.Lock()
	defer rewriter.rewriteMutex.Unlock()

	var err error
	tempFile, err := os.CreateTemp("", config.Properties.AppendFilename+".rewrite")
	if err != nil {
		return err
	}
	rewriter.tempFile = tempFile
	defer func() {
		_ = rewriter.tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	if err = rewriter.dumpDatabase(); err != nil {
		return err
	}
	err = os.Rename(rewriter.tempFile.Name(), config.Properties.AppendFilename)
	if err != nil {
		return err
	}
	fileInfo, _ := os.Stat(config.Properties.AppendFilename)
	rewriter.lastRewriteSize = fileInfo.Size()
	logger.Info("AOF rewrite completed successfully")
	return nil
}

func (rewriter *AofRewriter) dumpDatabase() error {
	for dbIndex := 0; dbIndex < config.Properties.Databases; dbIndex++ {
		// select db
		data := reply.MakeMultiBulkReply(utils.ToCommandLine("SELECT", strconv.Itoa(dbIndex))).ToBytes()
		_, err := rewriter.tempFile.Write(data)
		if err != nil {
			logger.Error("AOF Rewrite: Failed to write SELECT db", dbIndex, err)
			return err
		}
		// dump db
		rewriter.database.ForEach(dbIndex, func(key string, entity *databaseInterface.DataEntity, expiration *time.Time) bool {
			command := commandPool.Get().([]byte)[:0] // reset the buffer
			command = append(command, EntityToCommand(key, entity).ToBytes()...)
			if len(command) > 0 {
				_, _ = rewriter.tempFile.Write(command)
			}
			commandPool.Put(command)
			// TODO expiration
			return true
		})
	}
	return nil
}

func (rewriter *AofRewriter) TryLock() bool {
	return rewriter.rewriteMutex.TryLock()
}

func (rewriter *AofRewriter) Unlock() {
	rewriter.rewriteMutex.Unlock()
}
