package util

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// CaseInsensitiveSet is a mapping of lowercase strings to the original
// casing of those strings. It allows membership tests on the lowercase keys,
// while preserving the original values.
type CaseInsensitiveSet map[string]string

func NewCaseInsensitiveSet(items []string) CaseInsensitiveSet {
	m := make(CaseInsensitiveSet)
	for _, item := range items {
		m[strings.ToLower(item)] = item
	}
	return m
}

func (s *CaseInsensitiveSet) UnmarshalYAML(value *yaml.Node) error {
	var rawItems []string
	err := value.Decode(&rawItems)
	if err != nil {
		return err
	}

	if len(rawItems) > 0 {
		*s = NewCaseInsensitiveSet(rawItems)
	}
	return nil
}

// String implements pflag.Value.String, which prints the contents of the
// collection for use with command-line flags.
func (s *CaseInsensitiveSet) String() string {
	if s == nil {
		return ""
	}
	b := strings.Builder{}
	for _, item := range *s {
		b.WriteString(item)
		b.WriteString(",")
	}
	return b.String()
}

// Set implements pflag.Value.Set, which sets the new value of the collection
// from command-line flags.
func (s *CaseInsensitiveSet) Set(val string) error {
	*s = NewCaseInsensitiveSet(strings.Split(val, ","))
	return nil
}

// Type implements pflag.Value.Type, which returns a description of the
// flag type for display in command-line help.
func (s *CaseInsensitiveSet) Type() string {
	return "comma_separated_strings"
}

func (s CaseInsensitiveSet) Add(item string) {
	key := strings.ToLower(item)
	s[key] = item
}

func (s CaseInsensitiveSet) Contains(item string) bool {
	key := strings.ToLower(item)
	_, ok := s[key]
	return ok
}

func (s CaseInsensitiveSet) Value(item string) string {
	key := strings.ToLower(item)
	return s[key]
}
