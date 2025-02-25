package config

import (
	"bufio"
	"go-redis/lib/logger"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ServerProperties define global config properties
type ServerProperties struct {
	Bind        string `cfg:"bind"`
	Port        int    `cfg:"port"`
	MaxClients  int    `cfg:"maxclients"`
	RequirePass bool   `cfg:"requirepass"`
	Databases   int    `cfg:"databases"`

	AppendOnly                 bool   `cfg:"appendonly"`
	AppendFilename             string `cfg:"appendfilename"`
	AppendFsync                string `cfg:"appendfsync"`
	AutoAofRewriteMinSize      string `cfg:"auto-aof-rewrite-min-size"`
	AutoAofRewritePercentage   int64  `cfg:"auto-aof-rewrite-percentage"`
	AofRewriteIncrementalFsync bool   `cfg:"aof-rewrite-incremental-fsync"`
	NoAppendFsyncOnRewrite     bool   `cfg:"no-appendfsync-on-rewrite"`

	Peers []string `cfg:"peers"`
	Self  string   `cfg:"self"`
}

// Properties global config properties
var Properties *ServerProperties

func init() {
	// default config
	Properties = &ServerProperties{
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
}

func parse(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	// read config file
	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && strings.TrimLeft(line, " ")[0] == '#' {
			continue
		}
		pivot := strings.IndexAny(line, " ")
		if pivot > 0 && pivot < len(line)-1 { // separator found
			key := line[0:pivot]
			value := strings.Trim(line[pivot+1:], " ")
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	// parse format
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok || strings.TrimLeft(key, " ") == "" {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			// fill config
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				boolValue := "yes" == value
				fieldVal.SetBool(boolValue)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					fieldVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	return config
}

// SetupConfig read config file and store properties into Properties
func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()
	Properties = parse(file)
}
