// Shim for Go's "testing" package to support minimal testing types.
package testing

//----------------------------------------
// Top level functions

func Short() bool {
	return true // TODO configure somehow.
}

// Like AllocsPerRun() but returns an integer.
// TODO: actually compute allocations; for now return 0.
func AllocsPerRun2(runs int, f func()) (total int) {
	for i := 0; i < runs; i++ {
		f()
	}
	return 0
}

//----------------------------------------
// T

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

//----------------------------------------
// B
// TODO: actually implement

type B struct {
	N int
}

func (b *B) Cleanup(f func())                          { panic("not yet implemented") }
func (b *B) Error(args ...interface{})                 { panic("not yet implemented") }
func (b *B) Errorf(format string, args ...interface{}) { panic("not yet implemented") }
func (b *B) Fail()                                     { panic("not yet implemented") }
func (b *B) FailNow()                                  { panic("not yet implemented") }
func (b *B) Failed() bool                              { panic("not yet implemented") }
func (b *B) Fatal(args ...interface{})                 { panic("not yet implemented") }
func (b *B) Fatalf(format string, args ...interface{}) { panic("not yet implemented") }
func (b *B) Helper()                                   { panic("not yet implemented") }
func (b *B) Log(args ...interface{})                   { panic("not yet implemented") }
func (b *B) Logf(format string, args ...interface{})   { panic("not yet implemented") }
func (b *B) Name() string                              { panic("not yet implemented") }
func (b *B) ReportAllocs()                             { panic("not yet implemented") }
func (b *B) ReportMetric(n float64, unit string)       { panic("not yet implemented") }
func (b *B) ResetTimer()                               { panic("not yet implemented") }
func (b *B) Run(name string, f func(b *B)) bool        { panic("not yet implemented") }
func (b *B) RunParallel(body func(*PB))                { panic("not yet implemented") }
func (b *B) SetBytes(n int64)                          { panic("not yet implemented") }
func (b *B) SetParallelism(p int)                      { panic("not yet implemented") }
func (b *B) Setenv(key, value string)                  { panic("not yet implemented") }
func (b *B) Skip(args ...interface{})                  { panic("not yet implemented") }
func (b *B) SkipNow()                                  { panic("not yet implemented") }
func (b *B) Skipf(format string, args ...interface{})  { panic("not yet implemented") }
func (b *B) Skipped() bool                             { panic("not yet implemented") }
func (b *B) StartTimer()                               { panic("not yet implemented") }
func (b *B) StopTimer()                                { panic("not yet implemented") }
func (b *B) TempDir() string                           { panic("not yet implemented") }

//----------------------------------------
// PB
// TODO: actually implement

type PB struct {
}

func (pb *PB) Next() bool { panic("not yet implemented") }
