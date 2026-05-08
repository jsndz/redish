package ping

import (
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestPingExecute(t *testing.T) {
	s := store.New()
	
	t.Run("ping with no arguments", func(t *testing.T) {
		response, err := Execute([]interface{}{}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "+PONG\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}
	})

	t.Run("ping with arguments", func(t *testing.T) {
		_, err := Execute([]interface{}{"extra"}, s)
		if err == nil {
			t.Error("Expected error for wrong number of arguments, got nil")
		}
	})
}
