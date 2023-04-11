// Package cli defines the command-line features of conch.
package cli

import (
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/csdev/conch/internal/util"
)

// Selections are the different ways commits can be included based on impact.
type Selections struct {
	Breaking      bool
	Minor         bool
	Patch         bool
	Uncategorized bool
}

func (s *Selections) Any() bool {
	return s.Breaking || s.Minor || s.Patch || s.Uncategorized
}

// Filters are the different ways commits can be included based on their
// attributes or impact.
type Filters struct {
	Types  util.CaseInsensitiveSet
	Scopes util.CaseInsensitiveSet
	Selections
}

func (f *Filters) Any() bool {
	return f.Types != nil || f.Scopes != nil || f.Selections.Any()
}

// Outputs are the different ways that commit information can be displayed
// to the user on the command line.
type Outputs struct {
	List        bool
	Format      string
	Count       bool
	Impact      bool
	BumpVersion string
}

func (o *Outputs) Any() bool {
	return o.List || o.Format != "" || o.Count || o.Impact || o.BumpVersion != ""
}

// Template creates a new text template with the specified name and contents,
// suitable for formatting CLI output.
func Template(name string, contents string) (*template.Template, error) {
	c := strings.NewReplacer(`\\`, `\`, `\t`, "\t", `\n`, "\n").Replace(contents)
	return template.New(name).Parse(c)
}

func GetFileContents(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
