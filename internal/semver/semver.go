package semver

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Semver struct {
	Major      uint
	Minor      uint
	Patch      uint
	Prerelease []string
	Build      []string
}

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

func (v *Semver) NextMajor() *Semver {
	return &Semver{
		Major: v.Major + 1,
	}
}

func (v *Semver) NextMinor() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor + 1,
	}
}

func (v *Semver) NextPatch() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

func (v *Semver) NextRelease() *Semver {
	return &Semver{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

func (v *Semver) IsStable() bool {
	return v.Major > 0 && len(v.Prerelease) == 0
}
