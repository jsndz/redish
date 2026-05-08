package ping

import (
	"errors"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, _ *store.Store) error {
	if len(args) != 0 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	_, err := c.Write([]byte("+PONG\r\n"))
	return err
}
