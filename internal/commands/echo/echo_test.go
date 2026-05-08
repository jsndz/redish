package echo

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestEchoExecute(t *testing.T) {
	s := store.New()

	t.Run("echo valid argument", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		msg := "hello"
		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{msg}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(cli)
		response, err := reader.ReadString('\n') // $5
		if err != nil {
			t.Fatalf("Failed to read response header: %v", err)
		}

		content, err := reader.ReadString('\n') // hello
		if err != nil {
			t.Fatalf("Failed to read response content: %v", err)
		}

		expectedHeader := "$5\r\n"
		expectedContent := "hello\r\n"
		if response != expectedHeader {
			t.Errorf("Expected header %q, got %q", expectedHeader, response)
		}
		if content != expectedContent {
			t.Errorf("Expected content %q, got %q", expectedContent, content)
		}
	})

	t.Run("echo wrong number of arguments", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		c := client.New(server)
		err := Execute(c, []interface{}{}, s)
		if err == nil {
			t.Error("Expected error for zero arguments, got nil")
		}

		err = Execute(c, []interface{}{"\"one\", \"two\""}, s)
		if err == nil {
			t.Error("Expected error for two arguments, got nil")
		}
	})

	t.Run("echo invalid argument type", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		c := client.New(server)
		err := Execute(c, []interface{}{123}, s)
		if err == nil {
			t.Error("Expected error for non-string argument, got nil")
		}
	})
}
