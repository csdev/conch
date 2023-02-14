package commit

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/csdev/conch/internal/config"
	git "github.com/libgit2/git2go/v34"
)

// Commit represents a single conventional commit.
type Commit struct {
	Id          string
	Type        string
	Scope       string
	IsExclaimed bool
	Description string
	Body        string
	Footers     [][2]string // list of (token, value), preserving order and duplicates
}

// based on https://github.com/conventional-commits/parser/tree/v0.4.1#the-grammar
var firstLinePattern = regexp.MustCompile(`^` +
	`(?P<type>[^():!\pZ\x09-\x0D\x{FEFF}]+)` +
	`(?:\((?P<scope>[^()]+)\))?` +
	`(?P<exclaim>!?)` +
	`: ` +
	`(?P<description>.+)` +
	`$`)

func NewCommit(id string) *Commit {
	return &Commit{Id: id}
}

func (c *Commit) setFirstLine(s string) error {
	match := firstLinePattern.FindStringSubmatch(s)
	if match == nil {
		return fmt.Errorf("%s: syntax error: message does not have a proper type/scope/description", c.Id)
	}

	c.Type = match[firstLinePattern.SubexpIndex("type")]
	c.Scope = match[firstLinePattern.SubexpIndex("scope")]
	c.IsExclaimed = match[firstLinePattern.SubexpIndex("exclaim")] == "!"
	c.Description = match[firstLinePattern.SubexpIndex("description")]

	return nil
}

func (c *Commit) setMessage(msg string) error {
	scanner := bufio.NewScanner(strings.NewReader(msg))

	if ok := scanner.Scan(); !ok {
		return fmt.Errorf("%s: syntax error: message cannot be empty", c.Id)
	}
	c.setFirstLine(scanner.Text())

	return nil
}

func ParseRange(r string) ([]*Commit, error) {
	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	revwalk, err := repo.Walk()
	if err != nil {
		return nil, err
	}

	gitErr := revwalk.PushRange(r)
	if gitErr != nil {
		return nil, gitErr
	}
	defer revwalk.Free()

	commits := make([]*Commit, 0, 10)
	parseErr := NewParseError()

	gitErr = revwalk.Iterate(func(gitCommit *git.Commit) bool {
		id := gitCommit.AsObject().Id().String() // the full commit hash from the git oid
		c := NewCommit(id)
		e := c.setMessage(gitCommit.Message())
		if e == nil {
			commits = append(commits, c)
		} else {
			parseErr.Append(e)
		}

		return true // continues iteration
	})
	if gitErr != nil {
		return nil, gitErr
	}
	if parseErr.HasErrors() {
		return nil, parseErr
	}

	return commits, nil
}

// ApplyPolicy checks if the commit is semantically valid
// according to the supplied policy object.
func (c *Commit) ApplyPolicy(policy config.Policy) error {
	if policy.Type.Types != nil && !policy.Type.Types.Contains(c.Type) {
		return fmt.Errorf("%s: policy does not allow commit type: %s", c.Id, c.Type)
	}

	if c.Scope != "" {
		if policy.Scope.Required {
			return fmt.Errorf("%s: policy requires a commit scope", c.Id)
		}
		if policy.Scope.Scopes != nil && !policy.Scope.Scopes.Contains(c.Scope) {
			return fmt.Errorf("%s: policy does not allow commit scope: %s", c.Id, c.Scope)
		}
	}

	descLen := len(c.Description)
	if descLen < policy.Description.MinLength {
		return fmt.Errorf("%s: policy requires a description longer than %d chars",
			c.Id, policy.Description.MinLength)
	}
	if policy.Description.MaxLength > 0 && descLen > policy.Description.MaxLength {
		return fmt.Errorf("%s: policy limits the description to %d chars",
			c.Id, policy.Description.MaxLength)
	}

	// CAUTION: Tokens in footers need not be unique.
	// For example, Github uses one "Co-authored-by" footer for each co-author.
	// https://docs.github.com/en/pull-requests/committing-changes-to-your-project/creating-and-editing-commits/creating-a-commit-with-multiple-authors
	if c.Footers != nil {
		for _, footer := range c.Footers {
			token := footer[0]
			if policy.Footer.Tokens != nil && !policy.Footer.Tokens.Contains(token) {
				return fmt.Errorf("%s: policy does not allow footer token: %s", c.Id, token)
			}
		}
		// TODO: check requiredTokens
	}

	return nil
}

// Summary returns a one-line summary of the commit,
// in the format "type(scope)!: description".
func (c *Commit) Summary() string {
	var s strings.Builder
	s.WriteString(c.Type)

	if c.Scope != "" {
		s.WriteString("(")
		s.WriteString(c.Scope)
		s.WriteString(")")
	}

	if c.IsExclaimed {
		s.WriteString("!")
	}

	s.WriteString(": ")
	s.WriteString(c.Description)

	return s.String()
}
