package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
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

// Overridden by Attributes.String().
func (s Span) String() string {
	if s.Pos.Line == s.End.Line {
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

// Overridden by Attributes.String().
func (loc Location) String() string {
	return fmt.Sprintf("%s/%s:%s",
		loc.PkgPath,
		loc.File,
		loc.Span.String(),
	)
}

// Overridden by Attributes.IsZero().
// NOTE: DO NOT CHANGE.
func (loc Location) IsZero() bool {
	return loc == Location{}
}
