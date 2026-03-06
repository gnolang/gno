package panictoerr_test

import (
	"errors"
	"testing"

	pte "gno.land/p/aeddi/panictoerr"
	"gno.land/p/nt/uassert/v0"
	grc20 "gno.land/r/demo/defi/grc20factory"
)

// Test PanicToError with different types as panic value.
func TestSimplePanicToError(t *testing.T) {
	err := pte.PanicToError(func() {
		panic("string")
	})
	uassert.Equal(t, err.Error(), "string")

	err = pte.PanicToError(func() {
		panic(errors.New("error"))
	})
	uassert.Equal(t, err.Error(), "error")

	err = pte.PanicToError(func() {
		panic(42)
	})
	uassert.Equal(t, err.Error(), "42")
}

func TestRealmPanicToError(t *testing.T) {
	// Set a test realm to be able to call a realm.
	testRealm := testing.NewCodeRealm("gno.land/r/aeddi/panictoerr/test")
	testing.SetRealm(testRealm)

	const message = "token instance does not exist"
	var err error

	// Define a panicking function (not crossing).
	panicking := func() {
		grc20.Bank("unknown")
	}

	// panicking function should panic.
	uassert.PanicsWithMessage(t, message, panicking)

	// panicking function should panic when wrapped in AbortToError.
	uassert.PanicsWithMessage(t, message, func() { pte.AbortToError(panicking) })

	// panicking function should not panic when wrapped in PanicToError.
	uassert.NotPanics(
		t,
		func() { err = pte.PanicToError(panicking) },
		"panicking function should not panic when wrapped in PanicToError",
	)
	uassert.Equal(t, err.Error(), message)

	// panicking function should not panic when wrapped in PanicAbortToError.
	uassert.NotPanics(
		t,
		func() { err = pte.PanicAbortToError(panicking) },
		"panicking function should not panic when wrapped in PanicAbortToError",
	)
	uassert.Equal(t, err.Error(), message)

	// Define an aborting function (crossing).
	aborting := func() {
		grc20.Faucet(cross, "unknown")
	}

	// aborting function should abort.
	uassert.AbortsWithMessage(t, message, aborting)

	// aborting function should abort when wrapped in PanicToError.
	uassert.AbortsWithMessage(t, message, func() { pte.PanicToError(aborting) })

	// aborting function should not abort when wrapped in AbortToError.
	uassert.NotAborts(
		t,
		func() { err = pte.AbortToError(aborting) },
		"aborting function should not abort when wrapped in AbortToError",
	)
	uassert.Equal(t, err.Error(), message)

	// aborting function should not abort when wrapped in PanicAbortToError.
	uassert.NotAborts(
		t,
		func() { err = pte.PanicAbortToError(aborting) },
		"aborting function should not abort when wrapped in PanicAbortToError",
	)
	uassert.Equal(t, err.Error(), message)
}
