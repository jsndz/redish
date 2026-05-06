package lrange

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestLrangeExecute(t *testing.T) {
	s := store.New()
	s.Rpush("mylist", "one", "two", "three", "four", "five")

	t.Run("lrange with valid args", func(t *testing.T) {
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()
		key := "test_lrange"
		go func() {
			err := Execute(server, []interface{}{key, 1, 2}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()
		reader := bufio.NewReader(client)
		response, _ := reader.ReadString('\n')

		expected := "*2\r\n$1\r\none\r\n$1\r\ntwo\r\n"

		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	})
}
