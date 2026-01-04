package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	r "github.com/gnolang/gno/tm2/pkg/regx"
)

/* ========================================
```
<BlockNode>{
    Attributes{
	Label Name
	Span{
	    Pos{
                Line   int
		Column int
	    End:Pos{
		Line   int
		Column int
    StaticBlock{
	Loc:Location{
	    Span{ <copy of ../../Attributes/Span>
<Node>{
    Attributes{
	Label Name
	Span{
	    Pos{
                Line   int
		Column int
	    End:Pos{
		Line   int
		Column int
```
======================================== */

// Pos(isition)
type Pos struct {
	Line   int
	Column int
}

func (p Pos) GetPos() Pos {
	return p
}

func (p Pos) GetLine() int {
	return p.Line
}

func (p Pos) GetColumn() int {
	return p.Column
}

func (p Pos) Compare(p2 Pos) int {
	switch {
	case p.Line < p2.Line:
		return -1
	case p.Line == p2.Line:
		break
	case p.Line > p2.Line:
		return 1
	default:
		panic("should not happen")
	}
	switch {
	case p.Column < p2.Column:
		return -1
	case p.Column == p2.Column:
		return 0
	case p.Column > p2.Column:
		return 1
	default:
		panic("should not happen")
	}
}

// Overridden by Attributes.String().
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Overridden by Attributes.IsZero().
// NOTE: DO NOT CHANGE.
func (p Pos) IsZero() bool {
	return p == Pos{}
}

// ----------------------------------------
// Span
// Has a start (pos) and an end.
type Span struct {
	Pos // start
	End Pos
	Num int // positive if conflicts.
}

// Convenience with no changes.
func Span4(line, col, endLine, endCol int) Span {
	return Span{Pos: Pos{line, col}, End: Pos{endLine, endCol}}
}

func SpanFromGo(gofs *token.FileSet, gon ast.Node) Span {
	pos := gon.Pos()
	end := gon.End()
	posn := gofs.Position(pos)
	endn := gofs.Position(end)
	return Span{
		Pos: Pos{posn.Line, posn.Column},
		End: Pos{endn.Line, endn.Column},
	}
}

func (s Span) GetSpan() Span {
	return s
}

// If you need to update the span/pos/location of a node it should be re-parsed
// from an updated AST.  This is important because location is used as identity.
// Anyone with a node can still mutate these fields directly; the method guides.
// If you need to override the span (e.g. constructing a Location by mutating
// a copy) then call SetSpanOverride() instead of directly assigning to .Span.
func (s *Span) SetSpan(s2 Span) {
	if !s.IsZero() && (*s != s2) {
		panic(".Span can ony be set once. s:" + s.String() + " s2:" + s2.String())
	}
	*s = s2
}

// See documentation for SetSpan().
func (s *Span) SetSpanOverride(s2 Span) {
	*s = s2
}

func SpanFromMatch(line, col, endLine, endCol string) (span Span, err error) {
	if endLine == "" && endCol == "" {
		endLine = line
		endCol = col
	} else if endLine == "" {
		endLine = line
	}
	l, err := strconv.Atoi(line)
	if err != nil {
		return
	}
	c, err := strconv.Atoi(col)
	if err != nil {
		return
	}
	el, err := strconv.Atoi(endLine)
	if err != nil {
		return
	}
	ec, err := strconv.Atoi(endCol)
	if err != nil {
		return
	}
	span = Span4(l, c, el, ec)
	return
}

func ParseSpan(spanstr string) (span Span, err error) {
	match := ReSpan.Match(spanstr)
	if match == nil {
		return
	}
	line := match.Get("LINE")
	col := match.Get("COL")
	endLine := match.Get("ENDLINE")
	endCol := match.Get("ENDCOL")
	span, err = SpanFromMatch(line, col, endLine, endCol)
	return
}

