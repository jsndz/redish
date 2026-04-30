package commands

import (
	"errors"
	"net"
	"strings"

	"github.com/jsndz/redish/internal/commands/echo"
	"github.com/jsndz/redish/internal/commands/get"
	"github.com/jsndz/redish/internal/commands/ping"
	"github.com/jsndz/redish/internal/commands/set"
	"github.com/jsndz/redish/internal/store"
)

func Dispatch(conn net.Conn, arr []interface{}, st *store.Store) error {
	cmd, ok := arr[0].(string)
	if !ok {
		return errors.New("-ERR invalid command\r\n")
	}

	switch strings.ToUpper(cmd) {
	case "PING":
		return ping.Execute(conn, arr[1:], st)
	case "ECHO":
		return echo.Execute(conn, arr[1:], st)
	case "SET":
		return set.Execute(conn, arr[1:], st)
	case "GET":
		return get.Execute(conn, arr[1:], st)
	default:
		return errors.New("-ERR unknown command\r\n")
	}
}
