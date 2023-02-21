package commit

import (
	"errors"
	"regexp"
	"strings"
)

// Footer is a "token: value" or "token #value" pair.
type Footer struct {
	// Token is a word without whitespace, except for the special
	// "BREAKING CHANGE" token.
	Token string

	// Separator is either ": " or " #". It is not important when looking up
	// footers, but saving it allows us to reconstruct the original commit
	// message from this object.
	Separator string

	// Value is the text corresponding to the token. It may contain spaces
	// and newlines.
	Value string
}

var ErrFooterSep = errors.New("BREAKING CHANGE must be followed by a colon and space (: )")
var ErrFooterCaps = errors.New("BREAKING CHANGE token must be capitalized")

// IsBreakingChange checks whether this footer designates a breaking change.
// It returns an error if the breaking change footer is not formatted
// correctly according to the standard.
func (f *Footer) IsBreakingChange() (bool, error) {
	if f.Token == "BREAKING CHANGE" || f.Token == "BREAKING-CHANGE" {
		if f.Separator == ": " {
			return true, nil
		}
		return false, ErrFooterSep
	}
	normalizedToken := strings.ToLower(f.Token)
	if normalizedToken == "breaking change" || normalizedToken == "breaking-change" {
		return false, ErrFooterCaps
	}
	return false, nil
}

var footerPattern = regexp.MustCompile(`^` +
	`(?P<token>(?:BREAKING CHANGE|[^:\pZ\x09-\x0D\x{FEFF}]+))` +
	`(?P<separator>: | #)` +
	`(?P<value>.*)` +
	`$`)

// extractFooters parses footers from the lines of text that make up the
// final paragraph of the commit message. If no footers are detected,
// an empty slice is returned, indicating that the final paragraph is
// actually part of the commit body.
func extractFooters(lines []string) []Footer {
	footers := make([]Footer, 0, 5)
	var token string
	var separator string
	var value strings.Builder

	for _, line := range lines {
		match := footerPattern.FindStringSubmatch(line)
		if match == nil {
			if token == "" {
				// first line is not a footer -- abort
				// this allows us to distinguish a footers section
				// from the last paragraph of the body
				return []Footer{}
			} else {
				// continuation of previous footer
				// (conventional commits allows footer values to span multiple lines)
				value.WriteString("\n")
				value.WriteString(line)
			}
		} else {
			if token != "" {
				footers = append(footers, Footer{token, separator, value.String()})
			}
			token = match[footerPattern.SubexpIndex("token")]
			separator = match[footerPattern.SubexpIndex("separator")]
			value.Reset()
			value.WriteString(match[footerPattern.SubexpIndex("value")])
		}
	}

	if token != "" {
		footers = append(footers, Footer{token, separator, value.String()})
	}

	return footers
}
