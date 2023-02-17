package commit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				{"Refs", "1234"},
				{"Co-authored-by", "John Doe <john.doe@example>"},
			},
		},
		{
			description: "footers can have duplicate tokens",
			lines: []string{
				"Refs: 1234",
				"Refs: 5678",
			},
			footers: []Footer{
				{"Refs", "1234"},
				{"Refs", "5678"},
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
				{"Addendum", "foo\nbar\nbaz"},
			},
		},
		{
			description: "standard footer tokens cannot have whitespace",
			lines: []string{
				"issue ref: 1234",
			},
			footers: []Footer{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.footers, extractFooters(test.lines))
		})
	}
}

func TestSetFirstLine(t *testing.T) {
	syntaxErr := errors.New("0: syntax error: message does not have a proper type/scope/description")

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
				Type:        "feat",
				IsExclaimed: true,
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "it sets the optional scope and breaking change exclamation",
			message:     "feat(things)!: implement the thing",
			commit: &Commit{
				Id:          "0",
				Type:        "feat",
				Scope:       "things",
				IsExclaimed: true,
				Description: "implement the thing",
			},
			err: nil,
		},
		{
			description: "it accepts uppercase characters",
			message:     "Feat(Things): Implement the THING",
			commit: &Commit{
				Id:          "0",
				Type:        "Feat",
				Scope:       "Things",
				Description: "Implement the THING",
			},
			err: nil,
		},
		{
			description: "it accepts punctuation",
			message:     "feat.minor(the-things!)!: implement the thing!",
			commit: &Commit{
				Id:          "0",
				Type:        "feat.minor",
				Scope:       "the-things!",
				IsExclaimed: true,
				Description: "implement the thing!",
			},
			err: nil,
		},
		{
			description: "it accepts whitespace in the scope",
			message:     "feat(app widgets): implement the thing",
			commit: &Commit{
				Id:          "0",
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
				Type:        "typé",
				Scope:       "scopé",
				Description: "déscription",
			},
		},
		{
			description: "it does not allow an empty line",
			message:     "",
			commit:      &Commit{Id: "0"},
			err:         syntaxErr,
		},
		{
			description: "it does not allow whitespace in the type",
			message:     "feat : implement the thing",
			commit:      &Commit{Id: "0"},
			err:         syntaxErr,
		},
		{
			description: "it does not allow utf8 control whitespace in the type",
			message:     "feat\t: implement the thing",
			commit:      &Commit{Id: "0"},
			err:         syntaxErr,
		},
		{
			description: "it does not allow utf8 separator whitespace in the type",
			message:     "feat\u2002: implement the thing",
			commit:      &Commit{Id: "0"},
			err:         syntaxErr,
		},
		{
			description: "it does not allow utf8 bom/zwnbsp in the type",
			message:     "feat\ufeff: implement the thing",
			commit:      &Commit{Id: "0"},
			err:         syntaxErr,
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
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"Refs", "#1234"},
					{"Signed-off-by", "John Doe <john.doe@example>"},
				},
			},
		},
		{
			description: "summary, body, and footers",
			message:     "feat: implement the thing\n\n1a\n1b\n\n2a\n2b\n\nRefs: #1234",
			commit: &Commit{
				Id:          "0",
				Type:        "feat",
				Description: "implement the thing",
				Body:        "1a\n1b\n\n2a\n2b",
				Footers: []Footer{
					{"Refs", "#1234"},
				},
			},
			err: nil,
		},
		{
			description: "multi-line footers",
			message:     "feat: implement the thing\n\nBREAKING-CHANGE: the api\nis different\n",
			commit: &Commit{
				Id:          "0",
				Type:        "feat",
				Description: "implement the thing",
				Footers: []Footer{
					{"BREAKING-CHANGE", "the api\nis different"},
				},
			},
			err: nil,
		},
		{
			description: "message cannot be empty",
			message:     "",
			commit:      &Commit{Id: "0"},
			err:         errors.New("0: syntax error: message cannot be empty"),
		},
		{
			description: "first line must be correct",
			message:     "asdf",
			commit:      &Commit{Id: "0"},
			err:         errors.New("0: syntax error: message does not have a proper type/scope/description"),
		},
		{
			description: "blank line needed between summary and body",
			message:     "feat: implement the thing\nasdf\n",
			commit: &Commit{
				Id:          "0",
				Type:        "feat",
				Description: "implement the thing",
			},
			err: errors.New("0: syntax error: the commit summary must be followed by a blank line"),
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
