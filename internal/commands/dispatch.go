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
	replSub "github.com/jsndz/redish/internal/commands/repl"
	"github.com/jsndz/redish/internal/commands/rpush"
	"github.com/jsndz/redish/internal/commands/set"
	"github.com/jsndz/redish/internal/commands/unwatch"
	"github.com/jsndz/redish/internal/commands/watch"
	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
)

type CommandHandler func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error)

var registry = make(map[string]CommandHandler)

func Register(name string, handler CommandHandler) {
	registry[strings.ToUpper(name)] = handler
}

func init() {
	Register("PING", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return ping.Execute(args, st)
	})
	Register("ECHO", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return echo.Execute(args, st)
	})
	Register("SET", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return set.Execute(args, st)
	})
	Register("GET", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return get.Execute(args, st)
	})
	Register("INCR", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return incr.Execute(args, st)
	})
	Register("RPUSH", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return rpush.Execute(args, st)
	})
	Register("LRANGE", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return lrange.Execute(args, st)
	})
	Register("MULTI", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		c.InTx = true
		return multi.Execute(args, st)
	})
	Register("EXEC", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return exec.Execute(c, args, st, cfg, repl, Dispatch)
	})
	Register("DISCARD", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return discard.Execute(c, args, st)
	})
	Register("WATCH", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return watch.Execute(c, args, st)
	})
	Register("UNWATCH", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return unwatch.Execute(c, args, st)
	})
	Register("CONFIG", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return getconfig.Execute(c, args, st, cfg)
	})
	Register("REPLCONF", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return replSub.ExecuteReplConf(args, st, repl)
	})
	Register("PSYNC", func(c *client.Client, args []interface{}, st *store.Store, cfg *config.Config, repl *core.Replication) ([]byte, error) {
		return replSub.ExecutePsync(c, args, st, repl)
	})
}

func Dispatch(c *client.Client, arr []interface{}, st *store.Store, cfg *config.Config, replication *core.Replication) ([]byte, error) {
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

	if handler, ok := registry[cmd.Name]; ok {
		return handler(c, cmd.Args, st, cfg, replication)
	}

	return nil, errors.New("-ERR unknown command\r\n")
}
