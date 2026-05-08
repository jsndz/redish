package multi

import (
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestMulti(t *testing.T) {
	s := store.New()

	t.Run("multi works", func(t *testing.T) {
		response, err := Execute([]interface{}{}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "+OK\r\n"
		if string(response) != expected {
			t.Errorf("expected %q, got %q", expected, string(response))
		}
	})
}
