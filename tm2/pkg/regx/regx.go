package regx

import (
	"fmt"
	"regexp"
	"strings"
)

type regx string // do not expose.

func (rx regx) String() string { return string(rx) }

func (rx regx) Compile() *compiledRegx {
	if !strings.HasPrefix(string(rx), `^`) {
		rx = `^` + rx
	}
	if !strings.HasSuffix(string(rx), `$`) {
		rx = rx + `$`
	}
	rr := regexp.MustCompile(string(rx))
	return &compiledRegx{rx, *rr}
}

type match struct {
	names   []string
	matches []string
}

func (mm match) Get(name string) string {
	for i, n := range mm.names {
		if name == n {
			return mm.matches[i]
		}
	}
	panic(fmt.Sprintf("no named subexpression %q", name))
}

type compiledRegx struct {
	src regx
	re  regexp.Regexp
}

// Matches a string by default, adding ^$ if not already present.
// If you need to match a part of the string, implement "Find()".
func (cr *compiledRegx) Match(s string) *match {
	// s is now canonical w/ ^$.
	matches := cr.re.FindStringSubmatch(s)
	if matches == nil {
		return nil
	}
	return &match{cr.re.SubexpNames(), matches}
}

// Returns true if matches.
func (cr *compiledRegx) Matches(s string) bool {
	return cr.re.MatchString(s)
}

func (cr *compiledRegx) Regx() regx { return cr.src }

func r2s(xx regx) string                         { return string(xx) }                     // regx -> string
func sj(sz ...string) string                     { return strings.Join(sz, ``) }           // string join
func sjd(dd string, sz ...string) string         { return strings.Join(sz, dd) }           // string join
func esc(ch string) string                       { return `\` + ch }                       // escape char (string)
func spl(ss string) []string                     { return strings.Split(ss, ``) }          // split string by char
func sra(ss string, oo string, nn string) string { return strings.ReplaceAll(ss, oo, nn) } // alias

func E(cs ...string) regx          { return regx(tmsa(sra, esc, sj(cs...), spl(`\^-].$*+?()[{|`))) } // escape everything
func C(cc regx) regx               { return `[` + cc + `]` }                                         // [char class]
func CN(cc regx) regx              { return `[^` + cc + `]` }                                        // [^NOT char class]
func S(xs ...regx) regx            { return G(xs...) + `*` }                                         // repeat 0 or more times, eager
func Sl(xs ...regx) regx           { return G(xs...) + `*?` }                                        // repeat 0 or more times, lazy
func P(xs ...regx) regx            { return G(xs...) + `+` }                                         // repeat 1 or more times, eager
func Pl(xs ...regx) regx           { return G(xs...) + `+?` }                                        // repeat 1 or more times, lazy
func M(xs ...regx) regx            { return G(xs...) + `?` }                                         // maybe, maybe not
func R(l, h int, xs ...regx) regx  { return G(xs...) + regx(fmt.Sprintf(`{%d,%d}`, l, h)) }          // repeat l~h times
func G(xs ...regx) regx            { return `(?:` + regx(sj(mab(r2s, xs)...)) + `)` }                // unnamed group
func N(nn string, xs ...regx) regx { return `(?P<` + regx(nn) + `>` + G(xs...) + `)` }               // named capture
func L(xs ...regx) regx            { return `^` + G(xs...) + `$` }                                   // line
func O(xs ...regx) regx            { return G(regx(sjd(`|`, mab(r2s, xs)...))) }                     // or

var (
	C_d regx = `\d` // Matches any digit (0-9).
	C_D regx = `\D` // Matches any non-digit character.
	C_w regx = `\w` // Matches any alphanumeric character plus "_" (word character).
	C_W regx = `\W` // Matches any non-word character.
	C_s regx = `\s` // Matches any whitespace character (space, tab, newline, etc.).
	C_S regx = `\S` // Matches any non-whitespace character.
)

var (
	C_alnum regx = `[:alnum:]`
	C_cntrl regx = `[:cntrl:]`
	C_lower regx = `[:lower:]`
	C_space regx = `[:space:]`
	C_alpha regx = `[:alpha:]`
	C_digit regx = `[:digit:]`
	C_print regx = `[:print:]`
	C_upper regx = `[:upper:]`
	C_blank regx = `[:blank:]`
	C_graph regx = `[:graph:]`
	C_punct regx = `[:punct:]`
	C_hexad regx = `[:xdigit:]`
)

// aka "reduce".
func fab[F func(A, B) B, A any, B any](f F, aa []A, b B) B {
	for _, a := range aa {
		b = f(a, b)
	}
	return b
}

// aka "map" as reduction.
func mab[M func(A) B, A any, B any](m M, aa []A) []B {
	return fab(func(a A, bb []B) []B { return append(bb, m(a)) }, aa, make([]B, 0, len(aa)))
}

// like "mab" but t(s,a,m(a)=b)=s.
// e.g. strings.ReplaceAll("U$A", x, esc(x)) where x in `!@#$`,
// t: strings.ReplaceAll
// m: esc
// s: "U$A"
// a: "!", "@", "#", "$"
// m(a): "\!", "\@", "\#", "\$"
func tmsa[T func(S, A, B) S, M func(A) B, S any, A any, B any](t T, m M, s S, aa []A) S {
	return fab(func(a A, s S) S { return t(s, a, m(a)) }, aa, s)
}
