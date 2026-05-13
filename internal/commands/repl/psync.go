package repl

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/server"
	"github.com/jsndz/redish/internal/store"
	"github.com/jsndz/redish/util"
)

func ExecutePsync(args []interface{}, st *store.Store, repl *server.Replication) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("-ERR invalid number of args")
	}
	replId, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid arg")
	}
	_, ok = args[1].(string)
	if !ok {
		return nil, errors.New("-ERR invalid arg")
	}
	rdbData := util.ReadRDB()

	if replId == "?" {
		resp := fmt.Sprintf(
			"+FULLRESYNC %s %d\r\n",
			repl.ReplID,
			repl.ReplOffset,
		)
		header := fmt.Sprintf("$%d\r\n", len(rdbData))
		final := append([]byte(resp), []byte(header)...)
		final = append(final, rdbData...)
		return final, nil
	} else if replId != repl.ReplID {
		resp := fmt.Sprintf(
			"+FULLRESYNC %s %d\r\n",
			repl.ReplID,
			repl.ReplOffset,
		)
		return []byte(resp), nil
	}

	return []byte("+CONTINUE\r\n"), nil
}
