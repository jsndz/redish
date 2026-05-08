package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands/echo"
	"github.com/jsndz/redish/internal/commands/get"
	"github.com/jsndz/redish/internal/commands/incr"
	"github.com/jsndz/redish/internal/commands/lrange"
	"github.com/jsndz/redish/internal/commands/multi"
	"github.com/jsndz/redish/internal/commands/ping"
	"github.com/jsndz/redish/internal/commands/rpush"
	"github.com/jsndz/redish/internal/commands/set"
	"github.com/jsndz/redish/internal/store"
)

func Dispatch(c *client.Client, arr []interface{}, st *store.Store) ([]byte, error) {
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
		if !c.InTx {
			return nil, errors.New("-ERR EXEC without MULTI\r\n")
		}
		queue := c.TxQueue
		c.InTx = false
		c.TxQueue = nil
		res := []byte(fmt.Sprintf("*%d\r\n", len(queue)))
		for _, qCmd := range queue {
			qArr := append([]interface{}{qCmd.Name}, qCmd.Args...)
			qRes, err := Dispatch(c, qArr, st)
			if err != nil {
				res = append(res, []byte(err.Error())...)
			} else {
				res = append(res, qRes...)
			}
		}
		return res, nil
	default:
		return nil, errors.New("-ERR unknown command\r\n")
	}
}
