package lrange

import (
	"io"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestLrangeExecute(t *testing.T) {
	s := store.New()
	s.Rpush("mylist", "one", "two", "three", "four", "five")

	cli, server := net.Pipe()
	defer cli.Close()
	defer server.Close()

	go func() {
		c := client.New(server)
		err := Execute(
			c,
			[]interface{}{"mylist", "1", "2"},
			s,
		)

		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		server.Close()
	}()

	response, err := io.ReadAll(cli)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	expected := "*2\r\n$3\r\ntwo\r\n$5\r\nthree\r\n"

	if string(response) != expected {
		t.Errorf(
			"expected %q, got %q",
			expected,
			string(response),
		)
	}
}
