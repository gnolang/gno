package funcpkg

// Top is a top-level function declared in this realm.
func Top() {}

// GetFn is a crossing function returning a function value declared here.
func GetFn(cur realm) func() {
	return Top
}
