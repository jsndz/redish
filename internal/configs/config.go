package configs

import (
	"flag"
)

type Config struct {
	Dir             string
	AppendOnly      bool
	AppendDirName   string
	AppendFileName  string
	AppendFsyncMode string
}

func NewConfig() *Config {
	return &Config{
		Dir:             ".",
		AppendOnly:      false,
		AppendDirName:   "appendonlydir",
		AppendFileName:  "appendonly.aof",
		AppendFsyncMode: "everysec",
	}
}

func (cfg *Config) SetConfig() {

	flag.StringVar(&cfg.Dir, "dir", cfg.Dir, "data directory")
	flag.BoolVar(&cfg.AppendOnly, "appendonly", cfg.AppendOnly, "enable AOF")
	flag.StringVar(&cfg.AppendDirName, "appenddirname", cfg.AppendDirName, "append dir name")
	flag.StringVar(&cfg.AppendFileName, "appendfilename", cfg.AppendFileName, "append file name")
	flag.StringVar(&cfg.AppendFsyncMode, "appendfsync", cfg.AppendFsyncMode, "fsync mode")

	flag.Parse()

}
