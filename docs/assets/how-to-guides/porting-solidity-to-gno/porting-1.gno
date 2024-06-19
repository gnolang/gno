func shouldEqual(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %v(%T), got %v(%T)", expected, expected, got, got)
	}
}

func shouldErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, but got nil.")
	}
}

func shouldNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, but got err: %s.", err.Error())
	}
}

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should have panic")
		}
	}()
	f()
}

func shouldNoPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not have panic")
		}
	}()
	f()
}
