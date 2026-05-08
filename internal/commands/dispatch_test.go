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
	t.Run("Dispatch DISCARD WITH MULTI", func(t *testing.T) {
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
		res, _ = Dispatch(c, []interface{}{"DISCARD"}, s)
		expected := "+OK\r\n"
		if string(res) != expected {
			t.Errorf("Expected %q, got %q", expected, string(res))
		}
		if c.InTx {
			t.Error("Expected client NOT to be in transaction")
		}
	})
	t.Run("Dispatch DISCARD WITHOUT MULTI", func(t *testing.T) {

		_, err := Dispatch(c, []interface{}{"DISCARD"}, s)
		if err == nil {
			t.Error("Expected err")
		}
	})
	t.Run("EXEC does not rollback on command error", func(t *testing.T) {
		res, _ := Dispatch(c, []interface{}{"MULTI"}, s)

		if string(res) != "+OK\r\n" {
			t.Fatalf("expected +OK, got %q", string(res))
		}

		Dispatch(c, []interface{}{"SET", "foo", "bar"}, s)

		Dispatch(c, []interface{}{"GET"}, s)

		res, _ = Dispatch(c, []interface{}{"EXEC"}, s)

		expected := "*2\r\n+OK\r\n-ERR wrong number of arguments\r\n"

		if string(res) != expected {
			t.Errorf("expected %q, got %q", expected, string(res))
		}

		res, err := Dispatch(c, []interface{}{"GET", "foo"}, s)
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}

		if string(res) != "$3\r\nbar\r\n" {
			t.Errorf("expected value to persist, got %q", string(res))
		}
	})
}
