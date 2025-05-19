package expect

import "testing"

func TestValue(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		t.Parallel()

		Value(t, "foo").ToEqual("foo")
		Value(t, []byte("foo")).ToEqual([]byte("foo"))
		Value(t, true).ToEqual(true)
		Value(t, float32(1)).ToEqual(float32(1))
		Value(t, float64(1)).ToEqual(float64(1))
		Value(t, uint(1)).ToEqual(uint(1))
		Value(t, uint8(1)).ToEqual(uint8(1))
		Value(t, uint16(1)).ToEqual(uint16(1))
		Value(t, uint32(1)).ToEqual(uint32(1))
		Value(t, uint64(1)).ToEqual(uint64(1))
		Value(t, int(1)).ToEqual(int(1))
		Value(t, int8(1)).ToEqual(int8(1))
		Value(t, int16(1)).ToEqual(int16(1))
		Value(t, int32(1)).ToEqual(int32(1))
		Value(t, int64(1)).ToEqual(int64(1))
	})

	t.Run("not to equal", func(t *testing.T) {
		t.Parallel()

		Value(t, "foo").Not().ToEqual("bar")
		Value(t, []byte("foo")).Not().ToEqual([]byte("bar"))
		Value(t, true).Not().ToEqual(false)
		Value(t, float32(1)).Not().ToEqual(float32(2))
		Value(t, float64(1)).Not().ToEqual(float64(2))
		Value(t, uint(1)).Not().ToEqual(uint(2))
		Value(t, uint8(1)).Not().ToEqual(uint8(2))
		Value(t, uint16(1)).Not().ToEqual(uint16(2))
		Value(t, uint32(1)).Not().ToEqual(uint32(2))
		Value(t, uint64(1)).Not().ToEqual(uint64(2))
		Value(t, int(1)).Not().ToEqual(int(2))
		Value(t, int8(1)).Not().ToEqual(int8(2))
		Value(t, int16(1)).Not().ToEqual(int16(2))
		Value(t, int32(1)).Not().ToEqual(int32(2))
		Value(t, int64(1)).Not().ToEqual(int64(2))
	})

	t.Run("to be nil", func(t *testing.T) {
		t.Parallel()
		Value(t, nil).ToBeNil()
	})

	t.Run("not to be nil", func(t *testing.T) {
		t.Parallel()
		Value(t, "").Not().ToBeNil()
	})
}
