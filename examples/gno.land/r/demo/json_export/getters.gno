package json_export

import "gno.land/p/nt/avl"

// ============================================================================
// PRIMITIVES
// ============================================================================

func GetInt() int         { return persistedInt }
func GetInt8() int8       { return persistedInt8 }
func GetInt16() int16     { return persistedInt16 }
func GetInt32() int32     { return persistedInt32 }
func GetInt64() int64     { return persistedInt64 }
func GetUint() uint       { return persistedUint }
func GetUint8() uint8     { return persistedUint8 }
func GetUint16() uint16   { return persistedUint16 }
func GetUint32() uint32   { return persistedUint32 }
func GetUint64() uint64   { return persistedUint64 }
func GetFloat32() float32 { return persistedFloat32 }
func GetFloat64() float64 { return persistedFloat64 }
func GetString() string   { return persistedString }
func GetBool() bool       { return persistedBool }
func GetByte() byte       { return persistedByte }
func GetRune() rune       { return persistedRune }

// GetAllPrimitives returns all primitive types at once
func GetAllPrimitives() (int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool, byte, rune) {
	return persistedInt, persistedInt8, persistedInt16, persistedInt32, persistedInt64,
		persistedUint, persistedUint8, persistedUint16, persistedUint32, persistedUint64,
		persistedFloat32, persistedFloat64, persistedString, persistedBool, persistedByte, persistedRune
}

// ============================================================================
// DECLARED TYPES
// ============================================================================

func GetMyInt() MyInt       { return persistedMyInt }
func GetMyString() MyString { return persistedMyString }
func GetMyBool() MyBool     { return persistedMyBool }

// ============================================================================
// STRUCTS
// ============================================================================

func GetSimpleStruct() SimpleStruct { return persistedSimple }
func GetNestedStruct() Outer        { return persistedOuter }
func GetDeepStruct() Level1         { return persistedDeep }
func GetTaggedStruct() TaggedStruct { return persistedTagged }
func GetEmptyStruct() EmptyStruct   { return persistedEmpty }

// GetEphemeralStruct returns a non-persisted (unreal) struct
func GetEphemeralStruct() SimpleStruct {
	return SimpleStruct{ID: 999, Name: "ephemeral"}
}

// ============================================================================
// POINTERS
// ============================================================================

func GetStructPtr() *SimpleStruct { return persistedPtr }
func GetNilPtr() *SimpleStruct    { return persistedNilPtr }
func GetLinkedList() *Node        { return persistedList }

// GetEphemeralPtr returns a non-persisted pointer (no ObjectID)
func GetEphemeralPtr() *SimpleStruct {
	return &SimpleStruct{ID: 888, Name: "ephemeral_ptr"}
}

// ============================================================================
// COLLECTIONS
// ============================================================================

func GetIntSlice() []int                    { return persistedIntSlice }
func GetStringSlice() []string              { return persistedStringSlice }
func GetStructSlice() []SimpleStruct        { return persistedStructSlice }
func GetPtrSlice() []*SimpleStruct          { return persistedPtrSlice }
func GetIntArray() [3]int                   { return persistedIntArray }
func GetStringIntMap() map[string]int       { return persistedStringMap }
func GetStructMap() map[string]SimpleStruct { return persistedStructMap }

// ============================================================================
// ERRORS
// ============================================================================

func GetError() error        { return persistedError }
func GetErrorValue() MyError { return persistedError }

// ============================================================================
// MULTIPLE RETURN VALUES
// ============================================================================

func GetMultiple() (int, string, SimpleStruct) {
	return 42, "test", persistedSimple
}

func GetMultipleWithPtr() (int, *SimpleStruct, error) {
	return 1, persistedPtr, persistedError
}

// ============================================================================
// INTERFACE
// ============================================================================

func GetInterface() interface{} {
	return persistedSimple
}

func GetInterfacePtr() interface{} {
	return persistedPtr
}

// ============================================================================
// NIL VALUES
// ============================================================================

func GetNilSlice() []int           { return persistedNilSlice }
func GetNilMap() map[string]int    { return persistedNilMap }
func GetNilInterface() interface{} { return nil }

// ============================================================================
// RECURSIVE/CYCLIC EPHEMERAL STRUCTURES
// These are non-persisted (unreal) objects with cycles - no ObjectIDs
// ============================================================================

