package uassert

import "testing"

func extractPanicErr(r interface{}) string {
	err, ok := r.(error)
	if ok {
		return err.Error()
	}

	errStr, ok := r.(string)
	if ok {
		return errStr
	}

	return "unknown error"
}

// PanicsWithError asserts that the code inside the specified func panics,
// and that the recovered panic value is an error that satisfies the given message
func PanicsWithError(t *testing.T, errString string, f func()) bool {
	t.Helper()

	var valid bool

	defer func() {
		if r := recover(); r != nil {
			// Check if the error matches
			panicErr := extractPanicErr(r)
			if panicErr == errString {
				valid = true

				return
			}

			t.Fatalf("Function panicked with err, %s", panicErr)
		}
	}()

	// Run the callback
	f()

	return valid
}

// NotPanics asserts that the code inside the specified func does NOT panic
func NotPanics(t *testing.T, f func()) bool {
	t.Helper()

	valid := true

	defer func() {
		if r := recover(); r != nil {
			valid = false

			t.Fatalf("Function panicked with err, %s", extractPanicErr(r))
		}
	}()

	// Run the callback
	f()

	return valid
}
