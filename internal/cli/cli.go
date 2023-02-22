package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/csdev/conch/internal/util"
)

const StandardConfigFilename = "conch.yml"

// StandardConfigPath checks for a config file in the specified directory,
// and returns the path to it. If the config file does not exist, it returns
// an empty string.
func StandardConfigPath(dir string) (string, error) {
	p := filepath.Join(dir, StandardConfigFilename)
	_, err := os.Stat(p)
	if err == nil {
		// file exists
		return p, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		// file may exist, but some other error occurred
		return "", err
	}
	// file does not exist
	return "", nil
}

type Selections struct {
	Breaking      bool
	Minor         bool
	Patch         bool
	Uncategorized bool
}

func (s *Selections) Any() bool {
	return s.Breaking || s.Minor || s.Patch || s.Uncategorized
}

type Filters struct {
	Types  util.CaseInsensitiveSet
	Scopes util.CaseInsensitiveSet
	Selections
}

func (f *Filters) Any() bool {
	return f.Types != nil || f.Scopes != nil || f.Selections.Any()
}

type Outputs struct {
	List   bool
	Format string
	Count  bool
}

func (o *Outputs) Any() bool {
	return o.List || o.Format != "" || o.Count
}