// GetEphemeralSelfCycle returns a self-referential struct with a cycle
// Structure: node -> node (cycle back to itself)
func GetEphemeralSelfCycle() *SelfRef {
	node := &SelfRef{Name: "cyclic"}
	node.Self = node // Create cycle: node points to itself
	return node
}

// GetEphemeralMutualCycle returns mutually recursive structs with a cycle
// Structure: a -> b -> a (cycle)
func GetEphemeralMutualCycle() *MutualA {
	a := &MutualA{Value: 1}
	b := &MutualB{Name: "mutual"}
	a.B = b
	b.A = a // Create cycle: a -> b -> a
	return a
}

// GetEphemeralLinkedListCycle returns a linked list with a cycle
// Structure: 1 -> 2 -> 3 -> 1 (cycle back to head)
func GetEphemeralLinkedListCycle() *Node {
	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2}
	n3 := &Node{Value: 3}
	n1.Next = n2
	n2.Next = n3
	n3.Next = n1 // Create cycle: 3 points back to 1
	return n1
}

// GetEphemeralDeepChain returns a deep non-cyclic chain (no cycle)
// Structure: 1 -> 2 -> 3 -> 4 -> 5 -> nil
func GetEphemeralDeepChain() *Node {
	return &Node{
		Value: 1,
		Next: &Node{
			Value: 2,
			Next: &Node{
				Value: 3,
				Next: &Node{
					Value: 4,
					Next: &Node{
						Value: 5,
						Next:  nil,
					},
				},
			},
		},
	}
}

// GetEphemeralNestedSelfRef returns nested self-references without cycle
// Structure: outer -> inner -> nil
func GetEphemeralNestedSelfRef() *SelfRef {
	return &SelfRef{
		Name: "outer",
		Self: &SelfRef{
			Name: "inner",
			Self: nil,
		},
	}
}

// ============================================================================
// COMBINED - ALL TYPES IN ONE STRUCT
// ============================================================================

// GetAll returns a single struct containing all type variations
// This is the comprehensive test function for JSON export
func GetAll() *AllTypes {
	return persistedAll
}

// ============================================================================
// AVL TREE DEMO
// ============================================================================

// GetTree returns an AVL tree with 10 users for navigation demo
func GetTree() *avl.Tree { return persistedTree }

// GetTreeNode returns a specific user by key (e.g., "alice", "bob")
func GetTreeNode(key string) interface{} {
	val, exists := persistedTree.Get(key)
	if !exists {
		return nil
	}
	return val
}

// ============================================================================
// LONG LINKED LIST DEMO
// ============================================================================

// GetLongList returns a 10-node linked list for navigation demo
func GetLongList() *Node { return persistedLongList }

// GetListNode returns the nth node in the list (0-indexed)
func GetListNode(n int) *Node {
	node := persistedLongList
	for i := 0; i < n && node != nil; i++ {
		node = node.Next
	}
	return node
}

// ============================================================================
// DYNAMIC PARAMS DEMO
// ============================================================================

// GetByID returns a struct based on the ID parameter
func GetByID(id int) *SimpleStruct {
	switch id {
	case 1:
		return &persistedSimple
	case 2:
		return persistedPtr
	default:
		return &SimpleStruct{ID: id, Name: "dynamic"}
	}
}

// GetRange returns a slice of integers from start to end (exclusive)
func GetRange(start, end int) []int {
	if start < 0 || end <= start || end > len(persistedIntSlice) {
		return nil
	}
	return persistedIntSlice[start:end]
}

// Lookup returns a value from the string-int map by key
func Lookup(key string) int {
	return persistedStringMap[key]
}

