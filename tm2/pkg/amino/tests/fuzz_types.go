package tests

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

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

// ========================================
// Fuzz types for amino Go-tag coverage.
// These exercise amino:"write_empty", amino:"nil_elements", and amino:"unsafe".
// Not registered in Package (no pbbindings) — tested via reflect + genproto2 only.

// FuzzWriteEmpty: exercises amino:"write_empty" on various field types.
type FuzzWriteEmpty struct {
	Name   string   `amino:"write_empty"`
	Values []int32  `amino:"write_empty"`
	Inner  GnoVMPos `amino:"write_empty"`
	Data   []byte   `amino:"write_empty"`
	Count  int64    `amino:"write_empty"`
	Flag   bool     `amino:"write_empty"`
	Normal string   // control: no write_empty
}

// FuzzNilElements: exercises amino:"nil_elements" on pointer-to-struct slices.
type FuzzNilElements struct {
	Entries []*FuzzFieldInfo `amino:"nil_elements"`
	Poses   []*GnoVMPos      `amino:"nil_elements"`
	Name    string
}

// FuzzUnsafeFloat: exercises amino:"unsafe" for float types.
type FuzzUnsafeFloat struct {
	Score  float64 `amino:"unsafe"`
	Weight float32 `amino:"unsafe"`
	Label  string
	Count  int32
}

// FuzzFixedInt: exercises binary:"fixed64" on bare int/uint types (not just
// int64/uint64). Catches marshal/unmarshal wire-format divergence.
// Note: binary:"fixed32" on int/uint is rejected by amino's ValidateBasic.
type FuzzFixedInt struct {
	I64 int  `binary:"fixed64"`
	U64 uint `binary:"fixed64"`
}

// FuzzContainsAminoMarshaler: a struct with an AminoMarshaler struct field
// whose repr is itself a struct. Exercises the
// `IsAminoMarshaler && field-type==struct && repr-type==struct` path in
// gen_size.go / gen_marshal.go, which are easy to get out of sync on the
// "should we emit the field key?" decision.
type FuzzContainsAminoMarshaler struct {
	AM AminoMarshalerStruct1
}

// EmptyReprOnZero: a struct AminoMarshaler whose MarshalAmino returns the
// empty string for the zero value and a decimal representation otherwise.
// Mirrors std.Coin's "zero → empty string repr" behavior but is self-
// contained in the tests package. Used by FuzzNilEmptyRepr and as a
// standalone field to exercise the gen_marshal.go zero-check branch for
// struct Go type + non-struct repr where the repr IS zero (the path
// production AminoMarshalers rarely hit).
type EmptyReprOnZero struct {
	Val int32
}

func (e EmptyReprOnZero) MarshalAmino() (string, error) {
	if e.Val == 0 {
		return "", nil
	}
	return fmt.Sprintf("%d", e.Val), nil
}

