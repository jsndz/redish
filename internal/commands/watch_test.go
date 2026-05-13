package commands

import (
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestWatchExec(t *testing.T) {
	s := store.New()

	cli1, _ := net.Pipe()
	c1 := client.New(cli1)

	cli2, _ := net.Pipe()
	c2 := client.New(cli2)

	t.Run("Optimistic locking works - transaction aborts", func(t *testing.T) {
		// Client 1 watches 'balance'
		res, err := Dispatch(c1, []interface{}{"WATCH", "balance"}, s, nil, nil)
		if err != nil {
			t.Fatalf("WATCH failed: %v", err)
		}
		if string(res) != "+OK\r\n" {
			t.Fatalf("Expected +OK for WATCH, got %q", string(res))
		}

		// Client 1 starts MULTI
		Dispatch(c1, []interface{}{"MULTI"}, s, nil, nil)

		// Client 1 queues SET balance 20
		Dispatch(c1, []interface{}{"SET", "balance", "20"}, s, nil, nil)

		// Client 2 modifies 'balance'
		_, err = Dispatch(c2, []interface{}{"SET", "balance", "30"}, s, nil, nil)
		if err != nil {
			t.Fatalf("Client 2 SET failed: %v", err)
		}

		// Client 1 executes transaction
		res, err = Dispatch(c1, []interface{}{"EXEC"}, s, nil, nil)
		if err != nil {
			t.Fatalf("EXEC failed: %v", err)
		}

		expected := "*0\r\n"
		if string(res) != expected {
			t.Errorf("Expected aborted transaction %q, got %q", expected, string(res))
		}

		res, _ = Dispatch(c1, []interface{}{"GET", "balance"}, s, nil, nil)
		if string(res) != "$2\r\n30\r\n" {
			t.Errorf("Expected balance to be 30, got %q", string(res))
		}
	})

	t.Run("WATCH followed by successful EXEC", func(t *testing.T) {
		s := store.New()
		cli1, _ := net.Pipe()
		c1 := client.New(cli1)

		Dispatch(c1, []interface{}{"WATCH", "foo"}, s, nil, nil)
		Dispatch(c1, []interface{}{"MULTI"}, s, nil, nil)
		Dispatch(c1, []interface{}{"SET", "foo", "bar"}, s, nil, nil)

		res, err := Dispatch(c1, []interface{}{"EXEC"}, s, nil, nil)
		if err != nil {
			t.Fatalf("EXEC failed: %v", err)
		}

		expected := "*1\r\n+OK\r\n"
		if string(res) != expected {
			t.Errorf("Expected successful transaction %q, got %q", expected, string(res))
		}

		res, _ = Dispatch(c1, []interface{}{"GET", "foo"}, s, nil, nil)
		if string(res) != "$3\r\nbar\r\n" {
			t.Errorf("Expected foo to be bar, got %q", string(res))
		}
	})

	t.Run("WATCH multiple keys - one changes", func(t *testing.T) {
		s := store.New()
		cli1, _ := net.Pipe()
		c1 := client.New(cli1)
		cli2, _ := net.Pipe()
		c2 := client.New(cli2)

		Dispatch(c1, []interface{}{"WATCH", "key1"}, s, nil, nil)
		Dispatch(c1, []interface{}{"WATCH", "key2"}, s, nil, nil)

		Dispatch(c1, []interface{}{"MULTI"}, s, nil, nil)
		Dispatch(c1, []interface{}{"SET", "key1", "val1"}, s, nil, nil)

		Dispatch(c2, []interface{}{"SET", "key2", "changed"}, s, nil, nil)

		res, err := Dispatch(c1, []interface{}{"EXEC"}, s, nil, nil)
		if err != nil {
			t.Fatalf("EXEC failed: %v", err)
		}

		if string(res) != "*0\r\n" {
			t.Errorf("Expected aborted transaction, got %q", string(res))
		}
	})
}
