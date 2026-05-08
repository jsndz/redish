package get

import (
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestGetExecute(t *testing.T) {
	s := store.New()

	t.Run("get existing key", func(t *testing.T) {
		key := "foo"
		value := "bar"
		s.Set(key, value, 0)

		response, err := Execute([]interface{}{key}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "$3\r\nbar\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}
	})

	t.Run("get non-existing key", func(t *testing.T) {
		response, err := Execute([]interface{}{"nonexistent"}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "$-1\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}
	})

	t.Run("get wrong number of arguments", func(t *testing.T) {
		_, err := Execute([]interface{}{}, s)
		if err == nil {
			t.Error("Expected error for zero arguments, got nil")
		}
	})
}
