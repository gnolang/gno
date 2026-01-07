package json_export

import (
	"math"
	"strconv"

	"gno.land/p/nt/avl"
)

// PRIMITIVES

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

// DECLARED TYPES

func GetMyInt() MyInt       { return persistedMyInt }
func GetMyString() MyString { return persistedMyString }
func GetMyBool() MyBool     { return persistedMyBool }

// STRUCTS

func GetSimpleStruct() SimpleStruct { return persistedSimple }
func GetNestedStruct() Outer        { return persistedOuter }
func GetDeepStruct() Level1         { return persistedDeep }
func GetTaggedStruct() TaggedStruct { return persistedTagged }
func GetEmptyStruct() EmptyStruct   { return persistedEmpty }

// GetEphemeralStruct returns a non-persisted (unreal) struct
func GetEphemeralStruct() SimpleStruct {
	return SimpleStruct{ID: 999, Name: "ephemeral"}
}

// POINTERS

func GetStructPtr() *SimpleStruct { return persistedPtr }
func GetNilPtr() *SimpleStruct    { return persistedNilPtr }
func GetLinkedList() *Node        { return persistedList }

// GetEphemeralPtr returns a non-persisted pointer (no ObjectID)
func GetEphemeralPtr() *SimpleStruct {
	return &SimpleStruct{ID: 888, Name: "ephemeral_ptr"}
}

// COLLECTIONS

func GetIntSlice() []int                    { return persistedIntSlice }
func GetStringSlice() []string              { return persistedStringSlice }
func GetStructSlice() []SimpleStruct        { return persistedStructSlice }
func GetPtrSlice() []*SimpleStruct          { return persistedPtrSlice }
func GetIntArray() [3]int                   { return persistedIntArray }
func GetStringIntMap() map[string]int       { return persistedStringMap }
func GetStructMap() map[string]SimpleStruct { return persistedStructMap }

// ERRORS

func GetError() error        { return persistedError }
func GetErrorValue() MyError { return persistedError }

// MULTIPLE RETURN VALUES

func GetMultiple() (int, string, SimpleStruct) {
	return 42, "test", persistedSimple
}

func GetMultipleWithPtr() (int, *SimpleStruct, error) {
	return 1, persistedPtr, persistedError
}

// INTERFACE

func GetInterface() interface{} {
	return persistedSimple
}

func GetInterfacePtr() interface{} {
	return persistedPtr
}

// NIL VALUES

func GetNilSlice() []int           { return persistedNilSlice }
func GetNilMap() map[string]int    { return persistedNilMap }
func GetNilInterface() interface{} { return nil }

// RECURSIVE/CYCLIC EPHEMERAL STRUCTURES
// These are non-persisted (unreal) objects with cycles - no ObjectIDs

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

// COMBINED - ALL TYPES IN ONE STRUCT

// GetAll returns a single struct containing all type variations
// This is the comprehensive test function for JSON export
func GetAll() *AllTypes {
	return persistedAll
}

// AVL TREE DEMO

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

// LONG LINKED LIST DEMO

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

// DYNAMIC PARAMS DEMO

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

// EDGE CASES: VERTICAL (DEEP NESTING)

// GetDeepNestedStruct returns a linked list N levels deep
// Structure: {Value:0, Next: {Value:1, Next: {...Value:N-1, Next:nil}}}
func GetDeepNestedStruct(n int) *Node {
	if n <= 0 {
		return nil
	}
	root := &Node{Value: 0}
	current := root
	for i := 1; i < n; i++ {
		current.Next = &Node{Value: i}
		current = current.Next
	}
	return root
}

// GetDeepNestedSlice returns N levels of nested slices
// Structure: [[[[...["leaf"]...]]]]
func GetDeepNestedSlice(n int) []interface{} {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []interface{}{"leaf"}
	}
	return []interface{}{GetDeepNestedSlice(n - 1)}
}

// GetDeepNestedMap returns N levels of nested maps
// Structure: {"level0": {"level1": {...{"leaf": "value"}}}}
func GetDeepNestedMap(n int) map[string]interface{} {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return map[string]interface{}{"leaf": "value"}
	}
	key := "level" + strconv.Itoa(n-1)
	return map[string]interface{}{key: GetDeepNestedMap(n - 1)}
}

// EDGE CASES: HORIZONTAL (WIDE STRUCTURES)

// GetWideStruct returns a map with N fields (simulating wide struct)
// Structure: {"field0": 0, "field1": 1, ..., "fieldN-1": N-1}
func GetWideStruct(n int) map[string]interface{} {
	if n <= 0 {
		return make(map[string]interface{})
	}
	fields := make(map[string]interface{})
	for i := 0; i < n; i++ {
		fields["field"+strconv.Itoa(i)] = i
	}
	return fields
}

