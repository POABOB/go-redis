package database

import (
	"go-redis/aof"
	"go-redis/config"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type DatabaseEngine struct {
	databaseSet []*Database
	aofHandler  *aof.AofHandler
}

// NewDatabaseEngine returns a new instance of DatabaseEngine
func NewDatabaseEngine() *DatabaseEngine {
	databaseEngine := &DatabaseEngine{}
	if config.Properties.Databases <= 0 {
		config.Properties.Databases = 16
	}
	databaseSet := make([]*Database, config.Properties.Databases)
	for i := range databaseSet {
		database := MakeDatabase()
		database.index = i
		databaseSet[i] = database
	}
	databaseEngine.databaseSet = databaseSet
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(databaseEngine)
		if err != nil {
			panic(err)
		}
		databaseEngine.aofHandler = aofHandler
		for _, database := range databaseEngine.databaseSet {
			index := database.index
			database.addAofFunc = func(commandLine CommandLine) {
				databaseEngine.aofHandler.AddAof(index, commandLine)
			}
		}
	}
	return databaseEngine
}

func (dbEngine *DatabaseEngine) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	commandName := strings.ToLower(string(args[0]))
	if commandName == "select" {
		if len(args) != 2 {
			return reply.MakeArgsNumErrorReply(commandName)
		}
		return execSelect(client, dbEngine, args[1:])
	}
	dbIndex := client.GetDBIndex()
	return dbEngine.databaseSet[dbIndex].Exec(client, args)
}

// execSelect executes the select command
// SELECT index
func execSelect(connection resp.Connection, dbEngine *DatabaseEngine, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex >= len(dbEngine.databaseSet) {
		return reply.MakeStandardErrorReply("ERR DB index is out of range")
	}
	connection.SelectDB(dbIndex)
	return reply.MakeOkReply()
}

func (dbEngine *DatabaseEngine) Close() {
}

func (dbEngine *DatabaseEngine) AfterClientClose(_ resp.Connection) {
}
