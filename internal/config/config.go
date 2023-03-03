package config

import (
	"errors"
	"io"

	"github.com/csdev/conch/internal/util"
	"gopkg.in/yaml.v3"
)

type Type struct {
	Types util.CaseInsensitiveSet
	Minor util.CaseInsensitiveSet
	Patch util.CaseInsensitiveSet
}

type Scope struct {
	Required bool
	Scopes   util.CaseInsensitiveSet
}

type Description struct {
	MinLength int `yaml:"minLength"`
	MaxLength int `yaml:"maxLength"`
}

type Footer struct {
	RequiredTokens util.CaseInsensitiveSet `yaml:"requiredTokens"`
	Tokens         util.CaseInsensitiveSet
}

type Policy struct {
	Type
	Scope
	Description
	Footer
}

type Exclude struct {
	Prefixes util.CaseInsensitiveSet
}

type Config struct {
	Version int
	Policy
	Exclude
}

func Default() *Config {
	return &Config{
		Version: 1,
		Policy: Policy{
			Type: Type{
				Minor: util.NewCaseInsensitiveSet([]string{"feat"}),
				Patch: util.NewCaseInsensitiveSet([]string{"fix"}),
			},
			Description: Description{
				MinLength: 1,
			},
		},
	}
}

// Load unmarshals a yaml file to a Config object.
func Load(file io.Reader) (*Config, error) {
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var c Config
	err := decoder.Decode(&c)
	if err != nil {
		return nil, err
	}

	if c.Version != 1 {
		return nil, errors.New("only version 1 is supported")
	}

	return &c, nil
}
