package commit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBreakingChange(t *testing.T) {
	tests := []struct {
		description string
		footer      *Footer
		isBreaking  bool
		err         error
	}{
		{
			description: "ordinary token is not a breaking change",
			footer:      &Footer{"Ref", ": ", "1234"},
			isBreaking:  false,
			err:         nil,
		},
		{
			description: "breaking change token with whitespace",
			footer:      &Footer{"BREAKING CHANGE", ": ", "foo"},
			isBreaking:  true,
			err:         nil,
		},
		{
			description: "breaking change token with hyphen",
			footer:      &Footer{"BREAKING-CHANGE", ": ", "foo"},
			isBreaking:  true,
			err:         nil,
		},
		{
			description: "breaking change token must have correct separator",
			footer:      &Footer{"BREAKING CHANGE", " #", "foo"},
			isBreaking:  false,
			err:         ErrFooterSep,
		},
		{
			description: "breaking change token with whitespace must have correct capitalization",
			footer:      &Footer{"Breaking change", ": ", "foo"},
			isBreaking:  false,
			err:         ErrFooterCaps,
		},
		{
			description: "breaking change token with hyphen must have correct capitalization",
			footer:      &Footer{"Breaking-change", ": ", "foo"},
			isBreaking:  false,
			err:         ErrFooterCaps,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			isBreaking, err := test.footer.IsBreakingChange()
			assert.Equal(t, test.isBreaking, isBreaking)
			assert.Equal(t, test.err, err)
		})
	}
}

func TestExtractFooters(t *testing.T) {
	tests := []struct {
		description string
		lines       []string
		footers     []Footer
	}{
		{
			description: "empty lines has no footers",
			lines:       []string{},
			footers:     []Footer{},
		},
		{
			description: "plain text has no footers",
			lines:       []string{"some body text"},
			footers:     []Footer{},
		},
		{
			description: "the first line must be a footer",
			lines: []string{
				"some body text",
				"Refs: 1234",
			},
			footers: []Footer{},
		},
		{
			description: "footers can have different separators",
			lines: []string{
				"Refs #1234",
				"Co-authored-by: John Doe <john.doe@example>",
			},
			footers: []Footer{
				{"Refs", " #", "1234"},
				{"Co-authored-by", ": ", "John Doe <john.doe@example>"},
			},
		},
		{
			description: "footers can have duplicate tokens",
			lines: []string{
				"Refs: 1234",
				"Refs: 5678",
			},
			footers: []Footer{
				{"Refs", ": ", "1234"},
				{"Refs", ": ", "5678"},
			},
		},
		{
			description: "footer values can span multiple lines",
			lines: []string{
				"Addendum: foo",
				"bar",
				"baz",
			},
			footers: []Footer{
				{"Addendum", ": ", "foo\nbar\nbaz"},
			},
		},
		{
			description: "standard footer tokens cannot have whitespace",
			lines: []string{
				"issue ref: 1234",
			},
			footers: []Footer{},
		},
		{
			description: "breaking change token can have whitespace",
			lines: []string{
				"BREAKING CHANGE: removed field from API",
			},
			footers: []Footer{
				{"BREAKING CHANGE", ": ", "removed field from API"},
			},
		},
		{
			description: "breaking change token can have hyphen",
			lines: []string{
				"BREAKING-CHANGE: removed field from API",
			},
			footers: []Footer{
				{"BREAKING-CHANGE", ": ", "removed field from API"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.footers, extractFooters(test.lines))
		})
	}
}
