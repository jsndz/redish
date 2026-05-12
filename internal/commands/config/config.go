package config

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func Execute(c *client.Client, args []interface{}, st *store.Store, cfg *util.Config) ([]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("-ERR wrong number of arguments for 'CONFIG' command\r\n")
	}

	subCmd, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid subcommand for 'CONFIG' command\r\n")
	}

	switch subCmd {
	case "GET":
		return GetConfig(c, args[1:], st, cfg)
	default:
		return nil, errors.New("-ERR unknown subcommand for 'CONFIG' command\r\n")
	}
}

func GetConfig(c *client.Client, args []interface{}, st *store.Store, cfg *util.Config) ([]byte, error) {
	arg, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid request for GET\r\n")
	}
	switch arg {
	case "dir":
		dir := cfg.Dir

		return []byte(fmt.Sprintf("*2\r\n$3\r\ndir\r\n$%d\r\n%s\r\n", len(dir), dir)), nil
	case "appendonly":
		value := "no"
		if cfg.AppendOnly {
			value = "yes"
		}

		return []byte(
			fmt.Sprintf("*2\r\n$10\r\nappendonly\r\n$%d\r\n%s\r\n", len(value), value),
		), nil

	case "appenddirname":
		value := cfg.AppendDirName

		return []byte(
			fmt.Sprintf("*2\r\n$15\r\nappenddirname\r\n$%d\r\n%s\r\n", len(value), value),
		), nil

	case "appendfilename":
		value := cfg.AppendFileName

		return []byte(
			fmt.Sprintf("*2\r\n$14\r\nappendfilename\r\n$%d\r\n%s\r\n", len(value), value),
		), nil

	case "appendfsync":
		value := cfg.AppendFsyncMode

		return []byte(
			fmt.Sprintf("*2\r\n$12\r\nappendfsync\r\n$%d\r\n%s\r\n", len(value), value),
		), nil
	}
	return nil, errors.New("-ERR unknown configuration parameter\r\n")
}
