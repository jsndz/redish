package config

import (
	"flag"
	"sync"
)

type Config struct {
	Dir             string
	AppendOnly      bool
	AppendDirName   string
	AppendFileName  string
	AppendFsyncMode string

	mu sync.Mutex
}

func SetConfig() *Config {
	var cfg Config
	flag.StringVar(&cfg.Dir, "dir", ".", "data directory")

	flag.BoolVar(
		&cfg.AppendOnly,
		"appendonly",
		false,
		"enable append only mode",
	)

	flag.StringVar(
		&cfg.AppendDirName,
		"appenddirname",
		"appendonlydir",
		"append only directory name",
	)

	flag.StringVar(
		&cfg.AppendFileName,
		"appendfilename",
		"appendonly.aof",
		"append only file name",
	)

	flag.StringVar(
		&cfg.AppendFsyncMode,
		"appendfsync",
		"everysec",
		"fsync policy",
	)
	flag.Parse()

	return &cfg
}
