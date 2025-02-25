package main

import (
	"fmt"
	"go-redis/config"
	"go-redis/lib/logger"
	"go-redis/resp/handler"
	"go-redis/tcp"
	"os"
)

const configFile = "redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind:                       "127.0.0.1",
	Port:                       6379,
	AppendOnly:                 false,
	AppendFilename:             "appendOnly.aof",
	AppendFsync:                "everysec",
	AutoAofRewriteMinSize:      "64mb",
	AutoAofRewritePercentage:   100,
	AofRewriteIncrementalFsync: true,
	NoAppendFsyncOnRewrite:     false,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	logger.Setup(&logger.Settings{
		Path:       "./logs",
		Name:       "go-redis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{Addr: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port)},
		handler.MakeHandler(),
	)
	if err != nil {
		logger.Error(err)
	}
}