// GetAllEphemeral returns a non-persisted AllTypes struct (no ObjectIDs)
func GetAllEphemeral() AllTypes {
	return AllTypes{
		// Primitives
		Int:     1,
		Int8:    2,
		Int16:   3,
		Int32:   4,
		Int64:   5,
		Uint:    6,
		Uint8:   7,
		Uint16:  8,
		Uint32:  9,
		Uint64:  10,
		Float32: 1.1,
		Float64: 2.2,
		String:  "ephemeral",
		Bool:    false,
		Byte:    0xAB,
		Rune:    'X',

		// Declared types
		DeclaredInt:    MyInt(11),
		DeclaredString: MyString("ephemeral declared"),
		DeclaredBool:   MyBool(false),

		// Structs
		Simple: SimpleStruct{ID: 111, Name: "ephemeral_simple"},
		Nested: Outer{Name: "ephemeral_outer", Inner: Inner{Value: 222}},
		Deep:   Level1{L2: Level2{L3: Level3{Data: "ephemeral_deep"}}},
		Tagged: TaggedStruct{PublicName: "ephemeral_visible", Internal: 333, Skipped: "ephemeral_hidden"},
		Empty:  EmptyStruct{},

		// Pointers - ephemeral, no ObjectID
		StructPtr: &SimpleStruct{ID: 444, Name: "ephemeral_ptr"},
		NilPtr:    nil,
		List:      &Node{Value: 100, Next: &Node{Value: 200}},

		// Slices
		IntSlice:    []int{100, 200, 300},
		StringSlice: []string{"x", "y", "z"},
		StructSlice: []SimpleStruct{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
		PtrSlice:    []*SimpleStruct{{ID: 10, Name: "pa"}, {ID: 20, Name: "pb"}},

		// Arrays
		IntArray: [3]int{100, 200, 300},

		// Maps
		StringIntMap:    map[string]int{"a": 1, "b": 2},
		StringStructMap: map[string]SimpleStruct{"x": {ID: 1, Name: "mx"}},

		// Error
		Error: MyError{Message: "ephemeral error"},

		// Nil values
		NilSlice:     nil,
		NilMap:       nil,
		NilInterface: nil,
	}
}

// ============================================================================
// RENDER (for web interface)
// ============================================================================

func Render(_ string) string {
	return `# JSON Export Test Realm

This realm provides various type examples for testing JSON export functionality.

## Available Getters

### Primitives
- GetInt(), GetInt8(), GetInt16(), GetInt32(), GetInt64()
- GetUint(), GetUint8(), GetUint16(), GetUint32(), GetUint64()
- GetFloat32(), GetFloat64()
- GetString(), GetBool(), GetByte(), GetRune()
- GetAllPrimitives()

### Declared Types
- GetMyInt(), GetMyString(), GetMyBool()

### Structs
- GetSimpleStruct(), GetNestedStruct(), GetDeepStruct()
- GetTaggedStruct(), GetEmptyStruct()
- GetEphemeralStruct()

### Pointers
- GetStructPtr(), GetNilPtr(), GetLinkedList()
- GetEphemeralPtr()

### Collections
- GetIntSlice(), GetStringSlice(), GetStructSlice(), GetPtrSlice()
- GetIntArray()
- GetStringIntMap(), GetStructMap()

### Errors
- GetError(), GetErrorValue()

### Multiple Returns
- GetMultiple(), GetMultipleWithPtr()

### Interface
- GetInterface(), GetInterfacePtr(), GetNilInterface()

### Recursive/Cyclic (Ephemeral)
- GetEphemeralSelfCycle() - Self-referential cycle (node -> node)
- GetEphemeralMutualCycle() - Mutual recursion cycle (a -> b -> a)
- GetEphemeralLinkedListCycle() - Linked list cycle (1 -> 2 -> 3 -> 1)
- GetEphemeralDeepChain() - Deep chain without cycle
- GetEphemeralNestedSelfRef() - Nested self-refs without cycle

### Combined
- **GetAll()** - Returns all types in a single persisted struct
- **GetAllEphemeral()** - Returns all types in a non-persisted struct

### AVL Tree (Navigation Demo)
- **GetTree()** - Returns an AVL tree with 10 users (jump between nodes via ObjectIDs)
- **GetTreeNode(key string)** - Get user by key ("alice", "bob", etc.)

### Long Linked List (Navigation Demo)
- **GetLongList()** - Returns a 10-node linked list (navigate node by node)
- **GetListNode(n int)** - Get nth node in list (0-indexed)

### Dynamic Parameters
- **GetByID(id int)** - Get struct by ID (1=first, 2=pointed, other=dynamic)
- **GetRange(start, end int)** - Get slice of integers [start, end)
- **Lookup(key string)** - Get value from map ("one", "two", "three")

## Usage

Query with JSON format:
` + "```" + `
gnokey query vm/qeval -data "gno.land/r/demo/json_export.GetAll()" --format json
` + "```" + `

Web query examples:
` + "```" + `
/query/r/demo/json_export.GetTree()
/query/r/demo/json_export.GetTreeNode("alice")
/query/r/demo/json_export.GetLongList()
/query/r/demo/json_export.GetListNode(5)
/query/r/demo/json_export.GetByID(1)
/query/r/demo/json_export.Lookup("one")
` + "```" + `
`
}
