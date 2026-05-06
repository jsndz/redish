package echo

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestEchoExecute(t *testing.T) {
	s := store.New()

	t.Run("echo valid argument", func(t *testing.T) {
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()

		msg := "hello"
		go func() {
			err := Execute(server, []interface{}{msg}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(client)
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
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()

		err := Execute(server, []interface{}{}, s)
		if err == nil {
			t.Error("Expected error for zero arguments, got nil")
		}

		err = Execute(server, []interface{}{"\"one\", \"two\""}, s)
		if err == nil {
			t.Error("Expected error for two arguments, got nil")
		}
	})

	t.Run("echo invalid argument type", func(t *testing.T) {
		client, server := net.Pipe()
		defer client.Close()
		defer server.Close()

		err := Execute(server, []interface{}{123}, s)
		if err == nil {
			t.Error("Expected error for non-string argument, got nil")
		}
	})
}
