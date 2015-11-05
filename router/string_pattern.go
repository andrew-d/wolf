package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/context"
)

// StringPattern describes a parsed Sinatra-style string pattern.
type StringPattern struct {
	raw      string   // Raw (unparsed) pattern
	pats     []string // Name of each pattern (i.e. pats[0] in "/:foo/:bar" is "foo")
	breaks   []byte   // Break characters
	literals []string // Literal component before a pattern
	wildcard bool     // Has a wildcard match at the end?
}

func (s StringPattern) Prefix() string {
	return s.literals[0]
}

func (s StringPattern) Match(r *http.Request) bool {
	return s.match(r, nil, true)
}

func (s StringPattern) Run(r *http.Request, c *context.Context) {
	s.match(r, c, false)
}

func (s StringPattern) match(r *http.Request, c *context.Context, dryrun bool) bool {
	path := r.URL.Path

	var matches map[string]string

	// Only allocate when we're actually running the pattern - i.e. not when
	// we're just testing for a match.
	if !dryrun {
		if s.wildcard {
			matches = make(map[string]string, len(s.pats)+1)
		} else if len(s.pats) != 0 {
			matches = make(map[string]string, len(s.pats))
		}
	}

	for i, pat := range s.pats {
		// Get the literal that precedes this pattern, and verify that the path
		// starts with the literal.
		sli := s.literals[i]
		if !strings.HasPrefix(path, sli) {
			return false
		}
		path = path[len(sli):]

		m := 0
		bc := s.breaks[i]
		for ; m < len(path); m++ {
			if path[m] == bc || path[m] == '/' {
				break
			}
		}

		if m == 0 {
			// Empty strings are not matches, otherwise routes like
			// "/:foo" would match the path "/"
			return false
		}

		if !dryrun {
			matches[pat] = path[:m]
		}

		// Skip past this chunk
		path = path[m:]
	}

	// There's exactly one more literal than pat.
	tail := s.literals[len(s.pats)]
	if s.wildcard {
		// This last literal is everything before the wildcard, so the path
		// must start with it.
		if !strings.HasPrefix(path, tail) {
			return false
		}

		if !dryrun {
			matches["*"] = path[len(tail)-1:]
		}
	} else if path != tail {
		return false
	}

	// Don't modify the context if there isn't one, or if it's a dryrun.
	if c == nil || dryrun {
		return true
	}

	// Set URL parameters in the context
	*c = SetURLParams(*c, matches)
	return true
}

func (s StringPattern) String() string {
	return fmt.Sprintf("StringPattern(%q)", s.raw)
}

// "Break characters" are characters that can end patterns. They are not allowed
// to appear in pattern names. "/" was chosen because it is the standard path
// separator, and "." was chosen because it often delimits file extensions. ";"
// and "," were chosen because Section 3.3 of RFC 3986 suggests their use.
const bc = "/.;,"

var patternRe = regexp.MustCompile(`[` + bc + `]:([^` + bc + `]+)`)

// ParseStringPattern takes a Sinatra-style string pattern and decomposes it
// into its constituent components.
func ParseStringPattern(s string) StringPattern {
	raw := s

	// Check for wildcard matches, then trim the suffix if it's there.
	var wildcard bool
	if strings.HasSuffix(s, "/*") {
		s = s[:len(s)-1]
		wildcard = true
	}

	matches := patternRe.FindAllStringSubmatchIndex(s, -1)

	pats := make([]string, len(matches))
	breaks := make([]byte, len(matches))
	literals := make([]string, len(matches)+1)

	n := 0
	for i, match := range matches {
		a, b := match[2], match[3]
		literals[i] = s[n : a-1] // Need to leave off the colon
		pats[i] = s[a:b]

		// Break character at the end of the string is a '/', otherwise it's
		// the next character.
		if b == len(s) {
			breaks[i] = '/'
		} else {
			breaks[i] = s[b]
		}

		n = b
	}

	// Any remaining string is the last literal.
	literals[len(matches)] = s[n:]

	return StringPattern{
		raw:      raw,
		pats:     pats,
		breaks:   breaks,
		literals: literals,
		wildcard: wildcard,
	}
}
