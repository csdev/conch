package commit

import (
	"os"
	"testing"
	"time"

	"github.com/csdev/conch/internal/config"
	"github.com/csdev/conch/internal/util"
	git "github.com/libgit2/git2go/v34"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFirstLine(t *testing.T) {
	tests := []struct {
		description string
		message     string
		commit      *Commit
		err         error
	}{
		{
			description: "it sets the type and description",
			message:     "feat: implement the thing",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "it sets the optional scope",
			message:     "feat(things): implement the thing",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Scope:       "things",
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "it sets the optional breaking change exclamation",
			message:     "feat!: implement the thing",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				IsExclaimed: true,
				Description: "implement the thing",
				IsBreaking:  true,
			},
			err: nil,
		},
		{
			description: "it sets the optional scope and breaking change exclamation",
			message:     "feat(things)!: implement the thing",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Scope:       "things",
				IsExclaimed: true,
				Description: "implement the thing",
				IsBreaking:  true,
			},
			err: nil,
		},
		{
			description: "it accepts uppercase characters",
			message:     "Feat(Things): Implement the THING",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "Feat",
				Scope:       "Things",
				Description: "Implement the THING",
			},
			err: nil,
		},
		{
			description: "it accepts punctuation",
			message:     "feat.minor(the:things!)!: implement the thing!",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat.minor",
				Scope:       "the:things!",
				IsExclaimed: true,
				Description: "implement the thing!",
				IsBreaking:  true,
			},
			err: nil,
		},
		{
			description: "it accepts whitespace in the scope",
			message:     "feat(app widgets): implement the thing",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Scope:       "app widgets",
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "it accepts other utf8 characters",
			message:     "typé(scopé): déscription",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "typé",
				Scope:       "scopé",
				Description: "déscription",
			},
		},
		{
			description: "it does not allow an empty line",
			message:     "",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow a missing type",
			message:     "implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow an empty type",
			message:     ": implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow whitespace in the type",
			message:     "feat : implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow utf8 control whitespace in the type",
			message:     "feat\t: implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow utf8 separator whitespace in the type",
			message:     "feat\u2002: implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow utf8 bom/zwnbsp in the type",
			message:     "feat\ufeff: implement the thing",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "it does not allow an empty description",
			message:     "feat: ",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := NewCommit("0")
			err := c.setFirstLine(test.message)
			assert.Equal(t, test.commit, c)
			assert.Equal(t, test.err, err)
		})
	}
}

func TestSetMessage(t *testing.T) {
	tests := []struct {
		description string
		message     string
		commit      *Commit
		err         error
	}{
		{
			description: "summary only",
			message:     "feat: implement the thing\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "summary with extra blank lines",
			message:     "feat: implement the thing\n\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "summary and body",
			message:     "feat: implement the thing\n\ndescription line 1\ndescription line 2\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Body:        "description line 1\ndescription line 2",
			},
			err: nil,
		},
		{
			description: "summary and multiple body paragraphs",
			message:     "feat: implement the thing\n\n1a\n1b\n\n2a\n2b\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Body:        "1a\n1b\n\n2a\n2b",
			},
			err: nil,
		},
		{
			description: "summary and footers",
			message:     "feat: implement the thing\n\nRefs: #1234\nSigned-off-by: John Doe <john.doe@example>\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"Refs", ": ", "#1234"},
					{"Signed-off-by", ": ", "John Doe <john.doe@example>"},
				},
			},
		},
		{
			description: "summary, body, and footers",
			message:     "feat: implement the thing\n\n1a\n1b\n\n2a\n2b\n\nRefs: #1234",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Body:        "1a\n1b\n\n2a\n2b",
				Footers: []Footer{
					{"Refs", ": ", "#1234"},
				},
			},
			err: nil,
		},
		{
			description: "multi-line footers",
			message:     "feat: implement the thing\n\nRefs: 1234\n5678\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"Refs", ": ", "1234\n5678"},
				},
			},
			err: nil,
		},
		{
			description: "breaking change footer",
			message:     "feat: implement the thing\n\nBREAKING CHANGE: the API is different",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"BREAKING CHANGE", ": ", "the API is different"},
				},
				IsBreaking: true,
			},
			err: nil,
		},
		{
			description: "message cannot be empty",
			message:     "",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrEmpty("0"),
		},
		{
			description: "first line must be correct",
			message:     "asdf",
			commit:      &Commit{Id: "0", ShortId: "0"},
			err:         ErrSummary("0"),
		},
		{
			description: "blank line needed between summary and body",
			message:     "feat: implement the thing\nasdf\n",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
			},
			err: ErrBlankLine("0"),
		},
		{
			description: "breaking change must be reported correctly",
			message:     "feat: implement the thing\n\nbreaking-change: foo",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"breaking-change", ": ", "foo"},
				},
			},
			err: ErrSyntax("0", ErrFooterCaps.Error()),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := NewCommit("0")
			err := c.setMessage(test.message)
			assert.Equal(t, test.commit, c)
			assert.Equal(t, test.err, err)
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		description string
		msg         string
		cfg         *config.Config
		expected    bool
	}{
		{
			description: "it allows the commit if there are no prefixes",
			msg:         "Merge pull request #1234",
			cfg:         &config.Config{},
			expected:    false,
		},
		{
			description: "it allows the commit if there are no matching prefixes",
			msg:         "Merge pull request #1234",
			cfg: &config.Config{
				Exclude: config.Exclude{
					Prefixes: util.NewCaseInsensitiveSet([]string{
						"pull request",
						"#1234",
					}),
				},
			},
			expected: false,
		},
		{
			description: "it excludes the commit if the prefix matches exactly",
			msg:         "Merge pull request #1234",
			cfg: &config.Config{
				Exclude: config.Exclude{
					Prefixes: util.NewCaseInsensitiveSet([]string{"Merge pull request"}),
				},
			},
			expected: true,
		},
		{
			description: "it performs case insensitive matching",
			msg:         "Merge pull request #1234",
			cfg: &config.Config{
				Exclude: config.Exclude{
					Prefixes: util.NewCaseInsensitiveSet([]string{"merge pull request"}),
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, isExcluded(test.msg, test.cfg))
		})
	}
}

