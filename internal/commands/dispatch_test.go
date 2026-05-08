package commands

import (
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestDispatch(t *testing.T) {
	s := store.New()
	cli, _ := net.Pipe()
	c := client.New(cli)

	t.Run("Dispatch PING", func(t *testing.T) {
		res, err := Dispatch(c, []interface{}{"PING"}, s)
		if err != nil {
			t.Errorf("Dispatch returned error: %v", err)
		}
		expected := "+PONG\r\n"
		if string(res) != expected {
			t.Errorf("Expected %q, got %q", expected, string(res))
		}
	})

	t.Run("Dispatch MULTI and EXEC", func(t *testing.T) {
		res, _ := Dispatch(c, []interface{}{"MULTI"}, s)
		if string(res) != "+OK\r\n" {
			t.Errorf("Expected +OK for MULTI, got %q", string(res))
		}
		if !c.InTx {
			t.Error("Expected client to be in transaction")
		}

		res, _ = Dispatch(c, []interface{}{"SET", "k1", "v1"}, s)
		if string(res) != "+QUEUED\r\n" {
			t.Errorf("Expected +QUEUED, got %q", string(res))
		}

		res, _ = Dispatch(c, []interface{}{"GET", "k1"}, s)
		if string(res) != "+QUEUED\r\n" {
			t.Errorf("Expected +QUEUED, got %q", string(res))
		}

		res, _ = Dispatch(c, []interface{}{"EXEC"}, s)
		expected := "*2\r\n+OK\r\n$2\r\nv1\r\n"
		if string(res) != expected {
			t.Errorf("Expected %q, got %q", expected, string(res))
		}
		if c.InTx {
			t.Error("Expected client NOT to be in transaction")
		}
	})
}
