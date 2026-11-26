// Package ctw exercises "cross to write": this realm's stored state changes
// only during a crossing call into it.
package ctw

import (
	"strings"

	"gno.land/p/demo/tests/libtable/v0"
)

var Count int

type Local struct{ N int }

func (l *Local) Inc() { l.N++ }

var (
	Saved   = &Local{N: 1}
	Pending *Local
)

// Bump is exported and not crossing: called from elsewhere it is read-only.
func Bump() { Count++ }

// Add is the crossing entry: its helpers and methods write.
func Add(cur realm, n int) {
	Count += n
	bump()
	Saved.Inc()
}

func bump() { Count++ }

// NewLocal constructs this realm's object; any caller may call it.
func NewLocal(n int) *Local { return &Local{N: n} }

// Render builds and changes only temporaries it creates.
func Render(path string) string {
	var sb strings.Builder
	row := struct{ k string }{}
	row.k = "rows"
	sb.WriteString(row.k)
	l := NewLocal(1)
	l.Inc()
	sb.WriteString(":")
	sb.WriteString(strings.Repeat("x", l.N))
	return sb.String()
}

// StartX keeps a new object; IncPending, not crossing, may not change it.
func StartX(cur realm) *Local {
	Pending = &Local{N: 1}
	return Pending
}

func IncPending() { Pending.Inc() }

// ForEach calls back a plain closure; ForEachX a crossing one.
func ForEach(cur realm, f func(n int))             { f(1) }
func ForEachX(cur realm, f func(cur realm, n int)) { f(cross(cur), 1) }

// Iter is a stateful iterator of this realm's type: a caller cannot advance it
// without crossing, since Next is this realm's code writing its object.
type Iter struct{ i, n int }

func NewIter(n int) *Iter { return &Iter{n: n} }

func (it *Iter) Next() bool {
	it.i++
	return it.i <= it.n
}

// Conversions of this realm's own stored values are reads: fine without crossing.
type (
	Bag  []int
	Rec  struct{ N int }
	Rec2 Rec
)

var (
	Rows = Bag{1, 2, 3}
	rec  = &Rec{N: 7}
)

func RowCount() int { return len([]int(Rows)) }
func RecN() int     { return (*Rec2)(rec).N }

// This realm's own callbacks, reached through library code its crossing call
// named: a closure in a package variable, a table entry, a reassigned local, a
// local bound once to a named function (that function by name), and the
// library's own table.
var Boxes = []*libtable.Box{{N: 1}, {N: 2}}

func ViaApply(cur realm)    { libtable.Apply(func() { Count++ }) }
func ViaTable(cur realm)    { libtable.Table["run"](func() { Count++ }) }
func ViaAlias(cur realm)    { each := libtable.Each; each(Boxes, func(b *libtable.Box) { b.N++ }) }
func ViaDispatch(cur realm) { libtable.Dispatch("inc", Boxes[0]) }

func ViaLocal(cur realm) {
	each := libtable.Each
	each = libtable.Each
	each(Boxes, func(b *libtable.Box) { b.N++ })
}

func BoxSum() (s int) {
	for _, b := range Boxes {
		s += b.N
	}
	return s
}

// Hook hands out a closure that writes this realm: a holder cannot run it
// with this realm's authority.
func Hook() func() { return func() { Count++ } }

// Boom panics inside this realm's read-only entry.
func Boom() { panic("boom") }

// RecoverValue catches a panic from a library function called through a value.
func RecoverValue(cur realm) (r any) {
	defer func() { r = recover() }()
	f := libtable.Panics
	f()
	return nil
}
