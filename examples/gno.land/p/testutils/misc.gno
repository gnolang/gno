package testutils

// For testing std.GetCallerAt().
func WrapCall(fn func()) {
	fn()
}

// for testing access.
// NOTE: non-package variables cannot be overridden.
var TestVar1 int
var testVar2 int

func init() {
	TestVar1 = 123
	testVar2 = 456
}
