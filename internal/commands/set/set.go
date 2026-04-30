package set

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jsndz/redish/internal/store"
)

func Execute(conn net.Conn, args []interface{}, st *store.Store) error {
	if len(args) != 2 && len(args) != 4 {
		return errors.New("-ERR wrong number of arguments\r\n")
	}

	key, ok := args[0].(string)
	if !ok {
		return errors.New("-ERR invalid key\r\n")
	}

	value, ok := args[1].(string)
	if !ok {
		return errors.New("-ERR invalid value\r\n")
	}

	var ttl time.Duration
	if len(args) == 4 {
		op, ok := args[2].(string)
		if !ok {
			return errors.New("-ERR invalid expiration option\r\n")
		}

		exp, ok := args[3].(string)
		if !ok {
			return errors.New("-ERR invalid expiration value\r\n")
		}

		parsed, err := strconv.Atoi(exp)
		if err != nil {
			return errors.New("-ERR invalid expiration value\r\n")
		}

		switch strings.ToUpper(op) {
		case "EX":
			ttl = time.Duration(parsed) * time.Second
		case "PX":
			ttl = time.Duration(parsed) * time.Millisecond
		default:
			return errors.New("-ERR syntax error\r\n")
		}
	}

	st.Set(key, value, ttl)
	_, err := conn.Write([]byte("+OK\r\n"))
	return err
}
