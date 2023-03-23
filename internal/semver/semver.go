// Package semver implements the [Semantic Versioning] standard.
// It provides utilities for parsing, comparing, and incrementing version numbers.
//
// [Semantic Versioning]: https://semver.org/
package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Semver struct {
	Major uint
	Minor uint
	Patch uint

	// One or more prerelease identifiers. Nil if this is a normal release.
	Prerelease []string

	// One or more build metadata identifiers. Nil if not provided.
	// Most operations, including comparison, will ignore this field.
	Build []string
}

// ErrSemver indicates a malformed version string.
var ErrSemver = errors.New("invalid semantic version specifier")

// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
var semverPattern = regexp.MustCompile(`^` +
	`(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
	`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)` +
	`(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+` +
	`(?:\.[0-9a-zA-Z-]+)*))?` +
	`$`)

func mustUint(s string) uint {
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Panicf("%v", err)
	}
	if val < 0 {
		log.Panicf("expected uint: %d", val)
	}
	return uint(val)
}

// Parse converts a string to a Semver object.
// If the string is not a valid version specifier, it returns [ErrSemver].
func Parse(s string) (*Semver, error) {
	match := semverPattern.FindStringSubmatch(s)
	if match == nil {
		return nil, ErrSemver
	}

	v := &Semver{
		Major: mustUint(match[semverPattern.SubexpIndex("major")]),
		Minor: mustUint(match[semverPattern.SubexpIndex("minor")]),
		Patch: mustUint(match[semverPattern.SubexpIndex("patch")]),
	}

	prerelease := match[semverPattern.SubexpIndex("prerelease")]
	if prerelease != "" {
		v.Prerelease = strings.Split(prerelease, ".")
	}

	build := match[semverPattern.SubexpIndex("buildmetadata")]
	if build != "" {
		v.Build = strings.Split(build, ".")
	}

	return v, nil
}

// String returns the textual representation of the version object,
// in the format:
//
//	Major.Minor.Patch[-Prerelease][+Build]
func (v *Semver) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch))
	if v.Prerelease != nil {
		s.WriteString("-")
		s.WriteString(strings.Join(v.Prerelease, "."))
	}
	if v.Build != nil {
		s.WriteString("+")
		s.WriteString(strings.Join(v.Build, "."))
	}
	return s.String()
}

func compareIdent(a, b string) int {
	x, errX := strconv.Atoi(a)
	y, errY := strconv.Atoi(b)

	if errX == nil && errY == nil {
		if x < y {
			return -1
		} else if x > y {
			return 1
		}
		return 0
	} else if errX == nil {
		return -1
	} else if errY == nil {
		return 1
	}

	return strings.Compare(a, b)
}

// Compare checks if two version specifiers are different. It returns:
//   - -1 if the first version specifier has lower precedence
//   - 0 if the version specifiers are equal
//   - 1 if the first version specifier has higher precedence
//
// Build metadata is ignored.
func (v *Semver) Compare(other *Semver) int {
	if v.Major < other.Major {
		return -1
	} else if v.Major > other.Major {
		return 1
	}

	if v.Minor < other.Minor {
		return -1
	} else if v.Minor > other.Minor {
		return 1
	}

	if v.Patch < other.Patch {
		return -1
	} else if v.Patch > other.Patch {
		return 1
	}

	n1 := len(v.Prerelease)
	n2 := len(other.Prerelease)

	if n1 == 0 && n2 == 0 {
		return 0
	} else if n1 > 0 && n2 == 0 {
		return -1
	} else if n1 == 0 && n2 > 0 {
		return 1
	}

	var n int
	if n1 < n2 {
		n = n1
	} else {
		n = n2
	}

	for i := 0; i < n; i++ {
		result := compareIdent(v.Prerelease[i], other.Prerelease[i])
		if result != 0 {
			return result
		}
	}

	if n1 < n2 {
		return -1
	} else if n1 > n2 {
		return 1
	}

	return 0
}

// NextMajor returns a new Semver object representing the next major version
// in the sequence.
func (v *Semver) NextMajor() *Semver {
	return &Semver{
		Major: v.Major + 1,
	}
}

// NextMinor returns a new Semver object representing the next minor version
// in the sequence.
func (v *Semver) NextMinor() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor + 1,
	}
}

// NextPatch returns a new Semver object representing the next patch version
// in the sequence.
func (v *Semver) NextPatch() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

// NextRelease returns a new Semver object with prerelease and build
// information stripped. (It is used to determine the next stable release
// on a prerelease branch.)
func (v *Semver) NextRelease() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

// IsStable returns true if the version is not a prerelease, and the major
// version number is not 0. (Major version 0 is used for initial development).
func (v *Semver) IsStable() bool {
	return v.Major > 0 && len(v.Prerelease) == 0
}
