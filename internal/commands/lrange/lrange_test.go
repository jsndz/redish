package lrange

import (
	"testing"

	"github.com/jsndz/redish/internal/store"
)

func TestLrangeExecute(t *testing.T) {
	s := store.New()
	s.Rpush("mylist", "one", "two", "three", "four", "five")

	response, err := Execute(
		[]interface{}{"mylist", "1", "2"},
		s,
	)

	if err != nil {
		t.Errorf("Execute returned error: %v", err)
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
