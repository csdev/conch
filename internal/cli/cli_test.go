package cli

import (
	"os"
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

func TestGetFileContents(t *testing.T) {
	f, err := os.CreateTemp("", "conch_tests_")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	const fileContents = "test line 1\ntest line 2\n"
	_, err = f.WriteString(fileContents)
	require.NoError(t, err)

	tests := []struct {
		description      string
		filename         string
		expectedContents string
		expectedErr      error
	}{
		{
			description:      "it returns the contents of the file",
			filename:         f.Name(),
			expectedContents: fileContents,
			expectedErr:      nil,
		},
		{
			description:      "it returns an error for an invalid filename",
			filename:         "__bad_filename__",
			expectedContents: "",
			expectedErr:      os.ErrNotExist,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			s, err := GetFileContents(test.filename)
			assert.Equal(t, test.expectedContents, s)
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
}
