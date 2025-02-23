package database

import (
	"go-redis/config"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type Databases struct {
	databaseSet []*Database
}

// NewDatabases returns an new instance of Databases
func NewDatabases() *Databases {
	if config.Properties.Databases <= 0 {
		config.Properties.Databases = 16
	}
	databaseSet := make([]*Database, config.Properties.Databases)
	for i := range databaseSet {
		database := MakeDatabase()
		database.index = i
		databaseSet[i] = database
	}
	return &Databases{databaseSet: databaseSet}
}

func (dbs *Databases) Exec(client resp.Connection, args [][]byte) resp.Reply {
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
		return execSelect(client, dbs, args[1:])
	}
	dbIndex := client.GetDBIndex()
	return dbs.databaseSet[dbIndex].Exec(client, args)
}

// execSelect executes the select command
// SELECT index
func execSelect(connection resp.Connection, databases *Databases, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex >= len(databases.databaseSet) {
		return reply.MakeStandardErrorReply("ERR DB index is out of range")
	}
	connection.SelectDB(dbIndex)
	return reply.MakeOkReply()
}

func (dbs *Databases) Close() {
}

func (dbs *Databases) AfterClientClose(_ resp.Connection) {
}
