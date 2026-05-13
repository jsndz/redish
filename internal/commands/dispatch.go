package commands

import (
	"errors"
	"strings"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands/discard"
	"github.com/jsndz/redish/internal/commands/echo"
	"github.com/jsndz/redish/internal/commands/exec"
	"github.com/jsndz/redish/internal/commands/get"
	"github.com/jsndz/redish/internal/commands/getconfig"
	"github.com/jsndz/redish/internal/commands/incr"
	"github.com/jsndz/redish/internal/commands/lrange"
	"github.com/jsndz/redish/internal/commands/multi"
	"github.com/jsndz/redish/internal/commands/ping"
	"github.com/jsndz/redish/internal/commands/repl"
	"github.com/jsndz/redish/internal/commands/rpush"
	"github.com/jsndz/redish/internal/commands/set"
	"github.com/jsndz/redish/internal/commands/unwatch"
	"github.com/jsndz/redish/internal/commands/watch"
	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/internal/server"
	"github.com/jsndz/redish/internal/store"
)

func Dispatch(c *client.Client, arr []interface{}, st *store.Store, cfg *config.Config, replication *server.Replication) ([]byte, error) {
	cmdName, ok := arr[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid command\r\n")
	}
	cmd := client.Command{
		Name: strings.ToUpper(cmdName),
		Args: arr[1:],
	}

	if c.InTx &&
		cmd.Name != "EXEC" &&
		cmd.Name != "DISCARD" &&
		cmd.Name != "MULTI" {

		c.TxQueue = append(c.TxQueue, cmd)
		return []byte("+QUEUED\r\n"), nil
	}

	switch cmd.Name {
	case "PING":
		return ping.Execute(cmd.Args, st)
	case "ECHO":
		return echo.Execute(cmd.Args, st)
	case "SET":
		return set.Execute(cmd.Args, st)
	case "GET":
		return get.Execute(cmd.Args, st)
	case "INCR":
		return incr.Execute(cmd.Args, st)
	case "RPUSH":
		return rpush.Execute(cmd.Args, st)
	case "LRANGE":
		return lrange.Execute(cmd.Args, st)
	case "MULTI":
		c.InTx = true
		return multi.Execute(cmd.Args, st)
	case "EXEC":
		return exec.Execute(c, cmd.Args, st, cfg, replication, Dispatch)
	case "DISCARD":
		return discard.Execute(c, cmd.Args, st)
	case "WATCH":
		return watch.Execute(c, cmd.Args, st)
	case "UNWATCH":
		return unwatch.Execute(c, cmd.Args, st)
	case "CONFIG":
		return getconfig.Execute(c, cmd.Args, st, cfg)
	case "REPLCONF":
		return repl.ExecuteReplConf(cmd.Args, st)
	case "PSYNC":
		return repl.ExecutePsync(cmd.Args, st, replication)
	default:
		return nil, errors.New("-ERR unknown command\r\n")
	}
}
