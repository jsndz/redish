package set

import (
	"testing"
	"time"

	"github.com/jsndz/redish/internal/store"
)

func TestSetExecute(t *testing.T) {
	s := store.New()

	t.Run("set valid arguments", func(t *testing.T) {
		key := "foo"
		value := "bar"
		response, err := Execute([]interface{}{key, value}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "+OK\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}

		got, ok := s.Get(key)
		if !ok || got != value {
			t.Errorf("Expected store to have %q for key %q, got %q (ok=%v)", value, key, got, ok)
		}
	})

	t.Run("set with EX expiration", func(t *testing.T) {
		key := "ex_key"
		value := "ex_val"
		response, err := Execute([]interface{}{key, value, "EX", "1"}, s)
		if err != nil {
			t.Errorf("Execute returned error: %v", err)
		}

		expected := "+OK\r\n"
		if string(response) != expected {
			t.Errorf("Expected %q, got %q", expected, string(response))
		}

		got, ok := s.Get(key)
		if !ok || got != value {
			t.Errorf("Expected store to have %q, got %q", value, got)
		}

		time.Sleep(1100 * time.Millisecond)
		_, ok = s.Get(key)
		if ok {
			t.Error("Expected key to be expired")
		}
	})

	t.Run("set wrong number of arguments", func(t *testing.T) {
		_, err := Execute([]interface{}{"key"}, s)
		if err == nil {
			t.Error("Expected error for 1 argument, got nil")
		}

		_, err = Execute([]interface{}{"key", "val", "EX"}, s)
		if err == nil {
			t.Error("Expected error for 3 arguments, got nil")
		}
	})
}
