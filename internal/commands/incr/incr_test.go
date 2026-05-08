package incr

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestIncr(t *testing.T) {
	s := store.New()

	t.Run("Key exists and has a numerical value", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()
		s.Set("test_Incr", "4", 100*time.Second)
		go func() {
			c := client.New(server)
			err := Execute(
				c,
				[]interface{}{"test_Incr"},
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
		expected := ":5\r\n"

		if string(resp) != expected {
			t.Errorf(
				"expected %q, got %q",
				expected,
				string(resp),
			)
		}
	})
	t.Run("Key does not exist", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()
		go func() {
			c := client.New(server)
			err := Execute(
				c,
				[]interface{}{"test_Incr2"},
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
		expected := ":1\r\n"

		if string(resp) != expected {
			t.Errorf(
				"expected %q, got %q",
				expected,
				string(resp),
			)
		}
	})
	t.Run("Key does but the value is not a number", func(t *testing.T) {

		s.Set("test_Incr3", "not_a_number", 100*time.Second)
		err := Execute(
			nil,
			[]interface{}{"test_Incr3"},
			s,
		)
		if err == nil {
			t.Error("Expected error got nil")
		}

	})
}
