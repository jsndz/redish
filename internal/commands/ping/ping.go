package ping

import (
	"errors"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, _ *store.Store) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}

	return []byte("+PONG\r\n"), nil
}
