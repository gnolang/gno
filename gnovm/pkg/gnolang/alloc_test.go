package gnolang

import (
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
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

func TestAllocateOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "addition overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.bytes = math.MaxInt64 - 1
	alloc.Allocate(2) // Will cause overflow
}

func TestAllocateStringOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "addition overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateString(math.MaxInt64) // Will cause overflow
}

func TestAllocateBlockItemsOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "multiplication overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateBlockItems(math.MaxInt64 / 20) // Will cause overflow
}

func TestAllocateBlockOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "multiplication overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateBlock(math.MaxInt64 / 20) // Will cause overflow
}

func TestAllocateMapOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "multiplication overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateMap(math.MaxInt64 / 20) // Will cause overflow
}

func TestAllocateListArrayOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "multiplication overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateListArray(math.MaxInt64 / 20) // Will cause overflow
}

func TestAllocateStructFieldsOverflow(t *testing.T) {
	defer func() {
		panicMsg := recover()
		if panicMsg == nil {
			t.Errorf("Expected panic, got none")
		}
		require.Equal(t, "multiplication overflow", panicMsg)
	}()
	alloc := NewAllocator(math.MaxInt64)
	alloc.AllocateStructFields(math.MaxInt64 / 20) // Will cause overflow
}
