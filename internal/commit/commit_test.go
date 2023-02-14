package commit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