// GetWideSlice returns a slice with N elements
// Structure: [0, 1, 2, ..., N-1]
func GetWideSlice(n int) []int {
	if n <= 0 {
		return nil
	}
	slice := make([]int, n)
	for i := 0; i < n; i++ {
		slice[i] = i
	}
	return slice
}

// GetWideMap returns a map with N entries
// Structure: {"key0": 0, "key1": 1, ..., "keyN-1": N-1}
func GetWideMap(n int) map[string]int {
	if n <= 0 {
		return nil
	}
	m := make(map[string]int)
	for i := 0; i < n; i++ {
		m["key"+strconv.Itoa(i)] = i
	}
	return m
}

// GetWideStructSlice returns N SimpleStruct elements
func GetWideStructSlice(n int) []SimpleStruct {
	if n <= 0 {
		return nil
	}
	slice := make([]SimpleStruct, n)
	for i := 0; i < n; i++ {
		slice[i] = SimpleStruct{ID: i, Name: "item" + strconv.Itoa(i)}
	}
	return slice
}

// GetWidePointerArray returns an array of N pointers to structs
func GetWidePointerArray(n int) *PointerArray {
	if n <= 0 {
		return &PointerArray{Pointers: nil}
	}
	pointers := make([]*SimpleStruct, n)
	for i := 0; i < n; i++ {
		pointers[i] = &SimpleStruct{ID: i, Name: "ptr" + strconv.Itoa(i)}
	}
	return &PointerArray{Pointers: pointers}
}

// EDGE CASES: COMBINED STRESS

// GetWideAndDeep returns N fields per level, M levels deep
// Tests memory and recursion with both horizontal and vertical stress
func GetWideAndDeep(width, depth int) *GridNode {
	if depth <= 0 {
		return nil
	}
	fields := make(map[string]interface{})
	for i := 0; i < width; i++ {
		fields["field"+strconv.Itoa(i)] = i
	}
	return &GridNode{
		Level:  depth,
		Fields: fields,
		Child:  GetWideAndDeep(width, depth-1),
	}
}

// GetBinaryTree returns a complete binary tree with N levels
// Total nodes = 2^N - 1
func GetBinaryTree(levels int) *TreeNode {
	if levels <= 0 {
		return nil
	}
	return buildBinaryTree(0, levels)
}

func buildBinaryTree(value, remainingLevels int) *TreeNode {
	if remainingLevels <= 0 {
		return nil
	}
	return &TreeNode{
		Value: value,
		Left:  buildBinaryTree(value*2+1, remainingLevels-1),
		Right: buildBinaryTree(value*2+2, remainingLevels-1),
	}
}

// EDGE CASES: STRINGS

// GetLongString returns a string with N characters
func GetLongString(n int) string {
	if n <= 0 {
		return ""
	}
	// Repeat pattern for readable output
	pattern := "abcdefghij"
	result := ""
	for len(result) < n {
		result += pattern
	}
	return result[:n]
}

// GetUnicodeString returns a string with various unicode characters
func GetUnicodeString() string {
	return "Hello \u4e16\u754c \u3053\u3093\u306b\u3061\u306f \u0421\u0430\u043b\u044e\u0442"
}

// GetEmojiString returns a string with emojis
func GetEmojiString() string {
	return "\U0001F600 \U0001F680 \U0001F4BB \U0001F3C6 \U0001F389"
}

// GetNullByteString returns a string containing null bytes
func GetNullByteString() string {
	return "before\x00middle\x00after"
}

// GetSpecialCharsString returns a string with JSON-sensitive characters
func GetSpecialCharsString() string {
	return "quotes: \"double\" 'single' | backslashes: \\ \\\\ | newlines: \n\r\n | tabs: \t"
}

// GetStringEdgeCases returns all string edge cases in one struct
func GetStringEdgeCases(longStringLen int) *StringEdgeCases {
	return &StringEdgeCases{
		LongString:     GetLongString(longStringLen),
		UnicodeString:  GetUnicodeString(),
		EmojiString:    GetEmojiString(),
		NullByteString: GetNullByteString(),
		SpecialChars:   GetSpecialCharsString(),
		Newlines:       "line1\nline2\r\nline3",
		Quotes:         `"quoted" and 'quoted'`,
		Backslashes:    `path\to\file and \\server\share`,
		MixedContent:   "mix: \"\\\n\t\u4e16\U0001F600",
	}
}

// EDGE CASES: NUMERICS

