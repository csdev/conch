package cli

import (
	"github.com/csdev/conch/internal/util"
)

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
