package tests

// Condensed test types inspired by gnovm amino-registered types.
// These capture the key field patterns found across ~100 gnovm types:
//   - Interface fields (Type, Value, Expr, Stmt, etc.)
//   - Slices of interfaces ([]Expr, []Stmt, []TypedValue)
//   - Deeply nested embedded structs (Attributes → Location → Span → Pos)
//   - Byte arrays of various sizes ([8]byte, [20]byte)
//   - Pointer-to-struct fields
//   - Struct containing mix of embedded + named struct fields + primitives
//   - Custom amino marshaler returning struct
//   - Slices of structs that themselves contain interfaces

// ----------------------------------------
// GnoVMNode: models gnovm's AST nodes (BinaryExpr, CallExpr, etc.)
// Pattern: embedded Attributes, interface fields, slice of interfaces

type GnoVMNode struct {
	GnoVMAttrs
	Op    int32
	Left  Interface1
	Right Interface1
	Args  []Interface1
}

// ----------------------------------------
// GnoVMAttrs: models gnovm's Attributes + Location + Span + Pos chain
// Pattern: deeply nested embedded structs (4 levels)

type GnoVMAttrs struct {
	GnoVMLocation
	Label string
	Line  int
}

type GnoVMLocation struct {
	PkgPath string
	File    string
	GnoVMSpan
}

type GnoVMSpan struct {
	GnoVMPos
	End GnoVMPos
	Num int
}

type GnoVMPos struct {
	Line   int
	Column int
}

// ----------------------------------------
// GnoVMTypedValue: models gnovm's TypedValue
// Pattern: interface fields + fixed byte array ([8]byte like N field)

type GnoVMTypedValue struct {
	T Interface1 // Type interface
	V Interface1 // Value interface
	N [8]byte    // like TypedValue.N
}

// ----------------------------------------
// GnoVMBlock: models gnovm's Block/ArrayValue/StructValue
// Pattern: embedded ObjectInfo-like struct, slice of TypedValue-like structs

type GnoVMBlock struct {
	GnoVMObjectInfo
	Source Interface1
	Values []GnoVMTypedValue
}

type GnoVMObjectInfo struct {
	ID      GnoVMObjectID
	Hash    [20]byte // like Hashlet
	OwnerID GnoVMObjectID
	ModTime uint64
}

// GnoVMObjectID: models gnovm's ObjectID with amino marshaler
type GnoVMObjectID struct {
	PkgID   [20]byte
	NewTime uint64
}

// ----------------------------------------
// GnoVMFuncValue: models gnovm's FuncValue
// Pattern: many fields of mixed types, pointer to struct, bool flags

type GnoVMFuncValue struct {
	GnoVMObjectInfo
	Type      Interface1
	IsMethod  bool
	IsClosure bool
	Name      string
	Parent    Interface1
	Captures  []GnoVMTypedValue
	PkgPath   string
}

// ----------------------------------------
// GnoVMDeclaredType: models gnovm's DeclaredType
// Pattern: struct with Location, interface, slice of TypedValue-like

type GnoVMDeclaredType struct {
	PkgPath   string
	Name      string
	ParentLoc GnoVMLocation
	Base      Interface1
	Methods   []GnoVMTypedValue
}

// ----------------------------------------
// GnoVMRefValue: models gnovm's RefValue
// Pattern: struct fields, all non-zero checks

type GnoVMRefValue struct {
	ObjectID GnoVMObjectID
	Escaped  bool
	PkgPath  string
	Hash     [20]byte
}

// ----------------------------------------
// GnoVMFieldType: models gnovm's FieldType
// Pattern: interface field + string + bool

type GnoVMFieldType struct {
	Name     string
	Type     Interface1
	Embedded bool
	Tag      string
}

// ----------------------------------------
// GnoVMStructType: models gnovm's StructType
// Pattern: slice of structs that contain interfaces

type GnoVMStructType struct {
	PkgPath string
	Fields  []GnoVMFieldType
}

// ----------------------------------------
// GnoVMFileNode: models deep nesting of gnovm's file/package structure
// Pattern: embedded attrs + static block, string fields, slice of interfaces

type GnoVMFileNode struct {
	GnoVMAttrs
	FileName string
	PkgName  string
	Decls    []Interface1
}

// ----------------------------------------
// GnoVMPointerValue: models gnovm's PointerValue
// Pattern: pointer to struct + interface + int

type GnoVMPointerValue struct {
	TV    *GnoVMTypedValue
	Base  Interface1
	Index int
}

// ----------------------------------------
// GnoVMSliceValue: models gnovm's SliceValue
// Pattern: interface field + multiple int fields

type GnoVMSliceValue struct {
	Base   Interface1
	Offset int
	Length int
	Maxcap int
}

// ----------------------------------------
// GnoVMMapEntry: models gnovm's MapListItem (without circular pointers)
// Pattern: two TypedValue-like struct fields

type GnoVMMapEntry struct {
	Key   GnoVMTypedValue
	Value GnoVMTypedValue
}

// ========================================
// Fuzz-friendly types (no interface fields, so gofuzz can populate them)
// These exercise the patterns that the interface-containing types above cover
// but are safe for automated fuzzing.

// FuzzFieldInfo: like gnovm's FieldType but without interface
// Pattern: slice of these used in FuzzStructInfo
type FuzzFieldInfo struct {
	Name     string
	Embedded bool
	Tag      string
	Index    int
}

// FuzzStructInfo: models gnovm's StructType (slice of nested structs)
// Pattern: slice of struct fields
type FuzzStructInfo struct {
	PkgPath string
	Fields  []FuzzFieldInfo
}

// FuzzValueEntry: like TypedValue but fuzzable (no interfaces)
// Pattern: byte array + nested struct + primitives
type FuzzValueEntry struct {
	N    [8]byte
	Loc  GnoVMLocation
	Kind int32
	Data []byte
}

// FuzzBlock: models Block (embedded struct + slice of nested structs)
// Pattern: embedded ObjectInfo + slice of struct
type FuzzBlock struct {
	GnoVMObjectInfo
	Values []FuzzValueEntry
	Name   string
}

// FuzzFuncInfo: models FuncValue (many mixed fields + slice of nested structs)
// Pattern: embedded struct + bool flags + string fields + struct slice
type FuzzFuncInfo struct {
	GnoVMObjectInfo
	IsMethod  bool
	IsClosure bool
	Name      string
	PkgPath   string
	Captures  []FuzzValueEntry
}

// FuzzDeclInfo: models DeclaredType (nested Location + slice of structs)
// Pattern: nested non-embedded struct + struct slice
type FuzzDeclInfo struct {
	PkgPath   string
	Name      string
	ParentLoc GnoVMLocation
	Methods   []FuzzValueEntry
}

// FuzzFileInfo: models FileNode (deeply embedded + slice of structs)
// Pattern: 4-level embedded struct + slice of nested structs
type FuzzFileInfo struct {
	GnoVMAttrs
	FileName string
	PkgName  string
	Decls    []FuzzFieldInfo
}

// FuzzPtrNest: models PointerValue (pointer to nested struct)
// Pattern: pointer to struct containing byte array + nested struct
type FuzzPtrNest struct {
	Entry *FuzzValueEntry
	Index int
	Name  string
}

// FuzzDeepNest: deeply nested struct slices (3 levels)
// Pattern: struct containing slice of struct containing slice of struct
type FuzzDeepNest struct {
	Blocks []FuzzBlock
	Meta   GnoVMAttrs
}
