package testutils

// For testing std.CallerAt().
func WrapCall(fn func()) {
	fn()
}
