package echo

import (
	"errors"
	"fmt"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, _ *store.Store) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("-ERR wrong number of arguments\r\n")
	}

	msg, ok := args[0].(string)
	if !ok {
		return nil, errors.New("-ERR invalid argument\r\n")
	}

	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)), nil
}
