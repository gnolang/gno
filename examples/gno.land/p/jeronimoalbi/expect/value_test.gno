package expect_test

import (
	"errors"
	"testing"

	"gno.land/p/jeronimoalbi/expect"
)

func TestValue(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "foo").ToEqual("foo")
		expect.Value(t, []byte("foo")).ToEqual([]byte("foo"))
		expect.Value(t, stringer("foo")).ToEqual(stringer("foo"))
		expect.Value(t, true).ToEqual(true)
		expect.Value(t, float32(1)).ToEqual(float32(1))
		expect.Value(t, float64(1)).ToEqual(float64(1))
		expect.Value(t, uint(1)).ToEqual(uint(1))
		expect.Value(t, uint8(1)).ToEqual(uint8(1))
		expect.Value(t, uint16(1)).ToEqual(uint16(1))
		expect.Value(t, uint32(1)).ToEqual(uint32(1))
		expect.Value(t, uint64(1)).ToEqual(uint64(1))
		expect.Value(t, int(1)).ToEqual(int(1))
		expect.Value(t, int8(1)).ToEqual(int8(1))
		expect.Value(t, int16(1)).ToEqual(int16(1))
		expect.Value(t, int32(1)).ToEqual(int32(1))
		expect.Value(t, int64(1)).ToEqual(int64(1))
		expect.Value(t, errors.New("foo")).ToEqual(errors.New("foo"))
		expect.Value(t, errors.New("foo bar")).ToContainErrorString("foo")
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "foo").Not().ToEqual("bar")
		expect.Value(t, []byte("foo")).Not().ToEqual([]byte("bar"))
		expect.Value(t, stringer("foo")).Not().ToEqual(stringer("bar"))
		expect.Value(t, true).Not().ToEqual(false)
		expect.Value(t, float32(1)).Not().ToEqual(float32(2))
		expect.Value(t, float64(1)).Not().ToEqual(float64(2))
		expect.Value(t, uint(1)).Not().ToEqual(uint(2))
		expect.Value(t, uint8(1)).Not().ToEqual(uint8(2))
		expect.Value(t, uint16(1)).Not().ToEqual(uint16(2))
		expect.Value(t, uint32(1)).Not().ToEqual(uint32(2))
		expect.Value(t, uint64(1)).Not().ToEqual(uint64(2))
		expect.Value(t, int(1)).Not().ToEqual(int(2))
		expect.Value(t, int8(1)).Not().ToEqual(int8(2))
		expect.Value(t, int16(1)).Not().ToEqual(int16(2))
		expect.Value(t, int32(1)).Not().ToEqual(int32(2))
		expect.Value(t, int64(1)).Not().ToEqual(int64(2))
		expect.Value(t, errors.New("foo")).Not().ToEqual(errors.New("bar"))
		expect.Value(t, errors.New("foo")).Not().ToContainErrorString("bar")
	})

	t.Run("to be nil", func(t *testing.T) {
		t.Parallel()
		expect.Value(t, nil).ToBeNil()
		expect.Value(t, (*int)(nil)).ToBeNil() // typed nil
	})

	t.Run("not to be nil", func(t *testing.T) {
		t.Parallel()
		expect.Value(t, "").Not().ToBeNil()
	})

	t.Run("to be truthy", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "true").AsBoolean().ToBeTruthy()
		expect.Value(t, "TRUE").AsBoolean().ToBeTruthy()
		expect.Value(t, "t").AsBoolean().ToBeTruthy()
		expect.Value(t, "1").AsBoolean().ToBeTruthy()
		expect.Value(t, []byte("true")).AsBoolean().ToBeTruthy()
	})

	t.Run("not to be truthy", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "").AsBoolean().Not().ToBeTruthy()
		expect.Value(t, "false").AsBoolean().Not().ToBeTruthy()
		expect.Value(t, "FALSE").AsBoolean().Not().ToBeTruthy()
		expect.Value(t, "f").AsBoolean().Not().ToBeTruthy()
		expect.Value(t, "0").AsBoolean().Not().ToBeTruthy()
		expect.Value(t, []byte(nil)).AsBoolean().Not().ToBeTruthy()
		expect.Value(t, []byte("false")).AsBoolean().Not().ToBeTruthy()
	})

	t.Run("to be falsy", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "false").AsBoolean().ToBeFalsy()
		expect.Value(t, "FALSE").AsBoolean().ToBeFalsy()
		expect.Value(t, "f").AsBoolean().ToBeFalsy()
		expect.Value(t, "0").AsBoolean().ToBeFalsy()
		expect.Value(t, "").AsBoolean().ToBeFalsy()
		expect.Value(t, []byte(nil)).AsBoolean().ToBeFalsy()
	})

	t.Run("not to be falsy", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, "true").AsBoolean().Not().ToBeFalsy()
		expect.Value(t, "TRUE").AsBoolean().Not().ToBeFalsy()
		expect.Value(t, "t").AsBoolean().Not().ToBeFalsy()
		expect.Value(t, "1").AsBoolean().Not().ToBeFalsy()
		expect.Value(t, []byte("true")).AsBoolean().Not().ToBeFalsy()
	})

	t.Run("to equal stringer", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, address("foo")).AsString().ToEqual("foo")
	})

	t.Run("not to equal stringer", func(t *testing.T) {
		t.Parallel()

		expect.Value(t, address("foo")).AsString().Not().ToEqual("bar")
	})
}

type stringer string

func (s stringer) String() string { return string(s) }
