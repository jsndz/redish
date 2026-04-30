package ping

import (
	"errors"
	"net"

	"github.com/jsndz/redish/internal/store"
)

func Execute(conn net.Conn, args []interface{}, _ *store.Store) error {
	if len(args) != 0 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	_, err := conn.Write([]byte("+PONG\r\n"))
	return err
}