func makeTestRepo(t *testing.T, msgs []string) (string, []*git.Oid) {
	// make a git repo inside a temp directory that we can use for testing
	dir, err := os.MkdirTemp("", "conch_tests_")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	repo, err := git.InitRepository(dir, true)
	require.NoError(t, err)
	t.Cleanup(func() {
		repo.Free()
	})

	// get the current index and write it to a tree, so we can use it
	// to construct a blank commit
	// (we don't care about the files, just the commit messages)
	idx, err := repo.Index()
	require.NoError(t, err)

	tree, err := idx.WriteTree()
	require.NoError(t, err)

	// create a signature object, which is used to specify the author
	// and the committer
	sig := &git.Signature{
		Name:  "Test User",
		Email: "test.user@email.example",
		When:  time.Now(),
	}

	var head *git.Oid
	oids := make([]*git.Oid, 0, len(msgs))

	for _, msg := range msgs {
		head, err = repo.CreateCommitFromIds("HEAD", sig, sig, msg, tree, head)
		require.NoError(t, err)
		oids = append(oids, head)
	}

	return dir, oids
}

func TestParseRange(t *testing.T) {
	dir, oids := makeTestRepo(t, []string{
		"initial commit",
		"the next commit",
		"chore: the most recent commit",
	})

	tests := []struct {
		description     string
		repoPath        string
		rangeSpec       string
		cfg             *config.Config
		expectedCommits []*Commit
		expectedErr     error
	}{
		{
			description: "it returns the commits in the range",
			repoPath:    dir,
			rangeSpec:   "HEAD~1..",
			cfg:         config.Default(),
			expectedCommits: []*Commit{
				{
					Id:          oids[2].String(),
					ShortId:     oids[2].String()[:7],
					Type:        "chore",
					Description: "the most recent commit",
				},
			},
			expectedErr: nil,
		},
		{
			description:     "it returns an empty slice if there are no commits in the range",
			repoPath:        dir,
			rangeSpec:       "HEAD..HEAD",
			cfg:             config.Default(),
			expectedCommits: []*Commit{},
			expectedErr:     nil,
		},
		{
			description:     "it returns errors in the range",
			repoPath:        dir,
			rangeSpec:       "HEAD~2..HEAD~1",
			cfg:             config.Default(),
			expectedCommits: []*Commit{},
			expectedErr: &ParseError{
				Errors: []string{
					ErrSummary(oids[1].String()[:7]).Error(),
				},
			},
		},
		{
			description: "it excludes commits based on the config",
			repoPath:    dir,
			rangeSpec:   "HEAD~2..HEAD~1",
			cfg: &config.Config{
				Exclude: config.Exclude{
					Prefixes: util.NewCaseInsensitiveSet([]string{"the next"}),
				},
			},
			expectedCommits: []*Commit{},
			expectedErr:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			commits, err := ParseRange(test.repoPath, test.rangeSpec, test.cfg)
			assert.Equal(t, test.expectedCommits, commits)
			assert.Equal(t, test.expectedErr, err)
		})
	}

	tests2 := []struct {
		description  string
		repoPath     string
		rangeSpec    string
		errorPattern string
	}{
		{
			description:  "it returns an error for an invalid path",
			repoPath:     "./__invalid_path__",
			rangeSpec:    "..",
			errorPattern: "failed to resolve path",
		},
		{
			description:  "it returns an error for an invalid commit range",
			repoPath:     dir,
			rangeSpec:    "HEAD",
			errorPattern: "invalid revspec",
		},
	}

	for _, test := range tests2 {
		t.Run(test.description, func(t *testing.T) {
			commits, err := ParseRange(test.repoPath, test.rangeSpec, config.Default())
			assert.Equal(t, []*Commit{}, commits)
			assert.ErrorContains(t, err, test.errorPattern)
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		description     string
		msg             string
		cfg             *config.Config
		expectedCommits []*Commit
		expectedErr     error
	}{
		{
			description: "it returns a valid commit",
			msg:         "feat: a new thing",
			cfg:         config.Default(),
			expectedCommits: []*Commit{
				{
					Id:          "0",
					ShortId:     "0",
					Type:        "feat",
					Description: "a new thing",
				},
			},
			expectedErr: nil,
		},
		{
			description: "it excludes a commit based on the config",
			msg:         "revert the thing",
			cfg: &config.Config{
				Exclude: config.Exclude{
					Prefixes: util.NewCaseInsensitiveSet([]string{"revert"}),
				},
			},
			expectedCommits: []*Commit{},
			expectedErr:     nil,
		},
		{
			description:     "it returns an error for an invalid commit message",
			msg:             "revert the thing",
			cfg:             config.Default(),
			expectedCommits: []*Commit{},
			expectedErr:     ErrSummary("0"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			commits, err := ParseMessage(test.msg, test.cfg)
			assert.Equal(t, test.expectedCommits, commits)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestApplyPolicy(t *testing.T) {
	commit := &Commit{
		Id:          "0",
		ShortId:     "0",
		Type:        "chore",
		Scope:       "deps",
		Description: "upgrade stuff",
		Footers: []Footer{
			{"Refs", ": ", "1234"},
		},
	}

	tests := []struct {
		description string
		cfg         *config.Config
		err         error
	}{
		{
			description: "it reports no violations",
			cfg:         config.Default(),
			err:         nil,
		},
		{
			description: "it reports an unrecognized commit type",
			cfg: &config.Config{
				Policy: config.Policy{
					Type: config.Type{
						Types: util.NewCaseInsensitiveSet([]string{"feat", "fix"}),
					},
				},
			},
			err: ErrUnrecognizedType("0"),
		},
		{
			description: "it reports an unrecognized commit scope",
			cfg: &config.Config{
				Policy: config.Policy{
					Scope: config.Scope{
						Scopes: util.NewCaseInsensitiveSet([]string{"API"}),
					},
				},
			},
			err: ErrUnrecognizedScope("0"),
		},
		{
			description: "it checks for a description exceeding the min length",
			cfg: &config.Config{
				Policy: config.Policy{
					Description: config.Description{
						MinLength: 14,
					},
				},
			},
			err: ErrDescriptionLength("0", 14, 0),
		},
		{
			description: "it checks for a description exceeding the max length",
			cfg: &config.Config{
				Policy: config.Policy{
					Description: config.Description{
						MaxLength: 12,
					},
				},
			},
			err: ErrDescriptionLength("0", 1, 12),
		},
		{
			description: "it reports an unrecognized token in the footers",
			cfg: &config.Config{
				Policy: config.Policy{
					Footer: config.Footer{
						Tokens: util.NewCaseInsensitiveSet([]string{
							"BREAKING CHANGE",
							"BREAKING-CHANGE",
						}),
					},
				},
			},
			err: ErrUnrecognizedFooter("0", "Refs"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.err, commit.ApplyPolicy(test.cfg))
		})
	}
}

func TestApplyPolicy_RequiredFields(t *testing.T) {
	cfg := &config.Config{
		Policy: config.Policy{
			Scope: config.Scope{
				Required: true,
			},
			Footer: config.Footer{
				RequiredTokens: util.NewCaseInsensitiveSet([]string{
					"refs",
					"signed-off-by",
				}),
			},
		},
	}

	tests := []struct {
		description string
		commit      *Commit
		err         error
	}{
		{
			description: "it checks for a missing scope",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "chore",
				Description: "upgrade stuff",
				Footers: []Footer{
					{"Refs", ": ", "1234"},
					{"Signed-off-by", ": ", "John Doe <john.doe@example>"},
				},
			},
			err: ErrRequiredScope("0"),
		},
		{
			description: "it checks for missing footers",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "chore",
				Scope:       "deps",
				Description: "upgrade stuff",
				Footers: []Footer{
					{"Refs", ": ", "1234"},
				},
			},
			err: ErrRequiredFooters("0", util.NewCaseInsensitiveSet([]string{"signed-off-by"})),
		},
		{
			description: "it reports multiple missing footers",
			commit: &Commit{
				Id:          "0",
				ShortId:     "0",
				Type:        "chore",
				Scope:       "deps",
				Description: "upgrade stuff",
			},
			err: ErrRequiredFooters("0", util.NewCaseInsensitiveSet([]string{
				"refs",
				"signed-off-by",
			})),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.err, test.commit.ApplyPolicy(cfg))
		})
	}
}

