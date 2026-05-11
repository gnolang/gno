package testutils

// WrapCall adds a frame to the call stack, for testing call stack depth.
func WrapCall(fn func()) {
	fn()
}
