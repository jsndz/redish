package exec

import (
	"io"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/commands/multi"
	"github.com/jsndz/redish/internal/store"
)

func TestExec(t *testing.T) {
	s := store.New()
	t.Run("exec without calling multi gives error", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		c := client.New(server)
		err := Execute(
			c,
			[]interface{}{},
			s,
		)
		if err == nil {
			t.Error("Expected error got nil")
		}
	})
	t.Run("empty transaction", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		go func() {
			c := client.New(server)
			err := multi.Execute(c, []interface{}{}, s)
			if err != nil {
				t.Errorf("Execute returned error for multi: %v", err)
			}
			err = Execute(
				c,
				[]interface{}{},
				s,
			)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}

			server.Close()
		}()

		resp, err := io.ReadAll(cli)
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}
		expected := "+OK\r\n*0\r\n"

		if string(resp) != expected {
			t.Errorf(
				"expected %q, got %q",
				expected,
				string(resp),
			)
		}
	})
}
