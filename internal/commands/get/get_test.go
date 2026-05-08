package get

import (
	"bufio"
	"net"
	"testing"

	"github.com/jsndz/redish/internal/client"
	"github.com/jsndz/redish/internal/store"
)

func TestGetExecute(t *testing.T) {
	s := store.New()

	t.Run("get existing key", func(t *testing.T) {
		key := "foo"
		value := "bar"
		s.Set(key, value, 0)

		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{key}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(cli)
		header, _ := reader.ReadString('\n')
		content, _ := reader.ReadString('\n')

		expectedHeader := "$3\r\n"
		expectedContent := "bar\r\n"
		if header != expectedHeader || content != expectedContent {
			t.Errorf("Expected %q%q, got %q%q", expectedHeader, expectedContent, header, content)
		}
	})

	t.Run("get non-existing key", func(t *testing.T) {
		cli, server := net.Pipe()
		defer cli.Close()
		defer server.Close()

		go func() {
			c := client.New(server)
			err := Execute(c, []interface{}{"nonexistent"}, s)
			if err != nil {
				t.Errorf("Execute returned error: %v", err)
			}
			server.Close()
		}()

		reader := bufio.NewReader(cli)
		response, _ := reader.ReadString('\n')

		expected := "$-1\r\n"
		if response != expected {
			t.Errorf("Expected %q, got %q", expected, response)
		}
	})

	t.Run("get wrong number of arguments", func(t *testing.T) {
		err := Execute(nil, []interface{}{}, s)
		if err == nil {
			t.Error("Expected error for zero arguments, got nil")
		}
	})
}
