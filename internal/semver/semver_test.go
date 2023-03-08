package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		str string
		ver *Semver
	}{
		{"0.0.0", &Semver{}},
		{"0.0.1", &Semver{Major: 0, Minor: 0, Patch: 1}},
		{"0.1.0", &Semver{Major: 0, Minor: 1, Patch: 0}},
		{"1.0.0", &Semver{Major: 1, Minor: 0, Patch: 0}},
		{"123.456.789", &Semver{Major: 123, Minor: 456, Patch: 789}},

		{"0.0.0-0", &Semver{Prerelease: []string{"0"}}},
		{"0.0.0-alpha", &Semver{Prerelease: []string{"alpha"}}},
		{"0.0.0-alpha.0", &Semver{Prerelease: []string{"alpha", "0"}}},
		{"0.0.0-ALPHA", &Semver{Prerelease: []string{"ALPHA"}}},
		{"0.0.0-0abc", &Semver{Prerelease: []string{"0abc"}}},

		{"0.0.0+0", &Semver{Build: []string{"0"}}},
		{"0.0.0+alpha", &Semver{Build: []string{"alpha"}}},
		{"0.0.0+alpha.0", &Semver{Build: []string{"alpha", "0"}}},
		{"0.0.0+ALPHA", &Semver{Build: []string{"ALPHA"}}},
		{"0.0.0+00000", &Semver{Build: []string{"00000"}}},

		{"0.0.0-0+0", &Semver{Prerelease: []string{"0"}, Build: []string{"0"}}},

		{"1.2.3-a.b+c.d", &Semver{Major: 1, Minor: 2, Patch: 3,
			Prerelease: []string{"a", "b"}, Build: []string{"c", "d"}}},
	}

	for _, test := range tests {
		t.Run(test.str, func(t *testing.T) {
			v, err := Parse(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ver, v)
		})
	}

	tests2 := []struct {
		str string
	}{
		{""},
		{".."},
		{"0"},
		{"0.0"},
		{"0.0.00"},
		{"0.0.0-"},
		{"0.0.0-00"},
		{"0.0.0+"},
		{"v0.0.0"},
	}

	for _, test := range tests2 {
		t.Run(test.str, func(t *testing.T) {
			v, err := Parse(test.str)
			assert.Equal(t, ErrSemver, err)
			assert.Nil(t, v)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		ver *Semver
		str string
	}{
		{&Semver{}, "0.0.0"},
		{&Semver{Major: 1, Minor: 2, Patch: 3}, "1.2.3"},
		{&Semver{Prerelease: []string{"beta", "0"}}, "0.0.0-beta.0"},
		{&Semver{Build: []string{"beta", "0"}}, "0.0.0+beta.0"},

		{&Semver{Major: 1, Minor: 2, Patch: 3, Prerelease: []string{"a"}, Build: []string{"b"}},
			"1.2.3-a+b"},
	}

	for _, test := range tests {
		t.Run(test.str, func(t *testing.T) {
			assert.Equal(t, test.str, test.ver.String())
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		description string
		a           *Semver
		b           *Semver
	}{
		{
			description: "it returns zero for equal versions",
			a:           &Semver{},
			b:           &Semver{},
		},
		{
			description: "it checks for equal prerelease versions",
			a:           &Semver{Prerelease: []string{"alpha", "0"}},
			b:           &Semver{Prerelease: []string{"alpha", "0"}},
		},
		{
			description: "it ignores the build metadata",
			a:           &Semver{Build: []string{"asdf"}},
			b:           &Semver{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, 0, test.a.Compare(test.b))
		})
	}

	tests2 := []struct {
		description string
		a           *Semver
		b           *Semver
	}{
		{
			description: "it checks for different major versions",
			a:           &Semver{},
			b:           &Semver{Major: 1},
		},
		{
			description: "it checks for different minor versions",
			a:           &Semver{},
			b:           &Semver{Minor: 1},
		},
		{
			description: "it checks for different patch versions",
			a:           &Semver{},
			b:           &Semver{Patch: 1},
		},
		{
			description: "the major version number is most significant",
			a:           &Semver{Minor: 999, Patch: 999},
			b:           &Semver{Major: 1},
		},
		{
			description: "the minor version number  next most significant",
			a:           &Semver{Patch: 999},
			b:           &Semver{Minor: 1},
		},
		{
			description: "a prerelease has lower precedence than a stable version",
			a:           &Semver{Prerelease: []string{"alpha"}},
			b:           &Semver{},
		},
		{
			description: "numerical prereleases are compared according to value",
			a:           &Semver{Prerelease: []string{"0", "999"}},
			b:           &Semver{Prerelease: []string{"1"}},
		},
		{
			description: "prerelease names are compared according to string value",
			a:           &Semver{Prerelease: []string{"alpha"}},
			b:           &Semver{Prerelease: []string{"beta"}},
		},
		{
			description: "prerelease names are case sensitive",
			a:           &Semver{Prerelease: []string{"Beta"}},
			b:           &Semver{Prerelease: []string{"alpha"}},
		},
		{
			description: "a numeric identifier has lower precedence than a non-numeric one",
			a:           &Semver{Prerelease: []string{"1"}},
			b:           &Semver{Prerelease: []string{"a"}},
		},
		{
			description: "a shorter set of prerelease identifiers has lower precedence",
			a:           &Semver{Prerelease: []string{"alpha"}},
			b:           &Semver{Prerelease: []string{"alpha", "0"}},
		},
	}

	for _, test := range tests2 {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, -1, test.a.Compare(test.b))
			assert.Equal(t, 1, test.b.Compare(test.a))
		})
	}
}
