package set

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestSetExecute(t *testing.T) {
	s := store.New()

	t.Run("set valid arguments", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		key := "foo"
		value := "bar"
		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{key, value}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(cli)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		expected := "+OK\r\n"
		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}

		got, ok := s.Get(key)
		if !ok || got != value {
			t.Errorf("Expected store to have %q for key %q, got %q (ok=%v)", value, key, got, ok)
		}
	})

	t.Run("set with EX expiration", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		key := "ex_key"
		value := "ex_val"
		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{key, value, "EX", "1"}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(cli)
		reader.ReadString('\n') // Consume +OK

		got, ok := s.Get(key)
		if !ok || got != value {
			t.Errorf("Expected store to have %q, got %q", value, got)
		}

		time.Sleep(1100 * time.Millisecond)
		_, ok = s.Get(key)
		if ok {
			t.Error("Expected key to be expired")
		}
	})

	t.Run("set wrong number of arguments", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		c := client.New(server)
		err := Execute(c, []interface{}{"key"}, s)
		if err == nil {
			t.Error("Expected error for 1 argument, got nil")
		}

		err = Execute(c, []interface{}{"key", "val", "EX"}, s)
		if err == nil {
			t.Error("Expected error for 3 arguments, got nil")
		}
	})
}
