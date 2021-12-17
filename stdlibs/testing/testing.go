// Shim for Go's "testing" package to support minimal testing types.
package testing

type T struct {
	name    string
	didFail bool
}

func NewT(name string) *T {
	return &T{name: name}
}

// Not yet implemented:
// func (t *T) Cleanup(f func()) {
// func (t *T) Deadline() (deadline time.Time, ok bool)
func (t *T) Error(args ...interface{}) {
	t.Log(args...)
	t.Fail()
}
func (t *T) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}
func (t *T) Fail() {
	t.didFail = true
}
func (t *T) FailNow() {
	panic("TEST FAILED")
}
func (t *T) Failed() bool {
	return t.didFail
}
func (t *T) Fatal(args ...interface{}) {
	t.Log(args...)
	t.FailNow()
}
func (t *T) Fatalf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.FailNow()
}
func (t *T) Log(args ...interface{}) {
	for _, arg := range args {
		print(arg)
	}
	println("")
}
func (t *T) Logf(format string, args ...interface{}) {
	println("format:", format, "args:")
	for _, arg := range args {
		print(arg)
	}
	println("")
}
func (t *T) Name() string {
	return t.name
}
func (t *T) Parallel() {
	// does nothing.
}
func (t *T) Run(name string, f func(t *T)) bool {
	panic("not yet implemented")
}
func (t *T) Setenv(key, value string) {
	panic("not yet implemented")
}
func (t *T) Skip(args ...interface{}) {
	t.Log(args...)
	t.SkipNow()
}
func (t *T) SkipNow() {
	panic("not yet implemented")
}
func (t *T) Skipf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.SkipNow()
}
func (t *T) TempDir() string {
	panic("not yet implemented")
}