// GetNumericEdgeCases returns all numeric boundary values
func GetNumericEdgeCases() *NumericEdgeCases {
	return &NumericEdgeCases{
		MaxInt64:    math.MaxInt64,
		MinInt64:    math.MinInt64,
		MaxUint64:   math.MaxUint64,
		MaxInt32:    math.MaxInt32,
		MinInt32:    math.MinInt32,
		MaxInt16:    math.MaxInt16,
		MinInt16:    math.MinInt16,
		MaxInt8:     math.MaxInt8,
		MinInt8:     math.MinInt8,
		MaxFloat64:  math.MaxFloat64,
		MinFloat64:  -math.MaxFloat64,
		TinyFloat64: math.SmallestNonzeroFloat64,
		MaxFloat32:  math.MaxFloat32,
		MinFloat32:  -math.MaxFloat32,
	}
}

// GetMaxIntValues returns maximum integer values for each type
func GetMaxIntValues() (int, int8, int16, int32, int64) {
	return math.MaxInt, math.MaxInt8, math.MaxInt16, math.MaxInt32, math.MaxInt64
}

// GetMinIntValues returns minimum integer values for each type
func GetMinIntValues() (int, int8, int16, int32, int64) {
	return math.MinInt, math.MinInt8, math.MinInt16, math.MinInt32, math.MinInt64
}

// GetMaxUintValues returns maximum unsigned integer values
func GetMaxUintValues() (uint, uint8, uint16, uint32, uint64) {
	return math.MaxUint, math.MaxUint8, math.MaxUint16, math.MaxUint32, math.MaxUint64
}

// GetFloatEdgeCases returns float edge cases
func GetFloatEdgeCases() (float32, float32, float64, float64, float64) {
	return math.MaxFloat32, math.SmallestNonzeroFloat32,
		math.MaxFloat64, math.SmallestNonzeroFloat64, -math.MaxFloat64
}

// EDGE CASES: POINTERS

// GetDeepPointerChain returns N levels of pointer indirection
// Each level wraps a Node with a pointer to the next level
func GetDeepPointerChain(n int) *Node {
	if n <= 0 {
		return nil
	}
	root := &Node{Value: 0}
	current := root
	for i := 1; i < n; i++ {
		current.Next = &Node{Value: i}
		current = current.Next
	}
	return root
}

// GetMixedPointerSlice returns slice with mix of nil and non-nil pointers
func GetMixedPointerSlice(n int) []*SimpleStruct {
	if n <= 0 {
		return nil
	}
	slice := make([]*SimpleStruct, n)
	for i := 0; i < n; i++ {
		if i%3 == 0 {
			slice[i] = nil // Every third element is nil
		} else {
			slice[i] = &SimpleStruct{ID: i, Name: "mixed" + strconv.Itoa(i)}
		}
	}
	return slice
}

// RENDER (for web interface)

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

### Edge Cases (Stress Testing)

#### Vertical (Deep Nesting) - Dynamic N parameter
- **GetDeepNestedStruct(n int)** - N-level deep linked list chain
- **GetDeepNestedSlice(n int)** - N-level nested slices [[[...]]]
- **GetDeepNestedMap(n int)** - N-level nested maps

#### Horizontal (Wide Structures) - Dynamic N parameter
- **GetWideStruct(n int)** - Map with N fields
- **GetWideSlice(n int)** - Slice with N elements
- **GetWideMap(n int)** - Map with N entries
- **GetWideStructSlice(n int)** - N SimpleStruct elements
- **GetWidePointerArray(n int)** - N pointers to structs

#### Combined Stress - Dynamic width/depth parameters
- **GetWideAndDeep(width, depth int)** - N fields x M levels
- **GetBinaryTree(levels int)** - Complete binary tree (2^N - 1 nodes)

#### String Edge Cases
- **GetLongString(n int)** - String with N characters
- **GetUnicodeString()** - Unicode chars (Chinese, Japanese, Russian)
- **GetEmojiString()** - Emoji characters
- **GetNullByteString()** - String with null bytes
- **GetSpecialCharsString()** - JSON-sensitive: quotes, backslashes, newlines, tabs
- **GetStringEdgeCases(longLen int)** - All string edge cases combined

#### Numeric Edge Cases
- **GetNumericEdgeCases()** - All boundary values (MaxInt64, MinInt64, etc.)
- **GetMaxIntValues()** - Max int values (int, int8-64)
- **GetMinIntValues()** - Min int values
- **GetMaxUintValues()** - Max uint values
- **GetFloatEdgeCases()** - Float boundary values

#### Pointer Edge Cases
- **GetDeepPointerChain(n int)** - N levels of pointer indirection
- **GetMixedPointerSlice(n int)** - Mix of nil/non-nil pointers

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

Edge case stress tests:
` + "```" + `
/query/r/demo/json_export.GetDeepNestedStruct(100)
/query/r/demo/json_export.GetWideSlice(10000)
/query/r/demo/json_export.GetBinaryTree(10)
/query/r/demo/json_export.GetNumericEdgeCases()
/query/r/demo/json_export.GetStringEdgeCases(1000)
/query/r/demo/json_export.GetWideAndDeep(50, 20)
` + "```" + `
`
}
