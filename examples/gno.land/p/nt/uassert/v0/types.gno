package uassert

type TestingT interface {
	Helper()
	Skip(args ...any)
	Fatalf(fmt string, args ...any)
	Errorf(fmt string, args ...any)
	Logf(fmt string, args ...any)
	Fail()
	FailNow()
}
