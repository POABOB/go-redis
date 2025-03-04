package database

import (
	"go-redis/aof"
	"go-redis/config"
	databaseInterface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
	"time"
)

type StandaloneDatabase struct {
	dictEntity []*DictEntity
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase returns a new instance of StandaloneDatabase
func NewStandaloneDatabase() *StandaloneDatabase {
	databaseEngine := &StandaloneDatabase{}
	if config.Properties.Databases <= 0 {
		config.Properties.Databases = 16
	}
	dictEntity := make([]*DictEntity, config.Properties.Databases)
	for i := range dictEntity {
		database := MakeDatabase()
		database.index = i
		dictEntity[i] = database
	}
	databaseEngine.dictEntity = dictEntity
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(databaseEngine)
		if err != nil {
			panic(err)
		}
		databaseEngine.aofHandler = aofHandler
		for _, database := range databaseEngine.dictEntity {
			index := database.index
			database.addAofFunc = func(commandLine databaseInterface.CommandLine) {
				databaseEngine.aofHandler.AddAof(index, commandLine)
			}
		}
	}
	return databaseEngine
}

func (database *StandaloneDatabase) Exec(client resp.Connection, args databaseInterface.CommandLine) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	commandName := strings.ToLower(string(args[0]))
	// authenticate
	if !isAuthenticated(client) {
		return reply.MakeStandardErrorReply("NOAUTH Authentication required")
	}

	if commandName == "select" {
		if len(args) != 2 {
			return reply.MakeArgsNumErrorReply(commandName)
		}
		return execSelect(client, database, args[1:])
	}

	dbIndex := client.GetDBIndex()
	return database.dictEntity[dbIndex].Exec(client, args)
}

func (database *StandaloneDatabase) ForEach(dbIndex int, cb func(key string, data *databaseInterface.DataEntity, expiration *time.Time) bool) {
	if dbIndex >= len(database.dictEntity) || dbIndex < 0 {
		logger.Error("invalid db index")
		return
	}
	database.dictEntity[dbIndex].ForEach(cb)
}

// execSelect executes the select commands
// SELECT index
func execSelect(connection resp.Connection, database *StandaloneDatabase, args databaseInterface.CommandLine) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dictEntity) {
		return reply.MakeStandardErrorReply("ERR DB index is out of range")
	}
	connection.SelectDB(dbIndex)
	return reply.MakeOkReply()
}

// Close closes the aof handler gracefully
func (database *StandaloneDatabase) Close() {
	// graceful shutdown
	if database.aofHandler != nil {
		database.aofHandler.Close()
	}
}

func (database *StandaloneDatabase) AfterClientClose(_ resp.Connection) {
}
