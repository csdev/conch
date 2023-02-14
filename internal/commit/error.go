package commit

import "strings"

type ParseError struct {
	Errors []string
}

func NewParseError() *ParseError {
	return &ParseError{
		Errors: []string{},
	}
}

func (e *ParseError) Error() string {
	return strings.Join(e.Errors, "\n")
}

func (e *ParseError) Append(err error) {
	e.Errors = append(e.Errors, err.Error())
}

func (e *ParseError) HasErrors() bool {
	return len(e.Errors) > 0
}
