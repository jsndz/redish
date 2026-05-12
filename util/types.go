package util

import "flag"

type Config struct {
	Dir             string
	AppendOnly      bool
	AppendDirName   string
	AppendFileName  string
	AppendFsyncMode string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.Dir, "dir", "", "data directory")
	flag.BoolVar(&cfg.AppendOnly, "appendonly", false, "enable AOF")
	flag.StringVar(&cfg.AppendDirName, "appenddirname", "", "append dir name")
	flag.StringVar(&cfg.AppendFileName, "appendfilename", "", "append file name")
	flag.StringVar(&cfg.AppendFsyncMode, "appendfsync", "everysec", "fsync mode")

	flag.Parse()

	return cfg
}
