// Package config provides tools for loading conch.yml configuration files.
package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"

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

const StandardFilename = "conch.yml"

var ErrLocation = errors.New("location must be a valid directory")
var ErrVersion = errors.New("only version 1 is supported")

// Default returns the default configuration, which is used when the
// repository does not include its own configuration file.
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

// Discover looks for a configuration file in the specified directory,
// and returns the path to it. If the file does not exist, it returns
// an empty string. If the directory does not exist, it returns an error.
func Discover(dirname string) (string, error) {
	dirinfo, err := os.Stat(dirname)
	if err != nil {
		return "", err
	}
	if !dirinfo.IsDir() {
		return "", ErrLocation
	}

	p := filepath.Join(dirname, StandardFilename)
	_, err = os.Stat(p)
	if err == nil {
		// file exists
		return p, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		// file may exist, but some other error occurred
		return "", err
	}
	// file does not exist
	return "", nil
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
		return nil, ErrVersion
	}

	return &c, nil
}

// Open tries to get a Config from a file name or path.
// If the name is empty, it returns the default configuration.
// If the name is invalid, it returns an error.
func Open(filename string) (*Config, error) {
	if filename == "" {
		return Default(), nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return Load(file)
}
