package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const defaultConfig = `
# A standard configuration file for conch, the Conventional Commits checker.
version: 1

policy:
  type:
    types: []
    minor:
      - feat
    patch:
      - fix

  scope:
    required: false
    scopes: []

  description:
    minLength: 1
    maxLength: 0

  footer:
    requiredTokens: []
    tokens: []

exclude:
  prefixes: []
`

const extraneousConfig = `
version: 1

someExtraneousField: false
`

func TestDiscover(t *testing.T) {
	dir, err := os.MkdirTemp("", "conch_tests_")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	configPath := filepath.Join(dir, "conch.yml")
	_, err = os.Create(configPath)
	require.NoError(t, err)

	dir2, err := os.MkdirTemp("", "conch_tests_")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir2)
	})

	tests := []struct {
		description   string
		dirname       string
		expectedPath  string
		expectedError error
	}{
		{
			description:   "it returns the path to the config file",
			dirname:       dir,
			expectedPath:  configPath,
			expectedError: nil,
		},
		{
			description:   "it returns an empty path if the file does not exist",
			dirname:       dir2,
			expectedPath:  "",
			expectedError: nil,
		},
		{
			description:   "it returns an error if the directory does not exist",
			dirname:       "",
			expectedPath:  "",
			expectedError: os.ErrNotExist,
		},
		{
			description:   "it returns an error if the location is not a directory",
			dirname:       configPath,
			expectedPath:  "",
			expectedError: ErrLocation,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			p, err := Discover(test.dirname)
			assert.Equal(t, test.expectedPath, p)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		description    string
		fileContents   string
		expectedConfig *Config
		expectedError  error
	}{
		{
			description:    "minimal config can be decoded",
			fileContents:   `version: 1`,
			expectedConfig: &Config{Version: 1},
			expectedError:  nil,
		},
		{
			description:    "default config can be decoded",
			fileContents:   defaultConfig,
			expectedConfig: Default(),
			expectedError:  nil,
		},
		{
			description:    "empty config causes error",
			fileContents:   ``,
			expectedConfig: nil,
			expectedError:  errors.New("EOF"),
		},
		{
			description:    "wrong version causes error",
			fileContents:   `version: 2`,
			expectedConfig: nil,
			expectedError:  ErrVersion,
		},
		{
			description:    "extraneous field causes error",
			fileContents:   extraneousConfig,
			expectedConfig: nil,
			expectedError: &yaml.TypeError{
				Errors: []string{"line 4: field someExtraneousField not found in type config.Config"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			file := strings.NewReader(test.fileContents)
			cfg, err := Load(file)
			assert.Equal(t, test.expectedConfig, cfg)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestOpen(t *testing.T) {
	tempConfig, err := os.CreateTemp("", "conch_*.yml")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(tempConfig.Name())
	})

	_, err = tempConfig.WriteString(`version: 1`)
	require.NoError(t, err)

	tests := []struct {
		description    string
		filename       string
		expectedConfig *Config
		expectedError  error
	}{
		{
			description:    "it returns the default config for an empty filename",
			filename:       "",
			expectedConfig: Default(),
			expectedError:  nil,
		},
		{
			description:    "it opens the file and returns the config",
			filename:       tempConfig.Name(),
			expectedConfig: &Config{Version: 1},
			expectedError:  nil,
		},
		{
			description:    "it returns an error if the file does not exist",
			filename:       "./__some_bad_filename__",
			expectedConfig: nil,
			expectedError:  os.ErrNotExist,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cfg, err := Open(test.filename)
			assert.Equal(t, test.expectedConfig, cfg)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}