func TestApplyPolicySlice(t *testing.T) {
	commits := []*Commit{
		{
			Id:          "0",
			ShortId:     "0",
			Type:        "chore",
			Scope:       "deps",
			Description: "upgrade stuff",
			Footers: []Footer{
				{"Refs", ": ", "1234"},
			},
		},
		{
			Id:          "1",
			ShortId:     "1",
			Type:        "ci",
			Description: "add environment variables",
			Footers: []Footer{
				{"Refs", ": ", "5678"},
			},
		},
	}

	tests := []struct {
		description string
		cfg         *config.Config
		err         error
	}{
		{
			description: "no errors",
			cfg:         config.Default(),
			err:         nil,
		},
		{
			description: "multiple errors",
			cfg: &config.Config{
				Policy: config.Policy{
					Type: config.Type{
						Types: util.NewCaseInsensitiveSet([]string{"feat", "fix", "chore"}),
					},
					Scope: config.Scope{
						Scopes: util.NewCaseInsensitiveSet([]string{""}),
					},
				},
			},
			err: &ParseError{
				Errors: []string{
					ErrUnrecognizedScope("0").Error(),
					ErrUnrecognizedType("1").Error(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.err, ApplyPolicy(commits, test.cfg))
		})
	}
}

func TestSummary(t *testing.T) {
	tests := []struct {
		description string
		commit      *Commit
		summary     string
	}{
		{
			description: "type and description",
			commit: &Commit{
				Type:        "feat",
				Description: "implement the thing",
			},
			summary: "feat: implement the thing",
		},
		{
			description: "scope",
			commit: &Commit{
				Type:        "feat",
				Scope:       "things",
				Description: "implement the thing",
			},
			summary: "feat(things): implement the thing",
		},
		{
			description: "breaking change",
			commit: &Commit{
				Type:        "feat",
				Description: "implement the thing",
				IsBreaking:  true,
			},
			summary: "feat!: implement the thing",
		},
		{
			description: "breaking change with scope",
			commit: &Commit{
				Type:        "feat",
				Scope:       "things",
				Description: "implement the thing",
				IsBreaking:  true,
			},
			summary: "feat(things)!: implement the thing",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.summary, test.commit.Summary())
		})
	}
}