// Overridden by Attributes.String().
func (s Span) String() string {
	if s.Pos.Line == s.End.Line {
		if s.Pos.Column == s.End.Column {
			return fmt.Sprintf("%d:%d%s",
				s.Pos.Line, s.Pos.Column,
				strings.Repeat("'", s.Num), // e.g. 1:1 or 3:45''
			)
		}
		return fmt.Sprintf("%d:%d-%d%s",
			s.Pos.Line, s.Pos.Column, s.End.Column,
			strings.Repeat("'", s.Num), // e.g. 1:1-12'
		)
	} else {
		return fmt.Sprintf("%s-%s%s",
			s.Pos.String(), s.End.String(),
			strings.Repeat("'", s.Num), // e.g. 1:1-3:4'''
		)
	}
}

// Overridden by Attributes.IsZero().
// NOTE: DO NOT CHANGE.
func (s Span) IsZero() bool {
	return s == Span{}
}

// Suitable for Node. Less means earlier / higher level.
// Start (.Pos) determines node order before .End/ .Num,
// then the end (greater means containing, thus sooner),
// then the num (smaller means containing, thus sooner).
func (s Span) Compare(s2 Span) int {
	switch s.Pos.Compare(s2.Pos) {
	case -1: // s.Pos < s2.Pos
		return -1
	case 0: // s.Pos == s2.Pos
		break
	case 1: // s.Pos > s2.Pos
		return 1
	default:
		panic("should not happen")
	}
	switch s.End.Compare(s2.End) {
	case -1: // s.End < s2.End
		return 1 // see comment
	case 0: // s.End == s2.End
		break
	case 1:
		return -1 // see comment
	default:
		panic("should not happen")
	}
	switch {
	case s.Num < s2.Num:
		return -1
	case s.Num == s2.Num:
		return 0
	case s.Num > s2.Num:
		return 1
	default:
		panic("should not happen")
	}
}

// Union() returns a span for a new containing node w/ .Num possibly negative.
// NOTE: Span union math based on lines and columns, not the same math as for
// 2D container boxes. 2D has cardinality of 2, and span has cardinality of 1.
// See Span.Compare() to see a quirk where a greater end can mean lesser span.
// (we assume our 3D world has a cardinality of 3 but what if it is really 2?)
func (s Span) Union(s2 Span) (res Span) {
	if s.Pos.Compare(s2.Pos) < 0 {
		res.Pos = s.Pos
	} else {
		res.Pos = s2.Pos
	}
	if s.End.Compare(s2.End) < 0 {
		res.End = s2.End
	} else {
		res.End = s.End
	}
	// Only when s == s2 does .Num get set.
	if s.Pos == s2.Pos && s.End == s2.End {
		res.Num = min(s.Num, s2.Num) - 1 // maybe < 0.
	} else {
		res.Num = 0 // starts with zero.
	}
	return
}

// ----------------------------------------
// Location
// A Location is also an identifier for nodes.
// BlockNodes have these, while all Nodes have only .Span.
// (.Span field is duplicated in Node.Attributes and BlockNode.Location)
type Location struct {
	PkgPath string
	File    string
	Span
}

// Convenience with no modifications.
func Location3(pkgPath string, fname string, span Span) Location {
	return Location{PkgPath: pkgPath, File: fname, Span: span}
}

func ParseLocation(locstr string) (loc Location, err error) {
	match := ReLocation.Match(locstr)
	if match == nil {
		return
	}
	ppath := match.Get("PATH")
	fname := match.Get("FILE")
	line := match.Get("LINE")
	col := match.Get("COL")
	endLine := match.Get("ENDLINE")
	endCol := match.Get("ENDCOL")
	span, err := SpanFromMatch(line, col, endLine, endCol)
	if err != nil {
		return
	}
	loc = Location{PkgPath: ppath, File: fname, Span: span}
	return
}

