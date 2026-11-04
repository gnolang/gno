package markdown

import (
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// linearIDs is a parser.IDs that produces the same ids as goldmark's default
// generator, in linear time.
//
// goldmark's default tries "-1", "-2", ... from 1 for every duplicate, so a
// page with n identical headings costs O(n^2): 10k "## a" lines took about
// 5 s to render. A realm's Render output is attacker-controlled and can be
// close to 1 MiB, so linearIDs remembers the next suffix to try for each base
// id. Ids are only ever added, never removed, so every suffix below that
// counter is still taken and skipping them yields exactly the default's
// result.
type linearIDs struct {
	values map[string]bool
	next   map[string]int
}

var _ parser.IDs = (*linearIDs)(nil)

func newLinearIDs() *linearIDs {
	return &linearIDs{values: map[string]bool{}, next: map[string]int{}}
}

// Generate implements parser.IDs.
func (s *linearIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	result := slugID(value, kind)
	base := string(result)
	if !s.values[base] {
		s.values[base] = true
		return result
	}
	i := s.next[base]
	if i == 0 {
		i = 1
	}
	for {
		candidate := base + "-" + strconv.Itoa(i)
		i++
		if !s.values[candidate] {
			s.values[candidate] = true
			s.next[base] = i
			return []byte(candidate)
		}
	}
}

// Put implements parser.IDs.
func (s *linearIDs) Put(value []byte) {
	s.values[string(value)] = true
}

// slugID turns a heading text into an id the way goldmark's default
// generator does: ASCII letters and digits are kept (lowercased), spaces,
// '-' and '_' become '-', everything else is dropped.
func slugID(value []byte, kind ast.NodeKind) []byte {
	value = util.TrimRightSpace(util.TrimLeftSpace(value))
	result := make([]byte, 0, len(value))
	for i := 0; i < len(value); {
		v := value[i]
		l := util.UTF8Len(v)
		i += int(l)
		if l != 1 {
			continue
		}
		switch {
		case util.IsAlphaNumeric(v):
			if 'A' <= v && v <= 'Z' {
				v += 'a' - 'A'
			}
			result = append(result, v)
		case util.IsSpace(v) || v == '-' || v == '_':
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		if kind == ast.KindHeading {
			return []byte("heading")
		}
		return []byte("id")
	}
	return result
}
