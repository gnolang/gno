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
	println("PackageValue{}", unsafe.Sizeof(PackageValue{}))
}

func TestBlockGetShallowSize_WithRefNodeSource(t *testing.T) {
	t.Parallel()

	const numValues = 5
	normalBlock := &Block{
		Source: &FuncDecl{},
		Values: make([]TypedValue, numValues),
	}
	refNodeBlock := &Block{
		Source: RefNode{Location: Location{PkgPath: "gno.land/r/test/foo"}},
		Values: make([]TypedValue, numValues),
	}

	normalSize := normalBlock.GetShallowSize()
	refNodeSize := refNodeBlock.GetShallowSize()

	expectedRefNodeSize := normalSize + allocRefNode
	if refNodeSize != expectedRefNodeSize {
		t.Errorf("Block with RefNode source: GetShallowSize() = %d, want %d (normal %d + allocRefNode %d)",
			refNodeSize, expectedRefNodeSize, normalSize, allocRefNode)
	}
}

// TestAllocConstantsMatchActualSizes verifies that allocation constants
// match the actual sizes of the structs they represent.
func TestAllocConstantsMatchActualSizes(t *testing.T) {
	tests := []struct {
		name        string
		constant    int64
		actualSize  uintptr
		shouldMatch bool
	}{
		{"_allocPointerValue", _allocPointerValue, unsafe.Sizeof(PointerValue{}), true},
		{"_allocStructValue", _allocStructValue, unsafe.Sizeof(StructValue{}), true},
		{"_allocArrayValue", _allocArrayValue, unsafe.Sizeof(ArrayValue{}), true},
		{"_allocSliceValue", _allocSliceValue, unsafe.Sizeof(SliceValue{}), true},
		{"_allocFuncValue", _allocFuncValue, unsafe.Sizeof(FuncValue{}), true},
		{"_allocMapValue", _allocMapValue, unsafe.Sizeof(MapValue{}), true},
		{"_allocBoundMethodValue", _allocBoundMethodValue, unsafe.Sizeof(BoundMethodValue{}), true},
		{"_allocBlock", _allocBlock, unsafe.Sizeof(Block{}), true},
		{"_allocPackageValue", _allocPackageValue, unsafe.Sizeof(PackageValue{}), true},
		{"_allocTypeValue", _allocTypeValue, unsafe.Sizeof(TypeValue{}), true},
		{"_allocTypedValue", _allocTypedValue, unsafe.Sizeof(TypedValue{}), true},
	}

	hasFailures := false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldMatch {
				if tt.constant != int64(tt.actualSize) {
					t.Errorf("%s: constant=%d, actual size=%d (OFF by %d)",
						tt.name, tt.constant, tt.actualSize, int64(tt.actualSize)-tt.constant)
					hasFailures = true
				} else {
					t.Logf("%s: %d ✓ MATCH", tt.name, tt.constant)
				}
			}
		})
	}

	if hasFailures {
		t.Error("Some allocation constants do not match actual struct sizes")
	}
}

// TestPackageValueGetShallowSize tests the GetShallowSize calculation
// for PackageValue with various field configurations.
func TestPackageValueGetShallowSize(t *testing.T) {
	tests := []struct {
		name     string
		pv       *PackageValue
		expected int64
	}{
		{
			name: "uverse package (should return 0)",
			pv: &PackageValue{
				PkgPath: ".uverse",
			},
			expected: 0,
		},
		{
			name: "minimal package",
			pv: &PackageValue{
				PkgName: "test",
				PkgPath: "gno.land/p/test",
			},
			expected: allocPackage +
				allocString + int64(len("test")) + // PkgName
				allocString + int64(len("gno.land/p/test")), // PkgPath
		},
		{
			name: "package with FNames",
			pv: &PackageValue{
				PkgName: "demo",
				PkgPath: "gno.land/r/demo",
				FNames:  []string{"file1.gno", "file2.gno"},
			},
			expected: allocPackage +
				allocString + int64(len("demo")) +
				allocString + int64(len("gno.land/r/demo")) +
				fileBlockEntrySize("file1.gno") + // FNames[0]
				fileBlockEntrySize("file2.gno"), // FNames[1]
		},
		{
			name: "package with fBlocksMap (no FNames)",
			pv: &PackageValue{
				PkgName: "demo",
				PkgPath: "gno.land/r/demo",
				fBlocksMap: map[string]*Block{
					"key1": nil,
					"key2": nil,
				},
			},
			// fBlocksMap is derived from FNames; with no FNames,
			// the map entries are not counted.
			expected: allocPackage +
				allocString + int64(len("demo")) +
				allocString + int64(len("gno.land/r/demo")),
		},
		{
			name: "package with all fields",
			pv: &PackageValue{
				PkgName: "demo",
				PkgPath: "gno.land/r/demo",
				FNames:  []string{"file1.gno", "file2.gno"},
				FBlocks: make([]Value, 2),
				fBlocksMap: map[string]*Block{
					"file1.gno": nil,
					"file2.gno": nil,
				},
			},
			expected: allocPackage +
				allocString + int64(len("demo")) +
				allocString + int64(len("gno.land/r/demo")) +
				fileBlockEntrySize("file1.gno") + // FNames[0] (includes FBlocks + fBlocksMap)
				fileBlockEntrySize("file2.gno"), // FNames[1] (includes FBlocks + fBlocksMap)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := tt.pv.GetShallowSize()
			if size != tt.expected {
				t.Errorf("GetShallowSize() = %d, expected %d (diff: %d)",
					size, tt.expected, size-tt.expected)
			}
		})
	}
}

// TestOtherValueTypesGetShallowSize tests GetShallowSize for other value types.
func TestOtherValueTypesGetShallowSize(t *testing.T) {
	t.Run("StringValue", func(t *testing.T) {
		sv := StringValue("hello world")
		expected := allocString + int64(len("hello world"))
		if size := sv.GetShallowSize(); size != expected {
			t.Errorf("StringValue.GetShallowSize() = %d, expected %d", size, expected)
		}
	})

	t.Run("ArrayValue with Data", func(t *testing.T) {
		av := &ArrayValue{Data: make([]byte, 100)}
		expected := int64(allocArray + 100)
		if size := av.GetShallowSize(); size != expected {
			t.Errorf("ArrayValue.GetShallowSize() = %d, expected %d", size, expected)
		}
	})

	t.Run("ArrayValue with List", func(t *testing.T) {
		av := &ArrayValue{List: make([]TypedValue, 5)}
		expected := int64(allocArray + allocArrayItem*5)
		if size := av.GetShallowSize(); size != expected {
			t.Errorf("ArrayValue.GetShallowSize() = %d, expected %d", size, expected)
		}
	})

	t.Run("StructValue", func(t *testing.T) {
		sv := &StructValue{Fields: make([]TypedValue, 3)}
		expected := int64(allocStruct + allocStructField*3)
		if size := sv.GetShallowSize(); size != expected {
			t.Errorf("StructValue.GetShallowSize() = %d, expected %d", size, expected)
		}
	})

	t.Run("SliceValue", func(t *testing.T) {
		sv := &SliceValue{}
		expected := int64(allocSlice)
		if size := sv.GetShallowSize(); size != expected {
			t.Errorf("SliceValue.GetShallowSize() = %d, expected %d", size, expected)
		}
	})
}
