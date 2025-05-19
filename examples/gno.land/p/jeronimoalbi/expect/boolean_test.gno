package expect

import "testing"

func TestBooleanChecker(t *testing.T) {
	t.Run("to be truthy", func(t *testing.T) {
		t.Parallel()

		Value(t, true).AsBoolean().ToBeTruthy()
		Value(t, "true").AsBoolean().ToBeTruthy()
		Value(t, "TRUE").AsBoolean().ToBeTruthy()
		Value(t, "t").AsBoolean().ToBeTruthy()
		Value(t, "1").AsBoolean().ToBeTruthy()
		Value(t, []byte("foo")).AsBoolean().ToBeTruthy()
	})

	t.Run("not to be truthy", func(t *testing.T) {
		t.Parallel()

		Value(t, false).AsBoolean().Not().ToBeTruthy()
		Value(t, "false").AsBoolean().Not().ToBeTruthy()
		Value(t, "FALSE").AsBoolean().Not().ToBeTruthy()
		Value(t, "f").AsBoolean().Not().ToBeTruthy()
		Value(t, "0").AsBoolean().Not().ToBeTruthy()
		Value(t, []byte(nil)).AsBoolean().Not().ToBeTruthy()
	})

	t.Run("to be falsy", func(t *testing.T) {
		t.Parallel()

		Value(t, false).AsBoolean().ToBeFalsy()
		Value(t, "false").AsBoolean().ToBeFalsy()
		Value(t, "FALSE").AsBoolean().ToBeFalsy()
		Value(t, "f").AsBoolean().ToBeFalsy()
		Value(t, "0").AsBoolean().ToBeFalsy()
		Value(t, "").AsBoolean().ToBeFalsy()
		Value(t, []byte(nil)).AsBoolean().ToBeFalsy()
	})

	t.Run("not to be falsy", func(t *testing.T) {
		t.Parallel()

		Value(t, true).AsBoolean().Not().ToBeFalsy()
		Value(t, "true").AsBoolean().Not().ToBeFalsy()
		Value(t, "TRUE").AsBoolean().Not().ToBeFalsy()
		Value(t, "t").AsBoolean().Not().ToBeFalsy()
		Value(t, "1").AsBoolean().Not().ToBeFalsy()
		Value(t, []byte("foo")).AsBoolean().Not().ToBeFalsy()
	})
}
