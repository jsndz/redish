package echo

import (
	"errors"
	"fmt"
	"net"

	"github.com/jsndz/redish/internal/store"
)

func Execute(conn net.Conn, args []interface{}, _ *store.Store) error {
	if len(args) != 1 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	msg, ok := args[0].(string)
	if !ok {
		return errors.New("-ERR invalid argument\r\n")
	}

	_, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)))
	return err
}