// Overridden by Attributes.String().
func (loc Location) String() string {
	if loc.File == "" {
		return fmt.Sprintf("%s:%s",
			loc.PkgPath,
			loc.Span.String(),
		)
	} else {
		return fmt.Sprintf("%s/%s:%s",
			loc.PkgPath,
			loc.File,
			loc.Span.String(),
		)
	}
}

// Overridden by Attributes.IsZero().
// NOTE: DO NOT CHANGE.
func (loc Location) IsZero() bool {
	return loc == Location{}
}

func (loc Location) GetLocation() Location {
	return loc
}

func (loc Location) GetFile() string {
	return loc.File
}

func (loc *Location) SetLocation(loc2 Location) {
	if !loc.IsZero() && (*loc != loc2) {
		panic(".Location can ony be set once. loc:" + loc.String() + " loc2:" + loc2.String())
	}
	*loc = loc2
}

// ----------------------------------------
// Regexp for pos/span/location and more
//
// Regexp conventions:
//   - Re_xxx is (an exposed) regexp pattern *string*.
//   - Re_xxx must be composed of one outer group, or wrapped in `(?:___)`.
//   - Re_xxxLine must start with `^` and end with `$`, otherwise not.
//   - ReXxx is a *compiled* *regexp.Regexp instance.
//   - ReXxx should generally add `^`/`$` unless Re_xxxLine.
//   - Same rules apply for re_xxx and reXxx for unexposed values.
//   - All groups must be like (?:___) or (?P<key>___).
//   - Capture group names should not repeat.
//   - Also see: tm2/pkg/regx.
//
// NOTE: ReErrorLine is a regex designed to parse error details from a string.  It
// extracts the file location, line number, and error message from a formatted
// error string for both Go and Gno. cmd/gno/common.go uses it.
// TODO: Write exhaustive tests.

// Usage:
//   - Re_pos.Match("123:45").Get("LINE")
//   - Re_location.Match("gno.land/r/some/realm/somefile.gno:123:45").Get("FILE")
var (
	Re_pos         = r.G(r.N("LINE", r.P(r.C_d)), `:`, r.N("COL", r.P(r.C_d)))
	Re_posish      = r.G(r.N("LINE", r.P(r.C_d)), r.M(`:`, r.N("COL", r.P(r.C_d))))
	Re_end         = r.G(r.M(r.N("ENDLINE", r.P(r.C_d)), `:`), r.N("ENDCOL", r.P(r.C_d)))
	Re_primes      = r.N("PRIMES", r.S("`"))
	Re_span        = r.G(r.N("POS", Re_pos), r.G(`-`, r.N("END", Re_end)), Re_primes)
	Re_spanish     = r.G(r.N("POS", Re_posish), r.M(`-`, r.N("END", Re_end), Re_primes))
	Re_location    = r.G(r.N("PATH", r.Pl(r.CN(`:`))), r.M(`/`, r.N("FILE", r.P(r.CN(r.E(`/:`))))), `:`, r.N("SPAN", Re_span))
	Re_locationish = r.G(r.N("PATH", r.Pl(r.CN(`:`))), r.M(`/`, r.N("FILE", r.P(r.CN(r.E(`/:`))))), `:`, r.N("SPAN", Re_spanish))
	Re_errorLine   = r.L(r.N("LOC", Re_locationish), r.M(`:`), r.S(` `), r.N("MSG", r.S(`.`)))

	// Compile at init to avoid runtime compilation.
	ReLocation  = Re_location.Compile()
	ReSpan      = Re_span.Compile()
	ReErrorLine = Re_errorLine.Compile()
)

