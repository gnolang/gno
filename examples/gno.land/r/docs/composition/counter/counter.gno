package counter

import (
	"chain/runtime"
	"strconv"
)

var count int64 = 0

// allowedCaller is the only realm allowed to modify the counter
const allowedCaller = "gno.land/r/docs/composition"

// Increment increases the counter by 1 (restricted to composition realm)
func Increment(_ realm) {
	assertAuthorized()
	count++
}

// Decrement decreases the counter by 1 (restricted to composition realm)
func Decrement(_ realm) {
	assertAuthorized()
	count--
}

// Reset sets the counter back to zero (restricted to composition realm)
func Reset(_ realm) {
	assertAuthorized()
	count = 0
}

// assertAuthorized ensures only the composition realm can modify state
func assertAuthorized() {
	caller := runtime.PreviousRealm()
	if caller.PkgPath() != allowedCaller {
		panic("unauthorized: only " + allowedCaller + " can modify counter")
	}
}

// GetCount returns the current count
func GetCount() int64 {
	return count
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
