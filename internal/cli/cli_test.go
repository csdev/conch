package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	tests := []struct {
		description    string
		contents       string
		expectedOutput string
	}{
		{
			description:    "it replaces escape sequences",
			contents:       `\\ \t \n`,
			expectedOutput: "\\ \t \n",
		},
		{
			description:    "it handles overlapping backslash escapes",
			contents:       `\\n`,
			expectedOutput: `\n`,
		},
		{
			description:    "it produces a usable template for formatting objects",
			contents:       `{{ .K }}`,
			expectedOutput: "val",
		},
		{
			description:    "it can be empty",
			contents:       ``,
			expectedOutput: "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tpl, err := Template("mytemplate", test.contents)
			require.NoError(t, err)

			out := strings.Builder{}
			obj := struct{ K string }{"val"}

			err = tpl.Execute(&out, obj)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedOutput, out.String())
		})
	}
}
