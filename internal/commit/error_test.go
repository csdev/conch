package commit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		description string
		errorObject *ParseError
		expected    string
	}{
		{
			description: "empty object has empty error message",
			errorObject: &ParseError{
				Errors: []string{},
			},
			expected: "",
		},
		{
			description: "single error message is returned",
			errorObject: &ParseError{
				Errors: []string{"thing is broken"},
			},
			expected: "thing is broken",
		},
		{
			description: "multiple error messages are joined",
			errorObject: &ParseError{
				Errors: []string{"first thing is broken", "second thing is broken"},
			},
			expected: "first thing is broken\nsecond thing is broken",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, test.errorObject.Error())
		})
	}
}

func TestAppend(t *testing.T) {
	errorObject := NewParseError()
	errorObject.Append(errors.New("thing is broken"))
	assert.Equal(t, []string{"thing is broken"}, errorObject.Errors)
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		description string
		errorObject *ParseError
		expected    bool
	}{
		{
			description: "empty object has no errors",
			errorObject: &ParseError{
				Errors: []string{},
			},
			expected: false,
		},
		{
			description: "object with error has errors",
			errorObject: &ParseError{
				Errors: []string{"thing is broken"},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, test.errorObject.HasErrors())
		})
	}
}
