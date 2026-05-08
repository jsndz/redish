package commands

import (
	"errors"
	"strings"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands/echo"
	"github.com/jsndz/redish/internal/commands/get"
	"github.com/jsndz/redish/internal/commands/lrange"
	"github.com/jsndz/redish/internal/commands/multi"
	"github.com/jsndz/redish/internal/commands/ping"
	"github.com/jsndz/redish/internal/commands/rpush"
	"github.com/jsndz/redish/internal/commands/set"
	"github.com/jsndz/redish/internal/store"
)

func Dispatch(c *client.Client, arr []interface{}, st *store.Store) error {
	cmd, ok := arr[0].(string)
	if !ok {
		return errors.New("-ERR invalid command\r\n")
	}

	switch strings.ToUpper(cmd) {
	case "PING":
		return ping.Execute(c, arr[1:], st)
	case "ECHO":
		return echo.Execute(c, arr[1:], st)
	case "SET":
		return set.Execute(c, arr[1:], st)
	case "GET":
		return get.Execute(c, arr[1:], st)
	case "RPUSH":
		return rpush.Execute(c, arr[1:], st)
	case "LRANGE":
		return lrange.Execute(c, arr[1:], st)
	case "MULTI":
		return multi.Execute(c, arr[1:], st)

	default:
		return errors.New("-ERR unknown command\r\n")
	}
}
