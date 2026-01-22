package gnolang

import "testing"

func TestAllocatorBytesForSlice(t *testing.T) {
	t.Run("ShortString", func(t *testing.T) {
		alloc := NewAllocator(1024 * 1024)
		input := "hello world"

		tv := TypedValue{
			T: StringType,
			V: alloc.NewString(input),
		}

		_, bytesBefore := alloc.Status()

		// perform slice
		result := tv.GetSlice(alloc, 0, 5)
		_ = result.GetString()

		_, bytesAfter := alloc.Status()
		bytesUsed := bytesAfter - bytesBefore

		t.Logf("Short string slice: allocated %d bytes for 5-char slice", bytesUsed)
	})

	t.Run("LongerString", func(t *testing.T) {
		alloc := NewAllocator(1024 * 1024)
		input := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

		tv := TypedValue{
			T: StringType,
			V: alloc.NewString(input),
		}

		_, bytesBefore := alloc.Status()

		result := tv.GetSlice(alloc, 10, 50)
		_ = result.GetString()

		_, bytesAfter := alloc.Status()
		bytesUsed := bytesAfter - bytesBefore

		t.Logf("Longer string slice: allocated %d bytes for 40-char slice", bytesUsed)
	})

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

		t.Logf("Chained slices (3 operations): allocated %d bytes total", bytesUsed)
	})
}
