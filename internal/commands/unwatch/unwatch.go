package unwatch

import (
	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func Execute(c *client.Client, args []interface{}, st *store.Store) ([]byte, error) {
	st.RemoveAllWatchers(c)
	return []byte("+OK\r\n"), nil
}
