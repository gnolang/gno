// Package counter is a sub-component of /r/docs/composition.
// It holds a counter that only the composition realm can change.
package counter

import (
	"strconv"
)

var count int64

// allowedCaller is the only realm allowed to modify the counter
const allowedCaller = "gno.land/r/docs/composition"

// Increment increases the counter by 1 (restricted to composition realm)
func Increment(cur realm) {
	assertAuthorized(cur)
	count++
}

// Decrement decreases the counter by 1 (restricted to composition realm)
func Decrement(cur realm) {
	assertAuthorized(cur)
	count--
}

// Reset sets the counter back to zero (restricted to composition realm)
func Reset(cur realm) {
	assertAuthorized(cur)
	count = 0
}

// assertAuthorized ensures only the composition realm can modify state
func assertAuthorized(cur realm) {
	// cur.IsCurrent() must hold before we trust cur.Previous() to name the
	// caller; otherwise a non-crossing call could pass a non-current realm handle.
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	caller := cur.Previous()
	if caller.PkgPath() != allowedCaller {
		panic("unauthorized: only " + allowedCaller + " can modify counter")
	}
}

// GetCountStr returns the current count as a string
func GetCountStr() string {
	return strconv.FormatInt(count, 10)
}

// RenderValue returns a markdown snippet showing the current value
func RenderValue() string {
	return "**" + strconv.FormatInt(count, 10) + "**"
}

func Render(_ string) string {
	return "# Counter Component\n\nCurrent value: " + RenderValue() + "\n\nThis is a sub-component of [/r/docs/composition](/r/docs/composition)."
}
