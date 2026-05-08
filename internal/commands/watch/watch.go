package watch

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, st *store.Store) ([]byte, error) {
	if c.InTx {
		return nil, errors.New("-ERR Can't watch after transaction started")
	}
	key, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid key")
	}
	err := st.AddWatcher(key, c)
	if err != nil {
		return nil, errors.New("-ERR " + err.Error())
	}
	c.WatchedKeys[key] = true
	return []byte("+OK\r\n"), nil
}
