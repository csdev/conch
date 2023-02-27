package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaseInsensitiveSet(t *testing.T) {
	s := NewCaseInsensitiveSet([]string{"foo", "Bar"})

	tests := []struct {
		description string
		lookup      string
		contains    bool
		value       string
	}{
		{
			description: "it does not find a missing value",
			lookup:      "asdf",
			contains:    false,
			value:       "",
		},
		{
			description: "it finds a contained value",
			lookup:      "foo",
			contains:    true,
			value:       "foo",
		},
		{
			description: "the input is case insensitive",
			lookup:      "Foo",
			contains:    true,
			value:       "foo",
		},
		{
			description: "the set values are case insensitive",
			lookup:      "bar",
			contains:    true,
			value:       "Bar",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.contains, s.Contains(test.lookup))
			assert.Equal(t, test.value, s.Value(test.lookup))
		})
	}
}

func TestCopy(t *testing.T) {
	s := NewCaseInsensitiveSet([]string{"foo", "Bar"})
	s2 := s.Copy()
	s2.Remove("foo")

	// original set was not modified
	assert.True(t, s.Contains("foo"))
	assert.True(t, s.Contains("Bar"))

	// copy was modified
	assert.False(t, s2.Contains("foo"))
	assert.True(t, s2.Contains("Bar"))
}