func TestClassification(t *testing.T) {
	tests := []struct {
		description string
		commit      *Commit
		expected    int
	}{
		{
			description: "it identifies a breaking change",
			commit: &Commit{
				Type:        "feat",
				IsExclaimed: true,
				IsBreaking:  true,
			},
			expected: Breaking,
		},
		{
			description: "it identifies a minor change",
			commit: &Commit{
				Type: "feat",
			},
			expected: Minor,
		},
		{
			description: "it identifies a patch",
			commit: &Commit{
				Type: "fix",
			},
			expected: Patch,
		},
		{
			description: "it identifies an uncategorized change",
			commit: &Commit{
				Type: "chore",
			},
			expected: Uncategorized,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, test.commit.Classification(config.Default()))
		})
	}
}

func TestStripComments(t *testing.T) {
	tests := []struct {
		description string
		msg         string
		expected    string
	}{
		{
			description: "it works on an empty commit message",
			msg:         "",
			expected:    "",
		},
		{
			description: "it works on an empty line",
			msg:         "\n",
			expected:    "\n",
		},
		{
			description: "it removes a comment",
			msg:         "#comment\nsome text\n",
			expected:    "some text\n",
		},
		{
			description: "it ignores # in the middle of a line",
			msg:         "some text # not a comment\n",
			expected:    "some text # not a comment\n",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, StripComments(test.msg))
		})
	}
}
