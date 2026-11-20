// Package nestclosure builds nested closures, for tests of closures built while
// the caller runs view-only.
package nestclosure

// Outer returns a closure that builds the writer when called.
func Outer() func() func([]int) {
	return func() func([]int) { return func(v []int) { v[0] = 55 } }
}

// Maker returns a writer built under the caller's realm.
func Maker() func([]int) { return func(v []int) { v[0] = 44 } }
