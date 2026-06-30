package copytest

var (
	Slice = []int{1, 2, 3}

	Array = [0]int{} // Empty array.

	Array1 = [1]int{1}

	// Both byte slices, the first uses List and the second Data.

	ByteSlice = []byte{1, 2, 3}

	DataByteSlice = []byte("alfa")
)

// NamedBytes is a defined type over []byte; the copy() byte fast path must
// resolve it through baseOf to the same Data-backed path as an unnamed []byte.
type NamedBytes []byte

var DataNamedByteSlice = NamedBytes("alfa")
