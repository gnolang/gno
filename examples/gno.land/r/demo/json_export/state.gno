package json_export

import "gno.land/p/nt/avl"

// === PERSISTED PRIMITIVES ===

var (
	persistedInt     int     = 42
	persistedInt8    int8    = -8
	persistedInt16   int16   = -1600
	persistedInt32   int32   = -320000
	persistedInt64   int64   = -6400000000
	persistedUint    uint    = 42
	persistedUint8   uint8   = 255
	persistedUint16  uint16  = 65535
	persistedUint32  uint32  = 4294967295
	persistedUint64  uint64  = 18446744073709551615
	persistedFloat32 float32 = 3.14
	persistedFloat64 float64 = 2.718281828
	persistedString  string  = "hello world"
	persistedBool    bool    = true
	persistedByte    byte    = 0xFF
	persistedRune    rune    = 'G' // Unicode: 71
)

// === PERSISTED DECLARED TYPES ===

var (
	persistedMyInt    MyInt    = 100
	persistedMyString MyString = "declared string"
	persistedMyBool   MyBool   = true
)

// === PERSISTED STRUCTS ===

var (
	persistedSimple = SimpleStruct{ID: 1, Name: "first"}
	persistedOuter  = Outer{Name: "outer", Inner: Inner{Value: 99}}
	persistedDeep   = Level1{L2: Level2{L3: Level3{Data: "deep value"}}}
	persistedTagged = TaggedStruct{PublicName: "visible", Internal: 123, Skipped: "hidden"}
	persistedEmpty  = EmptyStruct{}
)

// === PERSISTED POINTERS ===

var (
	persistedPtr    = &SimpleStruct{ID: 2, Name: "pointed"}
	persistedNilPtr *SimpleStruct
)

// === PERSISTED LINKED LIST ===

var persistedList *Node

// === PERSISTED COLLECTIONS ===

var (
	persistedIntSlice    = []int{1, 2, 3, 4, 5}
	persistedStringSlice = []string{"alpha", "beta", "gamma"}
	persistedStructSlice = []SimpleStruct{
		{ID: 10, Name: "item1"},
		{ID: 20, Name: "item2"},
		{ID: 30, Name: "item3"},
	}
	persistedPtrSlice = []*SimpleStruct{
		{ID: 100, Name: "ptr_item1"},
		{ID: 200, Name: "ptr_item2"},
	}
	persistedIntArray  = [3]int{10, 20, 30}
	persistedStringMap = map[string]int{"one": 1, "two": 2, "three": 3}
	persistedStructMap = map[string]SimpleStruct{"first": {ID: 1, Name: "map_item1"}}
	persistedNilSlice  []int
	persistedNilMap    map[string]int
)

// === AVL TREE ===

var persistedTree *avl.Tree

// === LONG LINKED LIST ===

var persistedLongList *Node

// === PERSISTED ERROR ===

var persistedError = MyError{Message: "persisted error message"}

// === PERSISTED COMBINED (AllTypes) ===

var persistedAll *AllTypes

func init() {
	// Build linked list: 1 -> 2 -> 3 -> nil
	persistedList = &Node{
		Value: 1,
		Next: &Node{
			Value: 2,
			Next: &Node{
				Value: 3,
			},
		},
	}

	// Build AVL tree with user data for navigation demo
	persistedTree = avl.NewTree()
	persistedTree.Set("alice", &SimpleStruct{ID: 1, Name: "Alice"})
	persistedTree.Set("bob", &SimpleStruct{ID: 2, Name: "Bob"})
	persistedTree.Set("carol", &SimpleStruct{ID: 3, Name: "Carol"})
	persistedTree.Set("dave", &SimpleStruct{ID: 4, Name: "Dave"})
	persistedTree.Set("eve", &SimpleStruct{ID: 5, Name: "Eve"})
	persistedTree.Set("frank", &SimpleStruct{ID: 6, Name: "Frank"})
	persistedTree.Set("grace", &SimpleStruct{ID: 7, Name: "Grace"})
	persistedTree.Set("heidi", &SimpleStruct{ID: 8, Name: "Heidi"})
	persistedTree.Set("ivan", &SimpleStruct{ID: 9, Name: "Ivan"})
	persistedTree.Set("judy", &SimpleStruct{ID: 10, Name: "Judy"})

	// Build long linked list: 1 -> 2 -> ... -> 10 -> nil
	persistedLongList = buildLongList(10)

	// Build combined AllTypes struct
	persistedAll = &AllTypes{
		// Primitives
		Int:     persistedInt,
		Int8:    persistedInt8,
		Int16:   persistedInt16,
		Int32:   persistedInt32,
		Int64:   persistedInt64,
		Uint:    persistedUint,
		Uint8:   persistedUint8,
		Uint16:  persistedUint16,
		Uint32:  persistedUint32,
		Uint64:  persistedUint64,
		Float32: persistedFloat32,
		Float64: persistedFloat64,
		String:  persistedString,
		Bool:    persistedBool,
		Byte:    persistedByte,
		Rune:    persistedRune,

		// Declared types
		DeclaredInt:    persistedMyInt,
		DeclaredString: persistedMyString,
		DeclaredBool:   persistedMyBool,

		// Structs
		Simple: persistedSimple,
		Nested: persistedOuter,
		Deep:   persistedDeep,
		Tagged: persistedTagged,
		Empty:  persistedEmpty,

		// Pointers
		StructPtr: persistedPtr,
		NilPtr:    nil,
		List:      persistedList,

		// Slices
		IntSlice:    persistedIntSlice,
		StringSlice: persistedStringSlice,
		StructSlice: persistedStructSlice,
		PtrSlice:    persistedPtrSlice,

		// Arrays
		IntArray: persistedIntArray,

		// Maps
		StringIntMap:    persistedStringMap,
		StringStructMap: persistedStructMap,

		// Error
		Error: persistedError,

		// Nil values
		NilSlice:     nil,
		NilMap:       nil,
		NilInterface: nil,
	}
}

// buildLongList creates a linked list of n nodes
func buildLongList(n int) *Node {
	if n <= 0 {
		return nil
	}
	head := &Node{Value: 1}
	current := head
	for i := 2; i <= n; i++ {
		current.Next = &Node{Value: i}
		current = current.Next
	}
	return head
}
