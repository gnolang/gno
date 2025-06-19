package panictoerr

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

// PanicToError executes a function that might panic and, if it does,
// catches the panic and converts it to an error.
func PanicToError(mightPanic func()) (err error) {
	// Catch any panic that might occur and convert it to an error.
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case string:
				err = errors.New(v)
			case error:
				err = v
			default:
				err = errors.New(ufmt.Sprint(v))
			}
		}
	}()

	// Execute the function that might panic.
	mightPanic()

	return nil
}
