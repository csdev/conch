package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		description string
		document    string
		expected    CaseInsensitiveSet
	}{
		{
			description: "it decodes an empty set",
			document:    `MySet: []`,
			expected:    nil,
		},
		{
			description: "it decodes a set with items",
			document:    `MySet: ["A","b","c"]`,
			expected:    NewCaseInsensitiveSet([]string{"A", "b", "c"}),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			decoder := yaml.NewDecoder(strings.NewReader(test.document))
			decoder.KnownFields(true)

			var S struct {
				MySet CaseInsensitiveSet `yaml:"MySet"`
			}

			err := decoder.Decode(&S)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, S.MySet)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		description     string
		existingSet     CaseInsensitiveSet
		expectedPattern string
	}{
		{
			description:     "it returns an empty string if the set is nil",
			existingSet:     nil,
			expectedPattern: "^$",
		},
		{
			description:     "it returns an empty string if the set is empty",
			existingSet:     CaseInsensitiveSet{},
			expectedPattern: "^$",
		},
		{
			description:     "it returns the contents of the set in any order",
			existingSet:     NewCaseInsensitiveSet([]string{"asdf", "zxcv"}),
			expectedPattern: "^asdf,zxcv,|zxcv,asdf,$",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Regexp(t, test.expectedPattern, test.existingSet.String())
		})
	}
}

func TestSet(t *testing.T) {
	var s CaseInsensitiveSet
	s.Set("AAA,bbb")
	assert.Equal(t, NewCaseInsensitiveSet([]string{"AAA", "bbb"}), s)
}

func TestType(t *testing.T) {
	var s CaseInsensitiveSet
	assert.Equal(t, "comma_separated_strings", s.Type())
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

func TestAdd(t *testing.T) {
	tests := []struct {
		description string
		existingSet CaseInsensitiveSet
	}{
		{
			description: "it adds a new value",
			existingSet: CaseInsensitiveSet{},
		},
		{
			description: "it replaces an existing value",
			existingSet: NewCaseInsensitiveSet([]string{"foo"}),
		},
	}

	const newItem = "Foo"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.existingSet.Add(newItem)
			assert.True(t, test.existingSet.Contains(newItem))
			assert.Equal(t, newItem, test.existingSet.Value(newItem))
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		description string
		existingSet CaseInsensitiveSet
	}{
		{
			description: "it removes an existing item",
			existingSet: NewCaseInsensitiveSet([]string{"foo"}),
		},
		{
			description: "it is case insensitive",
			existingSet: NewCaseInsensitiveSet([]string{"Foo"}),
		},
		{
			description: "it ignores a missing item",
			existingSet: CaseInsensitiveSet{},
		},
	}

	const item = "foo"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.existingSet.Remove(item)
			assert.False(t, test.existingSet.Contains(item))
		})
	}
}
