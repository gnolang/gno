package testutils

// For testing std.GetCallerAt().
func WrapCall(fn func()) {
	fn()
}
