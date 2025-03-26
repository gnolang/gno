// Inspired by: https://cs.opensource.google/go/x/tools/+/master:gopls/internal/test/integration/fake/glob/glob.go

package packages

import (
	"errors"
	"fmt"
	"strings"
)

var ErrAdjacentSlash = errors.New("** may only be adjacent to '/'")

// Glob patterns can have the following syntax:
//   - `*` to match one or more characters in a path segment
//   - `**` to match any number of path segments, including none
//
// Expanding on this:
//   - '/' matches one or more literal slashes.
//   - any other character matches itself literally.
type Glob struct {
	elems []element // pattern elements
}

// Parse builds a Glob for the given pattern, returning an error if the pattern
// is invalid.
func Parse(pattern string) (*Glob, error) {
	g, _, err := parse(pattern)
	return g, err
}

func parse(pattern string) (*Glob, string, error) {
	g := new(Glob)
	for len(pattern) > 0 {
		switch pattern[0] {
		case '/':
			// Skip consecutive slashes
			for len(pattern) > 0 && pattern[0] == '/' {
				pattern = pattern[1:]
			}
			g.elems = append(g.elems, slash{})

		case '*':
			if len(pattern) > 1 && pattern[1] == '*' {
				if (len(g.elems) > 0 && g.elems[len(g.elems)-1] != slash{}) || (len(pattern) > 2 && pattern[2] != '/') {
					return nil, "", ErrAdjacentSlash
				}
				pattern = pattern[2:]
				g.elems = append(g.elems, starStar{})
				break
			}
			pattern = pattern[1:]
			g.elems = append(g.elems, star{})

		default:
			pattern = g.parseLiteral(pattern)
		}
	}
	return g, "", nil
}

func (g *Glob) parseLiteral(pattern string) string {
	end := strings.IndexAny(pattern, "*/")
	if end == -1 {
		end = len(pattern)
	}
	g.elems = append(g.elems, literal(pattern[:end]))
	return pattern[end:]
}

func (g *Glob) String() string {
	var b strings.Builder
	for _, e := range g.elems {
		fmt.Fprint(&b, e)
	}
	return b.String()
}

func (g *Glob) StarFreeBase() string {
	var b strings.Builder
	for _, e := range g.elems {
		if e == (star{}) || e == (starStar{}) {
			break
		}
		fmt.Fprint(&b, e)
	}
	return b.String()
}

// element holds a glob pattern element, as defined below.
type element fmt.Stringer

// element types.
type (
	slash    struct{} // One or more '/' separators
	literal  string   // string literal, not containing / or *
	star     struct{} // *
	starStar struct{} // **
)

func (s slash) String() string    { return "/" }
func (l literal) String() string  { return string(l) }
func (s star) String() string     { return "*" }
func (s starStar) String() string { return "**" }

// Match reports whether the input string matches the glob pattern.
func (g *Glob) Match(input string) bool {
	return match(g.elems, input)
}

func match(elems []element, input string) (ok bool) {
	var elem interface{}
	for len(elems) > 0 {
		elem, elems = elems[0], elems[1:]
		switch elem := elem.(type) {
		case slash:
			// Skip consecutive slashes in the input
			if len(input) == 0 || input[0] != '/' {
				return false
			}
			for len(input) > 0 && input[0] == '/' {
				input = input[1:]
			}

		case starStar:
			// Special cases:
			//  - **/a matches "a"
			//  - **/ matches everything
			//
			// Note that if ** is followed by anything, it must be '/' (this is
			// enforced by Parse).
			if len(elems) > 0 {
				elems = elems[1:]
			}

			// A trailing ** matches anything.
			if len(elems) == 0 {
				return true
			}

			// Backtracking: advance pattern segments until the remaining pattern
			// elements match.
			for len(input) != 0 {
				if match(elems, input) {
					return true
				}
				_, input = split(input)
			}
			return false

		case literal:
			if !strings.HasPrefix(input, string(elem)) {
				return false
			}
			input = input[len(elem):]

		case star:
			var segInput string
			segInput, input = split(input)

			elemEnd := len(elems)
			for i, e := range elems {
				if e == (slash{}) {
					elemEnd = i
					break
				}
			}
			segElems := elems[:elemEnd]
			elems = elems[elemEnd:]

			// A trailing * matches the entire segment.
			if len(segElems) == 0 {
				if len(elems) > 0 && elems[0] == (slash{}) {
					elems = elems[1:] // shift elems
				}
				break
			}

			// Backtracking: advance characters until remaining subpattern elements
			// match.
			matched := false
			for i := range segInput {
				if match(segElems, segInput[i:]) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}

		default:
			panic(fmt.Sprintf("segment type %T not implemented", elem))
		}
	}

	return len(input) == 0
}

// split returns the portion before and after the first slash
// (or sequence of consecutive slashes). If there is no slash
// it returns (input, nil).
func split(input string) (first, rest string) {
	i := strings.IndexByte(input, '/')
	if i < 0 {
		return input, ""
	}
	first = input[:i]
	for j := i; j < len(input); j++ {
		if input[j] != '/' {
			return first, input[j:]
		}
	}
	return first, ""
}
