package expect_test

import (
	"errors"
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestFunction(t *testing.T) {
	t.Run("not to fail", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return nil
		}).Not().ToFail()
	})

	t.Run("to fail", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return errors.New("Foo")
		}).ToFail()
	})

	t.Run("to fail with mesasge", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return errors.New("Foo")
		}).ToFail().WithMessage("Foo")
	})

	t.Run("to fail with different message", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return errors.New("Bar")
		}).ToFail().Not().WithMessage("Foo")
	})

	t.Run("to fail with error", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return errors.New("Foo")
		}).ToFail().WithError(errors.New("Foo"))
	})

	t.Run("to fail with different error", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return errors.New("Bar")
		}).ToFail().Not().WithError(errors.New("Foo"))
	})

	t.Run("not to panic", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			return nil
		}).Not().ToPanic()
	})

	t.Run("to panic", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			panic("Foo")
		}).ToPanic()
	})

	t.Run("to panic with message", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			panic("Foo")
		}).ToPanic().WithMessage("Foo")
	})

	t.Run("to panich with different message", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() error {
			panic("Foo")
		}).ToPanic().Not().WithMessage("Bar")
	})

	t.Run("to return value", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() any {
			return "foo"
		}).ToReturn("foo")

		expect.Func(t, func() (any, error) {
			return "foo", nil
		}).ToReturn("foo")
	})

	t.Run("not to return value", func(t *testing.T) {
		t.Parallel()

		expect.Func(t, func() any {
			return "foo"
		}).Not().ToReturn("bar")

		expect.Func(t, func() (any, error) {
			return "foo", nil
		}).Not().ToReturn("bar")
	})
}
