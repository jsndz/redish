package multi

import (
	"errors"

	"github.com/jsndz/redish/internal/store"
)

func Execute(args []interface{}, st *store.Store) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("-ERR no args are supported")
	}
	return []byte("+OK\r\n"), nil
}
