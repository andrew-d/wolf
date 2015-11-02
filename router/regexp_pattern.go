package router

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"regexp/syntax"

	"golang.org/x/net/context"
)

// RegexpPattern represents a Pattern obtained from a regexp.
type RegexpPattern struct {
	re     *regexp.Regexp
	prefix string
	names  []string
}

func (p RegexpPattern) Prefix() string {
	return p.prefix
}

func (p RegexpPattern) Match(r *http.Request, c *context.Context) bool {
	return p.match(r, c, false)
}

func (p RegexpPattern) Run(r *http.Request, c *context.Context) {
	p.match(r, c, false)
}

func (p RegexpPattern) match(r *http.Request, c *context.Context, dryrun bool) bool {
	matches := p.re.FindStringSubmatch(r.URL.Path)
	if matches == nil || len(matches) == 0 {
		return false
	}

	// If we have no context, it's a dryrun, or there are no named capture
	// groups (the `1` is because there's always the one matching group for the
	// regexp as a whole), then we don't need to continue.
	if c == nil || dryrun || len(matches) == 1 {
		return true
	}

	// Convert into a map of name --> match
	params := make(map[string]string, len(matches)-1)
	for i := 1; i < len(matches); i++ {
		params[p.names[i]] = matches[i]
	}

	*c = SetURLParams(*c, params)
	return true
}

func (p RegexpPattern) String() string {
	return fmt.Sprintf("RegexpPattern(%v)", p.re)
}

/*
I'm sorry, dear reader. I really am.

The problem here is to take an arbitrary regular expression and:
1. return a regular expression that is just like it, but left-anchored,
   preferring to return the original if possible.
2. determine a string literal prefix that all matches of this regular expression
   have, much like regexp.Regexp.Prefix(). Unfortunately, Prefix() does not work
   in the presence of anchors, so we need to write it ourselves.

What this actually means is that we need to sketch on the internals of the
standard regexp library to forcefully extract the information we want.

Unfortunately, regexp.Regexp hides a lot of its state, so our abstraction is
going to be pretty leaky. The biggest leak is that we blindly assume that all
regular expressions are perl-style, not POSIX. This is probably Mostly True, and
I think most users of the library probably won't be able to notice.
*/
func sketchOnRegex(re *regexp.Regexp) (*regexp.Regexp, string) {
	// Re-parse the regex from the string representation.
	rawRe := re.String()
	sRe, err := syntax.Parse(rawRe, syntax.Perl)
	if err != nil {
		// TODO: better way to warn?
		log.Printf("WARN(router): unable to parse regexp %v as perl. "+
			"This route might behave unexpectedly.", re)
		return re, ""
	}

	// Simplify and then compile the regex.
	sRe = sRe.Simplify()
	p, err := syntax.Compile(sRe)
	if err != nil {
		// TODO: better way to warn?
		log.Printf("WARN(router): unable to compile regexp %v. This "+
			"route might behave unexpectedly.", re)
		return re, ""
	}

	// If it's not left-anchored, we add that now.
	if p.StartCond()&syntax.EmptyBeginText == 0 {
		// I hope doing this is always legal...
		newRe, err := regexp.Compile(`\A` + rawRe)
		if err != nil {
			// TODO: better way to warn?
			log.Printf("WARN(router): unable to create a left-"+
				"anchored regexp from %v. This route might "+
				"behave unexpectedly", re)
			return re, ""
		}
		re = newRe
	}

	// We run the regular expression more or less by hand in order to calculate
	// the prefix.
	pc := uint32(p.Start)
	atStart := true
	i := &p.Inst[pc]
	var buf bytes.Buffer
OuterLoop:
	for {
		switch i.Op {

		// There's may be an 'empty' operation at the beginning of every regex,
		// due to OpBeginText.
		case syntax.InstEmptyWidth:
			if !atStart {
				break OuterLoop
			}

		// Captures and no-ops don't affect the prefix
		case syntax.InstCapture, syntax.InstNop:
			// nop!

		// We handle runes
		case syntax.InstRune, syntax.InstRune1, syntax.InstRuneAny,
			syntax.InstRuneAnyNotNL:

			atStart = false

			// If we don't have exactly one rune, or if the 'fold case' flag is
			// set, then we don't count this as part of the prefix.  Due to
			// unicode case-crazyness, it's too hard to deal with case
			// insensitivity...
			if len(i.Rune) != 1 ||
				syntax.Flags(i.Arg)&syntax.FoldCase != 0 {
				break OuterLoop
			}

			// Add to the prefix, continue.
			buf.WriteRune(i.Rune[0])

		// All other instructions may affect the prefix, so we continue.
		default:
			break OuterLoop
		}

		// Continue to the next instruction
		pc = i.Out
		i = &p.Inst[pc]
	}

	return re, buf.String()
}

// ParseRegexpPattern will turn the given Regexp into something that implements
// Pattern, possibly modifying it such that it is left-anchored.
func ParseRegexpPattern(re *regexp.Regexp) RegexpPattern {
	re, prefix := sketchOnRegex(re)
	rnames := re.SubexpNames()

	// We have to make our own copy since package regexp forbids us
	// from scribbling over the slice returned by SubexpNames().
	names := make([]string, len(rnames))
	for i, rname := range rnames {
		// If the group is un-named, we give it the special name '$X', where X
		// is a number.
		if rname == "" {
			rname = fmt.Sprintf("$%d", i)
		}
		names[i] = rname
	}

	return RegexpPattern{
		re:     re,
		prefix: prefix,
		names:  names,
	}
}
