package counter

import "strconv"

var count int64 = 0

// Increment increases the counter by 1
func Increment(_ realm) {
	count++
}

// Decrement decreases the counter by 1
func Decrement(_ realm) {
	count--
}

// GetCount returns the current count
func GetCount() int64 {
	return count
}

// GetCountStr returns the current count as a string
func GetCountStr() string {
	return strconv.FormatInt(count, 10)
}

// Reset sets the counter back to zero
func Reset(_ realm) {
	count = 0
}

// RenderValue returns a markdown snippet showing the current value
func RenderValue() string {
	return "**" + strconv.FormatInt(count, 10) + "**"
}

func Render(_ string) string {
	return "# Counter Component\n\nCurrent value: " + RenderValue() + "\n\nThis is a sub-component of [/r/docs/composition](/r/docs/composition)."
}
