package ping

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, _ *store.Store) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}
	if c != nil && c.SubscribeMode {
		return []byte("*2\r\n$4\r\npong\r\n$0\r\n\r\n"), nil
	}
	return []byte("+PONG\r\n"), nil
}
