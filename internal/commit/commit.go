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
	Footers     []Footer
	IsBreaking  bool
}

func ErrSyntax(id string, msg string) error {
	return fmt.Errorf("%s: syntax error: %s", id, msg)
}

func ErrEmpty(id string) error {
	return ErrSyntax(id, "commit message cannot be empty")
}

func ErrSummary(id string) error {
	return ErrSyntax(id, "commit summary must contain a valid type, optional scope, and description")
}

func ErrBlankLine(id string) error {
	return ErrSyntax(id, "the commit summary must be followed by a blank line")
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
		return ErrSummary(c.Id)
	}

	c.Type = match[firstLinePattern.SubexpIndex("type")]
	c.Scope = match[firstLinePattern.SubexpIndex("scope")]
	c.IsExclaimed = match[firstLinePattern.SubexpIndex("exclaim")] == "!"
	c.Description = match[firstLinePattern.SubexpIndex("description")]

	if c.IsExclaimed {
		c.IsBreaking = true
	}

	return nil
}

func (c *Commit) setMessage(msg string) error {
	scanner := bufio.NewScanner(strings.NewReader(msg))

	if ok := scanner.Scan(); !ok {
		return ErrEmpty(c.Id)
	}
	err := c.setFirstLine(scanner.Text())
	if err != nil {
		return err
	}

	if ok := scanner.Scan(); !ok {
		return nil // end of commit message (no body or footers)
	}

	if scanner.Text() != "" {
		return ErrBlankLine(c.Id)
	}

	// The body of the commit message may consist of multiple paragraphs,
	// each separated by a blank line. The final paragraph may be part of
	// the body, or it may actually be the footers.
	// Read in the remainder of the message, and keep track of where the final
	// paragraph begins, so we can apply footer matching to it.

	lines := make([]string, 0, 10)
	lineNum := 0
	parStart := -1
	isPar := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			isPar = false
		} else if !isPar {
			isPar = true
			parStart = lineNum
		}

		lines = append(lines, line)
		lineNum += 1
	}

	if parStart >= 0 {
		footers := extractFooters(lines[parStart:])
		if len(footers) == 0 {
			// No footers were detected. The commit body is the entire
			// block of text.
			c.Body = strings.Join(lines, "\n")
		} else {
			// Footers were extracted from the final paragraph.
			// The commit body consists of all the previous paragraphs.
			c.Body = strings.TrimRight(strings.Join(lines[:parStart], "\n"), "\n")
			c.Footers = footers
		}
	}

	for _, footer := range c.Footers {
		isBreaking, err := footer.IsBreakingChange()
		if err != nil {
			return ErrSyntax(c.Id, err.Error())
		}
		if isBreaking {
			c.IsBreaking = true
			break
		}
	}

	return nil
}

func ParseRange(repoPath string, rangeSpec string) ([]*Commit, error) {
	commits := make([]*Commit, 0, 10)

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return commits, err
	}
	defer repo.Free()

	revwalk, err := repo.Walk()
	if err != nil {
		return commits, err
	}

	gitErr := revwalk.PushRange(rangeSpec)
	if gitErr != nil {
		return commits, gitErr
	}
	defer revwalk.Free()

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
		return commits, gitErr
	}
	if parseErr.HasErrors() {
		return commits, parseErr
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
		for _, f := range c.Footers {
			if policy.Footer.Tokens != nil && !policy.Footer.Tokens.Contains(f.Token) {
				return fmt.Errorf("%s: policy does not allow footer token: %s", c.Id, f.Token)
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

	// Since the summary does not show the footers, always use an exclamation
	// point to indicate a breaking change (even if the original commit
	// message did not use one).
	if c.IsBreaking {
		s.WriteString("!")
	}

	s.WriteString(": ")
	s.WriteString(c.Description)

	return s.String()
}

const (
	Breaking = iota
	Minor
	Patch
	Uncategorized
)

func (c *Commit) Classification(cfg *config.Config) int {
	if c.IsBreaking {
		return Breaking
	}
	if cfg.Policy.Minor.Contains(c.Type) {
		return Minor
	}
	if cfg.Policy.Patch.Contains(c.Type) {
		return Patch
	}
	return Uncategorized
}
