package expect

type context struct {
	t      TestingT
	not    bool
	prefix string
}

func (c context) CheckExpectation(ok bool, fn, fn2 func()) {
	// Check must be false to pass when "not" is specified
	if ok && c.not {
		fn()
	}

	// Check must be true to pass when "not" is not specified
	if !ok && !c.not {
		fn2()
	}
}

func (c context) Fail(msg string, args ...any) {
	if c.prefix != "" {
		msg = c.prefix + " - " + msg
	}

	c.t.Fatalf(msg, args...)
}