func (e *EmptyReprOnZero) UnmarshalAmino(s string) error {
	if s == "" {
		e.Val = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	e.Val = int32(v)
	return nil
}

// FuzzNilEmptyRepr: []*EmptyReprOnZero amino:"nil_elements". Under
// nil_elements semantics both nil and a non-nil pointer whose MarshalAmino
// produces "" serialize identically (zero-length element) and both decode
// to nil, so strict DeepEqual roundtrip is intentionally lossy for the
// empty-repr case. The parity invariants that still apply are (1) encoder
// parity, (2) size correctness, and (3) cross-decoder agreement.
type FuzzNilEmptyRepr struct {
	Vals []*EmptyReprOnZero `amino:"nil_elements"`
}

// ========================================
// Types for amino:"skip_zero".

// SkipZeroAddr is a byte array whose repr is never empty, like
// crypto.Address: even its zero value marshals to a non-empty string, so
// amino writes it unless the field is tagged amino:"skip_zero". It is not
// registered, so it encodes inline as its string repr, as crypto.Address
// does in std.Signature.
type SkipZeroAddr [4]byte

func (a SkipZeroAddr) MarshalAmino() (string, error) {
	return "a" + hex.EncodeToString(a[:]), nil
}

func (a *SkipZeroAddr) UnmarshalAmino(s string) error {
	if s == "" { // as crypto.Address reads ""
		*a = SkipZeroAddr{}
		return nil
	}
	if len(s) != 1+2*len(a) || s[0] != 'a' {
		return fmt.Errorf("invalid SkipZeroAddr %q", s)
	}
	_, err := hex.Decode(a[:], []byte(s[1:]))
	return err
}

// SkipZeroInner is a struct whose zero value still encodes to bytes,
// through its untagged SkipZeroAddr.
type SkipZeroInner struct {
	Addr SkipZeroAddr
	N    int32
}

// SkipZeroList cannot be compared with ==, so generated code asks reflect
// whether it is zero.
type SkipZeroList struct {
	Addr SkipZeroAddr
	Ns   []int32
}

// FuzzSkipZero exercises amino:"skip_zero" on the kinds where it changes
// the encoding (an AminoMarshaler with a non-empty zero repr, byte and int
// arrays, an array of structs, structs whose zero value encodes, a float),
// on kinds amino already leaves out at zero (ints, strings, slices), next
// to the other tags, and against untagged controls.
type FuzzSkipZero struct {
	Addr       SkipZeroAddr     `amino:"skip_zero"`
	Bytes      [4]byte          `amino:"skip_zero"`
	Ints       [2]int64         `amino:"skip_zero"`
	Inners     [2]SkipZeroInner `amino:"skip_zero"`
	Inner      SkipZeroInner    `amino:"skip_zero"`
	List       SkipZeroList     `amino:"skip_zero"`
	Score      float64          `amino:"unsafe,skip_zero"`
	Fixed      int64            `binary:"fixed64" amino:"skip_zero"`
	Name       string           `json:"name,omitempty" amino:"skip_zero"`
	Nums       []int32          `amino:"skip_zero"`
	PlainAddr  SkipZeroAddr     // control: written even when zero
	PlainInner SkipZeroInner    // control: written even when zero
}

// SkipZeroInterface: amino:"skip_zero" on an interface field. A nil
// interface is zero and was left out already; a non-nil one holding a zero
// value is not zero, so it is written.
type SkipZeroInterface struct {
	I Interface1 `amino:"skip_zero"`
}

// SkipZeroNilElements: amino:"skip_zero" next to amino:"nil_elements". On
// a slice it changes nothing; on an array it leaves out the all-nil array,
// which nil_elements alone writes as nil elements.
type SkipZeroNilElements struct {
	Slice []*GnoVMPos  `amino:"nil_elements,skip_zero"`
	Array [2]*GnoVMPos `amino:"skip_zero,nil_elements"`
}

// SkipZeroOnly encodes to nothing at zero only because of its tag, so a
// field of this type is empty at zero, and genproto's Is*ReprEmpty must
// say so too.
type SkipZeroOnly struct {
	Addr SkipZeroAddr `amino:"skip_zero"`
}

// SkipZeroHolder holds a SkipZeroOnly, untagged.
type SkipZeroHolder struct {
	Only SkipZeroOnly
	N    int32
}

// SkipZeroElemRepr is SkipZeroElem's repr. Listed as the elements of
// SkipZeroElems's repr, its fields are written by genproto2's inline field
// writer, which does not convert a field to its repr, so none here is an
// AminoMarshaler; Inner holds one.
type SkipZeroElemRepr struct {
	N     int32
	Arr   [2]int32      `amino:"skip_zero"`
	Inner SkipZeroInner `amino:"skip_zero"`
}

// SkipZeroElem is a struct AminoMarshaler with a struct repr.
type SkipZeroElem struct {
	N     int32
	Arr   [2]int32
	Inner SkipZeroInner
}

func (e SkipZeroElem) MarshalAmino() (SkipZeroElemRepr, error) {
	return SkipZeroElemRepr(e), nil
}

func (e *SkipZeroElem) UnmarshalAmino(r SkipZeroElemRepr) error {
	*e = SkipZeroElem(r)
	return nil
}

// SkipZeroElems has a list repr of SkipZeroElem.
type SkipZeroElems struct {
	Items []SkipZeroElem
}

func (l SkipZeroElems) MarshalAmino() ([]SkipZeroElem, error) {
	return l.Items, nil
}

func (l *SkipZeroElems) UnmarshalAmino(r []SkipZeroElem) error {
	l.Items = r
	return nil
}

// InterfaceHeavy: benchmarks MarshalAnyBinary2 with multiple interface fields.
type InterfaceHeavy struct {
	Field1 Interface1
	Field2 Interface1
	Field3 Interface1
	Items  []Interface1
	Name   string
}
