package expect

import (
	"std"
	"testing"
)

func TestStringChecker(t *testing.T) {
	t.Run("to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, "foo").AsString().ToEqual("foo")
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()
		Value(t, "foo").AsString().Not().ToEqual("bar")
	})

	t.Run("to equal stringer", func(t *testing.T) {
		t.Parallel()
		Value(t, std.Address("foo")).AsString().ToEqual("foo")
	})

	t.Run("not to equal stringer", func(t *testing.T) {
		t.Parallel()
		Value(t, std.Address("foo")).AsString().Not().ToEqual("bar")
	})

	t.Run("to be empty", func(t *testing.T) {
		t.Parallel()
		Value(t, "").AsString().ToBeEmpty()
	})

	t.Run("not to be empty", func(t *testing.T) {
		t.Parallel()
		Value(t, "foo").AsString().Not().ToBeEmpty()
	})

	t.Run("same length", func(t *testing.T) {
		t.Parallel()
		Value(t, "foo").AsString().ToHaveLength(3)
	})

	t.Run("different length", func(t *testing.T) {
		t.Parallel()
		Value(t, "foo").AsString().Not().ToHaveLength(1)
	})
}
