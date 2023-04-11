package commit

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/csdev/conch/internal/config"
	"github.com/csdev/conch/internal/util"
	git "github.com/libgit2/git2go/v34"
	log "github.com/sirupsen/logrus"
)

// Commit represents a single conventional commit.
type Commit struct {
	Id          string
	ShortId     string
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

func ErrPolicy(id string, msg string) error {
	return fmt.Errorf("%s: policy error: %s", id, msg)
}

func ErrUnrecognizedType(id string) error {
	return ErrPolicy(id, "unrecognized commit type")
}

func ErrRequiredScope(id string) error {
	return ErrPolicy(id, "commit must have a scope")
}

func ErrUnrecognizedScope(id string) error {
	return ErrPolicy(id, "unrecognized commit scope")
}

func ErrDescriptionLength(id string, min int, max int) error {
	if min < 1 {
		min = 1
	}

	if max > 0 {
		return ErrPolicy(id, fmt.Sprintf("description must be between %d and %d chars long", min, max))
	}
	return ErrPolicy(id, fmt.Sprintf("description must be longer than %d chars", min))
}

func ErrUnrecognizedFooter(id string, token string) error {
	return ErrPolicy(id, fmt.Sprintf("unrecognized footer: %s", token))
}

func ErrRequiredFooters(id string, tokens util.CaseInsensitiveSet) error {
	ts := make([]string, 0, len(tokens))
	for token := range tokens {
		ts = append(ts, token)
	}
	sort.Strings(ts) // makes errors easily comparable
	return ErrPolicy(id, fmt.Sprintf("commit must include footers: %s", strings.Join(ts, ", ")))
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
	return &Commit{
		Id:      id,
		ShortId: id,
	}
}

func (c *Commit) setFirstLine(s string) error {
	match := firstLinePattern.FindStringSubmatch(s)
	if match == nil {
		return ErrSummary(c.ShortId)
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
		return ErrEmpty(c.ShortId)
	}
	err := c.setFirstLine(scanner.Text())
	if err != nil {
		return err
	}

	if ok := scanner.Scan(); !ok {
		return nil // end of commit message (no body or footers)
	}

	if scanner.Text() != "" {
		return ErrBlankLine(c.ShortId)
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
			return ErrSyntax(c.ShortId, err.Error())
		}
		if isBreaking {
			c.IsBreaking = true
			break
		}
	}

	return nil
}

func isExcluded(msg string, cfg *config.Config) bool {
	if cfg.Exclude.Prefixes == nil {
		return false
	}
	m := strings.ToLower(msg)
	for prefix := range cfg.Exclude.Prefixes {
		if strings.HasPrefix(m, prefix) {
			return true
		}
	}
	return false
}

// IterRange parses all of the commit messages in the range. For each commit,
// it invokes the callback function with the parsed Commit object, or an
// error if the commit did not obey the Conventional Commits standard.
// The callback function can abort the iteration by returning false.
func IterRange(repoPath string, rangeSpec string, cfg *config.Config, f func(*Commit, error) bool) error {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}
	defer repo.Free()

	revwalk, err := repo.Walk()
	if err != nil {
		return err
	}

	gitErr := revwalk.PushRange(rangeSpec)
	if gitErr != nil {
		return gitErr
	}
	defer revwalk.Free()

	return revwalk.Iterate(func(gitCommit *git.Commit) bool {
		msg := gitCommit.Message()
		if isExcluded(msg, cfg) {
			return true // continues iteration, skipping over commit parsing
		}

		obj := gitCommit.AsObject()
		id := obj.Id().String() // the full commit hash from the git oid
		c := NewCommit(id)

		sid, err := obj.ShortId()
		if err != nil {
			log.Panicf("broken git repo? failed to get short id of commit %s: %v", id, err)
		}
		c.ShortId = sid

		e := c.setMessage(msg)
		return f(c, e)
	})
}

// ParseRange parses all of the commit messages in the range and returns
// a slice of the resulting Commit objects. If an error occurs, the slice
// may contain a partial set of all the commits that were successfully
// processed so far.
func ParseRange(repoPath string, rangeSpec string, cfg *config.Config) ([]*Commit, error) {
	commits := make([]*Commit, 0, 10)
	parseErr := NewParseError()

	err := IterRange(repoPath, rangeSpec, cfg, func(c *Commit, err error) bool {
		if err != nil {
			parseErr.Append(err)
		} else {
			commits = append(commits, c)
		}
		return true
	})

	if err != nil {
		return commits, err
	}
	if parseErr.HasErrors() {
		return commits, parseErr
	}
	return commits, nil
}

func ParseMessage(msg string, cfg *config.Config) (*Commit, error) {
	c := NewCommit("0")
	err := c.setMessage(msg)
	return c, err
}

// ApplyPolicy checks if the commit is semantically valid
// according to the supplied policy object.
func (c *Commit) ApplyPolicy(cfg *config.Config) error {
	policy := &cfg.Policy
	if policy.Type.Types != nil && !policy.Type.Types.Contains(c.Type) {
		return ErrUnrecognizedType(c.ShortId)
	}

	if c.Scope == "" {
		if policy.Scope.Required {
			return ErrRequiredScope(c.ShortId)
		}
	} else {
		if policy.Scope.Scopes != nil && !policy.Scope.Scopes.Contains(c.Scope) {
			return ErrUnrecognizedScope(c.ShortId)
		}
	}

	descLen := len(c.Description)
	min := policy.Description.MinLength
	max := policy.Description.MaxLength
	if (descLen < min) || (max > 0 && descLen > max) {
		return ErrDescriptionLength(c.ShortId, min, max)
	}

	// CAUTION: Tokens in footers need not be unique.
	// For example, Github uses one "Co-authored-by" footer for each co-author.
	// https://docs.github.com/en/pull-requests/committing-changes-to-your-project/creating-and-editing-commits/creating-a-commit-with-multiple-authors
	var reqTokens util.CaseInsensitiveSet
	if policy.Footer.RequiredTokens != nil {
		reqTokens = policy.Footer.RequiredTokens.Copy()
	}

	for _, f := range c.Footers {
		if policy.Footer.Tokens != nil && !policy.Footer.Tokens.Contains(f.Token) {
			return ErrUnrecognizedFooter(c.ShortId, f.Token)
		}
		reqTokens.Remove(f.Token)
	}

	if len(reqTokens) > 0 {
		return ErrRequiredFooters(c.ShortId, reqTokens)
	}

	return nil
}

func ApplyPolicy(commits []*Commit, cfg *config.Config) error {
	parseErr := NewParseError()

	for _, c := range commits {
		err := c.ApplyPolicy(cfg)
		if err != nil {
			parseErr.Append(err)
		}
	}

	if parseErr.HasErrors() {
		return parseErr
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

// StripComments removes all lines that start with "#" from the input,
// and returns the resulting string.
func StripComments(msg string) string {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	var out strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}

	return out.String()
}
