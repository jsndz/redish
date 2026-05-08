package multi

import (
	"io"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestMulti(t *testing.T) {
	s := store.New()

	t.Run("multi works", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{
				[]interface{}{"MULTI"},
			}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		resp, err := io.ReadAll(cli)
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}
		expected := "+OK\r\n"

		if string(resp) != expected {
			t.Errorf(
				"expected %q, got %q",
				expected,
				string(resp),
			)
		}
	})
}
