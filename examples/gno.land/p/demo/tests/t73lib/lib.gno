// Package t73lib takes a realm value as an ordinary argument and reads it.
package t73lib

// Describe reads r for the length of the call only.
func Describe(label string, r realm) string {
	return label + ": " + r.PkgPath()
}
