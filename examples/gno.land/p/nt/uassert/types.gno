package uassert

type TestingT interface {
	Helper()
	Skip(args ...interface{})
	Fatalf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Logf(fmt string, args ...interface{})
	Fail()
	FailNow()
}
