package gnolang

import (
	"testing"
	"unsafe"
)

func TestAllocSizes(t *testing.T) {
	t.Parallel()

	// go elemental
	println("_allocPointer", unsafe.Sizeof(&StructValue{}))
	println("_allocSlice", unsafe.Sizeof([]byte("12345678901234567890123456789012345678901234567890")))
	// gno types
	println("PointerValue{}", unsafe.Sizeof(PointerValue{}))
	println("StructValue{}", unsafe.Sizeof(StructValue{}))
	println("ArrayValue{}", unsafe.Sizeof(ArrayValue{}))
	println("SliceValue{}", unsafe.Sizeof(SliceValue{}))
	println("FuncValue{}", unsafe.Sizeof(FuncValue{}))
	println("MapValue{}", unsafe.Sizeof(MapValue{}))
	println("BoundMethodValue{}", unsafe.Sizeof(BoundMethodValue{}))
	println("Block{}", unsafe.Sizeof(Block{}))
	println("TypeValue{}", unsafe.Sizeof(TypeValue{}))
	println("TypedValue{}", unsafe.Sizeof(TypedValue{}))
	println("ObjectInfo{}", unsafe.Sizeof(ObjectInfo{}))
}

func TestAllocatorBytesForSlice(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		low      int
		high     int
		expected int64
	}{
		{
			name:     "ShortString",
			input:    "hello world",
			low:      0,
			high:     5,
			expected: allocStringRef,
		},
		{
			name:     "LongerString",
			input:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			low:      10,
			high:     50,
			expected: allocStringRef,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alloc := NewAllocator(1024 * 1024)

			tv := TypedValue{
				T: StringType,
				V: alloc.NewString(tc.input),
			}

			_, bytesBefore := alloc.Status()

			result := tv.GetSlice(alloc, tc.low, tc.high)
			_ = result.GetString()

			_, bytesAfter := alloc.Status()
			bytesUsed := bytesAfter - bytesBefore

			if bytesUsed != tc.expected {
				t.Fatalf("unexpected allocation: got %d bytes, want %d", bytesUsed, tc.expected)
			}
		})
	}

	t.Run("ChainedSlice", func(t *testing.T) {
		alloc := NewAllocator(1024 * 1024)
		input := "abcdefghijklmnopqrstuvwxyz"

		tv := TypedValue{
			T: StringType,
			V: alloc.NewString(input),
		}

		_, bytesBefore := alloc.Status()

		// Chain of slices
		s1 := tv.GetSlice(alloc, 0, 20)
		s2 := s1.GetSlice(alloc, 5, 15)
		s3 := s2.GetSlice(alloc, 2, 8)
		_ = s3.GetString()

		_, bytesAfter := alloc.Status()
		bytesUsed := bytesAfter - bytesBefore

		expected := int64(3) * allocStringRef
		if bytesUsed != expected {
			t.Fatalf("unexpected allocation: got %d bytes, want %d", bytesUsed, expected)
		}
	})
}
