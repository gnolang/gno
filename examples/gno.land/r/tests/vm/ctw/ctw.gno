// Package ctw exercises "cross to write": this realm's stored state changes
// only during a crossing call into it.
package ctw

import "strings"

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
func ForEach(cur realm, f func(n int))              { f(1) }
func ForEachX(cur realm, f func(cur realm, n int)) { f(cross(cur), 1) }
