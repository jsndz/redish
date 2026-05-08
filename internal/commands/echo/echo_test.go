package echo

import (
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestEchoExecute(t *testing.T) {
	s := store.New()

	t.Run("echo valid argument", func(t *testing.T) {
		msg := "hello"
		response, err := Execute([]interface{}{msg}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "$5\r\nhello\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}
	})

	t.Run("echo wrong number of arguments", func(t *testing.T) {
		_, err := Execute([]interface{}{}, s)
		if err == nil {
			t.Error("Expected error for zero arguments, got nil")
		}

		_, err = Execute([]interface{}{"one", "two"}, s)
		if err == nil {
			t.Error("Expected error for two arguments, got nil")
		}
	})

	t.Run("echo invalid argument type", func(t *testing.T) {
		_, err := Execute([]interface{}{123}, s)
		if err == nil {
			t.Error("Expected error for non-string argument, got nil")
		}
	})
}
