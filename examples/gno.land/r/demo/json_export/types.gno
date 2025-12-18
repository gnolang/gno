package json_export

// === DECLARED TYPES ===

type MyInt int
type MyString string
type MyBool bool

// === SIMPLE STRUCT ===

type SimpleStruct struct {
	ID   int
	Name string
}

// === NESTED STRUCTS ===

type Inner struct {
	Value int
}

type Outer struct {
	Name  string
	Inner Inner
}

// Deeply nested (3 levels)
type Level3 struct {
	Data string
}

type Level2 struct {
	L3 Level3
}

type Level1 struct {
	L2 Level2
}

// === STRUCT WITH JSON TAGS ===

type TaggedStruct struct {
	PublicName string `json:"public_name"`
	Internal   int    `json:"internal_id"`
	Skipped    string `json:"-"`
}

// === STRUCT WITH POINTER (Linked List) ===

type Node struct {
	Value int
	Next  *Node
}

// === RECURSIVE/CYCLIC STRUCTS ===

// SelfRef is a struct that can reference itself (for cycle testing)
type SelfRef struct {
	Name string
	Self *SelfRef
}

// MutualA and MutualB form a mutual recursion pair
type MutualA struct {
	Value int
	B     *MutualB
}

type MutualB struct {
	Name string
	A    *MutualA
}

// === ERROR TYPE ===

type MyError struct {
	Message string
}

func (e MyError) Error() string {
	return e.Message
}

// === EMPTY STRUCT ===

type EmptyStruct struct{}

// ============================================================================
// EDGE CASE TYPES (for stress testing)
// ============================================================================

// TreeNode for binary tree tests
type TreeNode struct {
	Value int
	Left  *TreeNode
	Right *TreeNode
}

// GridNode for wide+deep tests (N fields x M levels)
type GridNode struct {
	Level  int
	Fields map[string]interface{}
	Child  *GridNode
}

// StringEdgeCases holds problematic string values
type StringEdgeCases struct {
	LongString     string
	UnicodeString  string
	EmojiString    string
	NullByteString string
	SpecialChars   string
	Newlines       string
	Quotes         string
	Backslashes    string
	MixedContent   string
}

// NumericEdgeCases holds all numeric boundary values
type NumericEdgeCases struct {
	MaxInt64    int64
	MinInt64    int64
	MaxUint64   uint64
	MaxInt32    int32
	MinInt32    int32
	MaxInt16    int16
	MinInt16    int16
	MaxInt8     int8
	MinInt8     int8
	MaxFloat64  float64
	MinFloat64  float64
	TinyFloat64 float64
	MaxFloat32  float32
	MinFloat32  float32
}

// PointerArray for wide pointer arrays
type PointerArray struct {
	Pointers []*SimpleStruct
}

// === COMBINED STRUCT (for GetAll) ===

// AllTypes combines all type variations into a single struct for comprehensive testing
type AllTypes struct {
	// Primitives
	Int     int     `json:"int"`
	Int8    int8    `json:"int8"`
	Int16   int16   `json:"int16"`
	Int32   int32   `json:"int32"`
	Int64   int64   `json:"int64"`
	Uint    uint    `json:"uint"`
	Uint8   uint8   `json:"uint8"`
	Uint16  uint16  `json:"uint16"`
	Uint32  uint32  `json:"uint32"`
	Uint64  uint64  `json:"uint64"`
	Float32 float32 `json:"float32"`
	Float64 float64 `json:"float64"`
	String  string  `json:"string"`
	Bool    bool    `json:"bool"`
	Byte    byte    `json:"byte"`
	Rune    rune    `json:"rune"`

	// Declared types
	DeclaredInt    MyInt    `json:"declared_int"`
	DeclaredString MyString `json:"declared_string"`
	DeclaredBool   MyBool   `json:"declared_bool"`

	// Structs
	Simple SimpleStruct `json:"simple_struct"`
	Nested Outer        `json:"nested_struct"`
	Deep   Level1       `json:"deep_struct"`
	Tagged TaggedStruct `json:"tagged_struct"`
	Empty  EmptyStruct  `json:"empty_struct"`

	// Pointers
	StructPtr *SimpleStruct `json:"struct_ptr"`
	NilPtr    *SimpleStruct `json:"nil_ptr"`
	List      *Node         `json:"linked_list"`

	// Slices
	IntSlice    []int           `json:"int_slice"`
	StringSlice []string        `json:"string_slice"`
	StructSlice []SimpleStruct  `json:"struct_slice"`
	PtrSlice    []*SimpleStruct `json:"ptr_slice"`

	// Arrays
	IntArray [3]int `json:"int_array"`

	// Maps
	StringIntMap    map[string]int          `json:"string_int_map"`
	StringStructMap map[string]SimpleStruct `json:"string_struct_map"`

	// Error
	Error MyError `json:"error"`

	// Nil values
	NilSlice     []int          `json:"nil_slice"`
	NilMap       map[string]int `json:"nil_map"`
	NilInterface interface{}    `json:"nil_interface"`
}
