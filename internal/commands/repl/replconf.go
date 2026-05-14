package repl

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jsndz/redish/internal/core"
	"github.com/jsndz/redish/internal/store"
)

func ExecuteReplConf(args []interface{}, st *store.Store, repl *core.Replication) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("-ERR invalid number of args")
	}
	subcmd, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid arg")
	}

	switch subcmd {
	case "listening-port":
		return []byte("+OK\r\n"), nil
	case "capa":
		return []byte("+OK\r\n"), nil
	case "GETACK":
		replOffsetStr := strconv.FormatInt(repl.ReplOffset, 10)
		return []byte(fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%s\r\n", len(replOffsetStr), replOffsetStr)), nil
	default:
		return nil, errors.New("-ERR Invalid command")
	}

}
