package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
  prefixes:
    - "initial commit"
`

const extraneousConfig = `
version: 1

someExtraneousField: false
`

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
			expectedError:  errors.New("only version 1 is supported"),
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