/* Compare to...
const (
	Re_pos         = `(?:(?P<line>\d+):(?P<col>\d+))`                                               // exact (for gno).
	Re_posish      = `(?:(?P<line>\d+)(?::(?P<col>\d+))?)`                                          // relaxed: col optional.
	Re_end         = `(?:(?:(?P<endline>\d+):)?(?P<endcol>\d+))`                                    // exact (endline optional).
	Re_primes      = `(?P<primes>'*)`                                                               // exact for superimposed nodes.
	Re_span        = `(?:(?P<pos>` + Re_pos + `)(?:-(?P<end>` + Re_end + `))` + Re_primes + `)`     // exact for gno.Span.
	Re_spanish     = `(?:(?P<pos>` + Re_posish + `)(?:-(?P<end>` + Re_end + `)` + Re_primes + `)?)` // relajada: end optional.
	Re_location    = `(?:(?P<path>[^:]+)(?:/(?P<file>[^:]+))?:(?P<span>` + Re_span + `))`           // exact for gno.Location (but relaxed for path/file).
	Re_locationish = `(?:(?P<path>[^:]+)(?:/(?P<file>[^:]+))?:(?P<span>` + Re_spanish + `))`        // relaxed for Go & Gno.
	Re_errorLine   = `^(?P<loc>` + Re_locationish + `):? *(P:<msg>.*)$`                             // relaxed for Go & Gno error line.
)

// Usage:
//
// match := ReLocation.FindStringSubmatch(locstr)
// match[ReLocPathIndex]    --> path
// match[ReLocFileIndex]    --> file
// match[ReLocSpanIndex]    --> span
// match[ReLocLineIndex]    --> start line
// match[ReLocEndLineIndex] --> end line
// ...
var (
	ReSpan             = regexp.MustCompile(`^` + Re_span + `$`)
	ReSpanPosIndex     = ReSpan.SubexpIndex("POS")
	ReSpanLineIndex    = ReSpan.SubexpIndex("LINE")
	ReSpanColIndex     = ReSpan.SubexpIndex("COL")
	ReSpanEndIndex     = ReSpan.SubexpIndex("END")
	ReSpanEndLineIndex = ReSpan.SubexpIndex("ENDLINE")
	ReSpanEndColIndex  = ReSpan.SubexpIndex("ENDCOL")

	ReLocation        = regexp.MustCompile(`^` + Re_location + `$`)
	ReLocPathIndex    = ReLocation.SubexpIndex("PATH")
	ReLocFileIndex    = ReLocation.SubexpIndex("FILE")
	ReLocSpanIndex    = ReLocation.SubexpIndex("SPAN")
	ReLocPosIndex     = ReLocation.SubexpIndex("POS")
	ReLocLineIndex    = ReLocation.SubexpIndex("LINE")
	ReLocColIndex     = ReLocation.SubexpIndex("COL")
	ReLocEndIndex     = ReLocation.SubexpIndex("END")
	ReLocEndLineIndex = ReLocation.SubexpIndex("ENDLINE")
	ReLocEndColIndex  = ReLocation.SubexpIndex("ENDCOL")
	ReLocPrimesIndex  = ReLocation.SubexpIndex("PRIMES")

	ReErrorLine           = regexp.MustCompile(Re_errorLine)
	ReErrLineLocIndex     = ReErrorLine.SubexpIndex("LOC")
	ReErrLinePathIndex    = ReErrorLine.SubexpIndex("PATH")
	ReErrLineFileIndex    = ReErrorLine.SubexpIndex("FILE")
	ReErrLineSpanIndex    = ReErrorLine.SubexpIndex("SPAN")
	ReErrLinePosIndex     = ReErrorLine.SubexpIndex("POS")
	ReErrLineLineIndex    = ReErrorLine.SubexpIndex("LINE")
	ReErrLineColIndex     = ReErrorLine.SubexpIndex("COL")
	ReErrLineEndIndex     = ReErrorLine.SubexpIndex("END")
	ReErrLineEndLineIndex = ReErrorLine.SubexpIndex("ENDLINE")
	ReErrLineEndColIndex  = ReErrorLine.SubexpIndex("ENDCOL")
	ReErrLinePrimesIndex  = ReErrorLine.SubexpIndex("PRIMES")
	ReErrLineMsgIndex     = ReErrorLine.SubexpIndex("MSG")
)
*/
