package ping

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestPingExecute(t *testing.T) {
	s := store.New()
	
	t.Run("ping with no arguments", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{}, s)
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

		expected := "+PONG\r\n"
		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	})

	t.Run("ping with arguments", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		c := client.New(server)
		err := Execute(c, []interface{}{"extra"}, s)
		if err == nil {
			t.Error("Expected error for wrong number of arguments, got nil")
		}
	})
}
