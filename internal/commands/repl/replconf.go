package repl

import (
	"errors"

	"github.com/jsndz/redish/internal/store"
)

func ExecuteReplConf(args []interface{}, st *store.Store) ([]byte, error) {
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
	default:
		return nil, errors.New("-ERR Invalid command")
	}

}
