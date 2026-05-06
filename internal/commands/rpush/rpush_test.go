package rpush

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestRpushExecute(t *testing.T) {
	s := store.New()

	t.Run("rpush single value", func(t *testing.T) {
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()

		key := "mylist"
		value := "v1"
		go func() {
			err := Execute(server, []interface{}{key, value}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(client)
		response, _ := reader.ReadString('\n')

		expected := ":1\r\n"
		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	})

	t.Run("rpush multiple values", func(t *testing.T) {
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()

		key := "mylist"
		go func() {
			err := Execute(server, []interface{}{key, "v2", "v3"}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(client)
		response, _ := reader.ReadString('\n')

		expected := ":3\r\n" // Previous 1 + 2 new ones
		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	})

	t.Run("rpush wrong number of arguments", func(t *testing.T) {
		err := Execute(nil, []interface{}{"key"}, s)
		if err == nil {
			t.Error("Expected error for 1 argument, got nil")
		}
	})

	t.Run("rpush to wrong type", func(t *testing.T) {
		key := "stringkey"
		s.Set(key, "value", 0)

		err := Execute(nil, []interface{}{key, "v1"}, s)
		if err == nil {
			t.Error("Expected error for wrong type, got nil")
		}
	})
}
