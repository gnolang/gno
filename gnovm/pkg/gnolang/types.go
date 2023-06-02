package gnolang

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// NOTE: TypeID() implementations are currently
// experimental, and will probably get replaced with
// some other system.  TypeID may become variable length,
// rather than contain a single Hashlet.

// ----------------------------------------
// (runtime) Type

type Type interface {
	assertType()

	Kind() Kind     // penetrates *DeclaredType & *NativeType
	TypeID() TypeID // deterministic
	String() string // for dev/debugging
	Elem() Type     // for TODO... types
	GetPkgPath() string
}

type TypeID string

func (tid TypeID) IsZero() bool {
	return tid == ""
}

func (tid TypeID) Bytes() []byte {
	return []byte(tid)
}

func (tid TypeID) String() string {
	return string(tid)
}

func typeid(f string, args ...interface{}) (tid TypeID) {
	fs := fmt.Sprintf(f, args...)
	x := TypeID(fs)
	if debug {
		debug.Println("TYPEID", fs)
	}
	return x
}

// Complex types are pointers, but due to the design goal
// of the language to enable mass scale persistence, we
// cannot use pointer equality to test for type equality.
// Instead, for checking equality use the TypeID.
func (PrimitiveType) assertType()   {}
func (*PointerType) assertType()    {}
func (FieldType) assertType()       {}
func (*ArrayType) assertType()      {}
func (*SliceType) assertType()      {}
func (*StructType) assertType()     {}
func (*FuncType) assertType()       {}
func (*MapType) assertType()        {}
func (*InterfaceType) assertType()  {}
func (*TypeType) assertType()       {}
func (*DeclaredType) assertType()   {}
func (*PackageType) assertType()    {}
func (*ChanType) assertType()       {}
func (*NativeType) assertType()     {}
func (blockType) assertType()       {}
func (*tupleType) assertType()      {}
func (RefType) assertType()         {}
func (MaybeNativeType) assertType() {}

// ----------------------------------------
// Primitive types

type PrimitiveType int

const (
	InvalidType PrimitiveType = 1 << iota
	UntypedBoolType
	BoolType
	UntypedStringType
	StringType
	IntType
	Int8Type
	Int16Type
	UntypedRuneType
	Int32Type
	Int64Type
	UintType
	Uint8Type
	DataByteType
	Uint16Type
	Uint32Type
	Uint64Type
	Float32Type
	Float64Type
	UntypedBigintType
	BigintType
	UntypedBigdecType
	BigdecType
	// UintptrType
)

// Used for converting constant binary expressions.
// Smaller number means more specific.
// Spec: "If the untyped operands of a binary operation (other than a shift) are
// of different kinds, the result is of the operand's kind that appears later
// in this list: integer, rune, floating-point, complex. For example, an
// untyped integer constant divided by an untyped complex constant yields an
// untyped complex constant."
func (pt PrimitiveType) Specificity() int {
	switch pt {
	case InvalidType:
		panic("invalid type has no specificity")
	case BoolType:
		return 0
	case StringType:
		return 0
	case IntType:
		return 0
	case Int8Type:
		return 0
	case Int16Type:
		return 0
	case Int32Type:
		return 0
	case Int64Type:
		return 0
	case UintType:
		return 0
	case Uint8Type, DataByteType:
		return 0
	case Uint16Type:
		return 0
	case Uint32Type:
		return 0
	case Uint64Type:
		return 0
	case Float32Type:
		return 0
	case Float64Type:
		return 0
	case BigintType:
		return 1
	case BigdecType:
		return 2
	case UntypedBigdecType:
		return 3
	case UntypedStringType:
		return 4
	case UntypedBigintType:
		return 4
	case UntypedRuneType:
		return 5
	case UntypedBoolType:
		return 6
	default:
		panic(fmt.Sprintf("unexpected primitive type %v", pt))
	}
}

func (pt PrimitiveType) Kind() Kind {
	switch pt {
	case InvalidType:
		panic("invalid type has no kind")
	case BoolType, UntypedBoolType:
		return BoolKind
	case StringType, UntypedStringType:
		return StringKind
	case IntType:
		return IntKind
	case Int8Type:
		return Int8Kind
	case Int16Type:
		return Int16Kind
	case Int32Type, UntypedRuneType:
		return Int32Kind
	case Int64Type:
		return Int64Kind
	case UintType:
		return UintKind
	case Uint8Type, DataByteType:
		return Uint8Kind
	case Uint16Type:
		return Uint16Kind
	case Uint32Type:
		return Uint32Kind
	case Uint64Type:
		return Uint64Kind
	case Float32Type:
		return Float32Kind
	case Float64Type:
		return Float64Kind
	case BigintType, UntypedBigintType:
		return BigintKind
	case BigdecType, UntypedBigdecType:
		return BigdecKind
	default:
		panic(fmt.Sprintf("unexpected primitive type %v", pt))
	}
}

func (pt PrimitiveType) TypeID() TypeID {
	switch pt {
	case InvalidType:
		panic("invalid type has no typeid")
	case UntypedBoolType:
		return typeid("<untyped> bool")
	case BoolType:
		return typeid("bool")
	case UntypedStringType:
		return typeid("<untyped> string")
	case StringType:
		return typeid("string")
	case IntType:
		return typeid("int")
	case Int8Type:
		return typeid("int8")
	case Int16Type:
		return typeid("int16")
	case UntypedRuneType:
		return typeid("<untyped> rune")
	case Int32Type:
		return typeid("int32")
	case Int64Type:
		return typeid("int64")
	case UintType:
		return typeid("uint")
	case Uint8Type:
		return typeid("uint8")
	case DataByteType:
		// should not be persisted...
		panic("untyped data byte type has no typeid")
	case Uint16Type:
		return typeid("uint16")
	case Uint32Type:
		return typeid("uint32")
	case Uint64Type:
		return typeid("uint64")
	case Float32Type:
		return typeid("float32")
	case Float64Type:
		return typeid("float64")
	case UntypedBigintType:
		return typeid("<untyped> bigint")
	case BigintType:
		return typeid("bigint")
	case UntypedBigdecType:
		return typeid("<untyped> bigdec")
	case BigdecType:
		return typeid("bigdec")
	default:
		panic(fmt.Sprintf("unexpected primitive type %v", pt))
	}
}

func (pt PrimitiveType) String() string {
	switch pt {
	case InvalidType:
		return string("<invalid type>")
	case UntypedBoolType:
		return string("<untyped> bool")
	case BoolType:
		return string("bool")
	case UntypedStringType:
		return string("<untyped> string")
	case StringType:
		return string("string")
	case IntType:
		return string("int")
	case Int8Type:
		return string("int8")
	case Int16Type:
		return string("int16")
	case UntypedRuneType:
		return string("<untyped> int32")
	case Int32Type:
		return string("int32")
	case Int64Type:
		return string("int64")
	case UintType:
		return string("uint")
	case Uint8Type:
		return string("uint8")
	case DataByteType:
		return string("<databyte> uint8")
	case Uint16Type:
		return string("uint16")
	case Uint32Type:
		return string("uint32")
	case Uint64Type:
		return string("uint64")
	case Float32Type:
		return string("float32")
	case Float64Type:
		return string("float64")
	case UntypedBigintType:
		return string("<untyped> bigint")
	case BigintType:
		return string("bigint")
	case UntypedBigdecType:
		return string("<untyped> bigdec")
	case BigdecType:
		return string("bigdec")
	default:
		panic(fmt.Sprintf("unexpected primitive type %d", pt))
	}
}

func (pt PrimitiveType) Elem() Type {
	if pt.Kind() == StringKind {
		// NOTE: this is different than Go1.
		return Uint8Type
	} else {
		panic("non-string primitive types have no elements")
	}
}

func (pt PrimitiveType) GetPkgPath() string {
	return ""
}

// ----------------------------------------
// Field type (partial)

type Tag string

type FieldType struct {
	Name     Name
	Type     Type
	Embedded bool
	Tag      Tag
}

func (ft FieldType) Kind() Kind {
	panic("FieldType is a pseudotype of unknown kind")
}

func (ft FieldType) TypeID() TypeID {
	s := ""
	if ft.Name == "" {
		s += ft.Type.TypeID().String()
	} else {
		s += string(ft.Name) + " " + ft.Type.TypeID().String()
	}
	return typeid(s)
}

func (ft FieldType) String() string {
	tag := ""
	if ft.Tag != "" {
		tag = " " + strconv.Quote(string(ft.Tag))
	}
	if ft.Name == "" {
		return fmt.Sprintf("(embedded) %s%s", ft.Type.String(), tag)
	} else {
		return fmt.Sprintf("%s %s%s", ft.Name, ft.Type.String(), tag)
	}
}

func (ft FieldType) Elem() Type {
	panic("FieldType is a pseudotype with no elements")
}

func (ft FieldType) GetPkgPath() string {
	panic("FieldType is a pseudotype with no package path")
}

// ----------------------------------------
// FieldTypeList

type FieldTypeList []FieldType

// FieldTypeList implements sort.Interface.
func (l FieldTypeList) Len() int {
	return len(l)
}

// FieldTypeList implements sort.Interface.
func (l FieldTypeList) Less(i, j int) bool {
	iname, jname := l[i].Name, l[j].Name
	if iname == jname {
		panic(fmt.Sprintf("duplicate name found in field list: %s", iname))
	}
	return iname < jname
}

// FieldTypeList implements sort.Interface.
func (l FieldTypeList) Swap(i, j int) {
	t := l[i]
	l[i] = l[j]
	l[j] = t
}

// User should call sort for interface methods.
// XXX how though?
func (l FieldTypeList) TypeID() TypeID {
	ll := len(l)
	s := ""
	for i, ft := range l {
		if ft.Name == "" {
			s += ft.Type.TypeID().String()
		} else {
			s += string(ft.Name) + " " + ft.Type.TypeID().String()
		}
		if i != ll-1 {
			s += ";"
		}
	}
	return typeid(s)
}

// For use in fields of packages, structs, and interfaces, where any
// unexported lowercase fields are private and unequal to other package
// types.
func (l FieldTypeList) TypeIDForPackage(pkgPath string) TypeID {
	ll := len(l)
	s := ""
	for i, ft := range l {
		fn := ft.Name
		if isUpper(string(fn)) {
			s += string(fn) + " " + ft.Type.TypeID().String()
		} else {
			s += pkgPath + "." + string(fn) + " " + ft.Type.TypeID().String()
		}
		if i != ll-1 {
			s += ";"
		}
	}
	return typeid(s)
}

func (l FieldTypeList) HasUnexported() bool {
	for _, ft := range l {
		if debug {
			if ft.Name == "" {
				// incorrect usage.
				panic("should not happen")
			}
		}
		if !isUpper(string(ft.Name)) {
			return true
		}
	}
	return false
}

func (l FieldTypeList) String() string {
	ll := len(l)
	s := ""
	for i, ft := range l {
		s += string(ft.Name) + " " + ft.Type.TypeID().String()
		if i != ll-1 {
			s += ";"
		}
	}
	return s
}

func (l FieldTypeList) StringWithCommas() string {
	ll := len(l)
	s := ""
	for i, ft := range l {
		s += string(ft.Name) + " " + ft.Type.String()
		if i != ll-1 {
			s += ","
		}
	}
	return s
}

// Like TypeID() but without considering field names;
// used for function parameters and results.
func (l FieldTypeList) UnnamedTypeID() TypeID {
	ll := len(l)
	s := ""
	for i, ft := range l {
		s += ft.Type.TypeID().String()
		if i != ll-1 {
			s += ";"
		}
	}
	return typeid(s)
}

func (l FieldTypeList) Types() []Type {
	res := make([]Type, len(l))
	for i, ft := range l {
		res[i] = ft.Type
	}
	return res
}

// ----------------------------------------
// Array type

type ArrayType struct {
	Len int
	Elt Type
	Vrd bool

	typeid TypeID
}

func (at *ArrayType) Kind() Kind {
	return ArrayKind
}

func (at *ArrayType) TypeID() TypeID {
	if at.typeid.IsZero() {
		at.typeid = typeid("[%d]%s", at.Len, at.Elt.TypeID().String())
	}
	return at.typeid
}

func (at *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", at.Len, at.Elt.String())
}

func (at *ArrayType) Elem() Type {
	return at.Elt
}

func (at *ArrayType) GetPkgPath() string {
	return ""
}

// ----------------------------------------
// Slice type

var gByteSliceType = &SliceType{
	Elt: Uint8Type,
	Vrd: false,
}

type SliceType struct {
	Elt Type
	Vrd bool // used for *FuncType.HasVarg()

	typeid TypeID
}

func (st *SliceType) Kind() Kind {
	return SliceKind
}

func (st *SliceType) TypeID() TypeID {
	if st.typeid.IsZero() {
		if st.Vrd {
			st.typeid = typeid("...%s", st.Elt.TypeID().String())
		} else {
			st.typeid = typeid("[]%s", st.Elt.TypeID().String())
		}
	}
	return st.typeid
}

func (st *SliceType) String() string {
	if st.Vrd {
		return fmt.Sprintf("...%s", st.Elt.String())
	} else {
		return fmt.Sprintf("[]%s", st.Elt.String())
	}
}

func (st *SliceType) Elem() Type {
	return st.Elt
}

func (st *SliceType) GetPkgPath() string {
	return ""
}

// ----------------------------------------
// Pointer type

type PointerType struct {
	Elt Type

	typeid TypeID
}

func (pt *PointerType) Kind() Kind {
	return PointerKind
}

func (pt *PointerType) TypeID() TypeID {
	if pt.typeid.IsZero() {
		pt.typeid = typeid("*%s", pt.Elt.TypeID().String())
	}
	return pt.typeid
}

func (pt *PointerType) String() string {
	if pt == nil {
		panic("invalid nil pointer type")
	} else if pt.Elt == nil {
		panic("invalid nil pointer element type")
	} else {
		return fmt.Sprintf("*%v", pt.Elt)
	}
}

func (pt *PointerType) Elem() Type {
	return pt.Elt
}

func (pt *PointerType) GetPkgPath() string {
	return pt.Elt.GetPkgPath()
}

func (pt *PointerType) FindEmbeddedFieldType(callerPath string, n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, field Type, accessError bool,
) {
	// Recursion guard.
	if m == nil {
		m = map[Type]struct{}{pt: (struct{}{})}
	} else if _, exists := m[pt]; exists {
		return nil, false, nil, nil, false
	} else {
		m[pt] = struct{}{}
	}
	// ...
	switch cet := pt.Elt.(type) {
	case *DeclaredType, *StructType:
		// Pointer to declared types and structs
		// expose embedded methods and fields.
		// See tests/selector_test.go for examples.
		trail, hasPtr, rcvr, field, accessError = findEmbeddedFieldType(callerPath, cet, n, m)
		if trail != nil { // found
			hasPtr = true // pt *is* a pointer.
			switch trail[0].Type {
			case VPField:
				// Case 1: If trail is of form [VPField, VPField, ... VPPtrMethod],
				// that is, one or more fields followed by a pointer method,
				// convert to [VPSubrefField, VPSubrefField, ... VPDerefPtrMethod].
				if func() bool {
					for i, path := range trail {
						if i < len(trail)-1 {
							if path.Type != VPField {
								return false
							}
						} else {
							if path.Type != VPPtrMethod {
								return false
							}
						}
					}
					return true
				}() {
					for i := range trail {
						if i < len(trail)-1 {
							trail[i].Type = VPSubrefField
						} else {
							trail[i].Type = VPDerefPtrMethod
						}
					}
					return
				} else {
					// Case 2: otherwise, is just a deref field.
					trail[0].Type = VPDerefField
					switch trail[0].Depth {
					case 0:
						// *PointerType > *StructType.Field has depth 0.
					case 1:
						// *DeclaredType > *StructType.Field has depth 1 (& type VPField).
						// *PointerType > *DeclaredType > *StructType.Field has depth 2.
						trail[0].Depth = 2
						/*
							// If trail[-1].Type == VPPtrMethod, set VPDerefPtrMethod.
							if len(trail) > 1 && trail[1].Type == VPPtrMethod {
								trail[1].Type = VPDerefPtrMethod
							}
						*/
					default:
						panic("should not happen")
					}
					return
				}
			case VPValMethod:
				trail[0].Type = VPDerefValMethod
				return
			case VPPtrMethod:
				trail[0].Type = VPDerefPtrMethod
				return
			case VPDerefValMethod, VPDerefPtrMethod:
				panic("should not happen")
			default:
				panic("should not happen")
			}
		} else { // not found
			return
		}
	case *NativeType:
		npt := &NativeType{
			Type: reflect.PtrTo(cet.Type),
		}
		return npt.FindEmbeddedFieldType(n, m)
	default:
		// nester pointers or pointer to interfaces
		// and other pointer types do not expose their methods.
		return
	}
}

// ----------------------------------------
// Struct type

type StructType struct {
	PkgPath string
	Fields  []FieldType

	typeid TypeID
}

func (st *StructType) Kind() Kind {
	return StructKind
}

func (st *StructType) TypeID() TypeID {
	if st.typeid.IsZero() {
		// NOTE Struct types expressed or declared in different packages
		// may have the same TypeID if and only if neither have
		// unexported fields.  st.PkgPath is only included in field
		// names that are not uppercase.
		st.typeid = typeid(
			"struct{%s}",
			FieldTypeList(st.Fields).TypeIDForPackage(st.PkgPath),
		)
	}
	return st.typeid
}

func (st *StructType) String() string {
	return fmt.Sprintf("struct{%s}",
		FieldTypeList(st.Fields).String())
}

func (st *StructType) Elem() Type {
	panic("struct types have no (universal) elements")
}

func (st *StructType) GetPkgPath() string {
	return st.PkgPath
}

// NOTE only works for exposed non-embedded fields.
func (st *StructType) GetPathForName(n Name) ValuePath {
	for i := 0; i < len(st.Fields); i++ {
		ft := st.Fields[i]
		if ft.Name == n {
			if i > 2<<16-1 {
				panic("too many fields")
			}
			return NewValuePathField(0, uint16(i), n)
		}
		if st, ok := ft.Type.(*StructType); ok {
			if ft.Name != "" {
				// skip fields not promoted.
				i += len(st.Fields)
			}
		}
	}
	panic(fmt.Sprintf("struct type %s has no field %s",
		st.String(), n))
}

func (st *StructType) GetStaticTypeOfAt(path ValuePath) Type {
	if debug {
		if path.Depth != 0 {
			panic("expected path.Depth of 0")
		}
	}
	return st.Fields[path.Index].Type
}

// Searches embedded fields to find matching method or field,
// which may be embedded. This function is slow. DeclaredType uses
// this. There is probably no need to cache positive results here;
// it may be better to implement it on DeclaredType. The resulting
// ValuePaths may be modified.  If not found, all returned values
// are nil; for consistency, check the trail.
func (st *StructType) FindEmbeddedFieldType(callerPath string, n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, field Type, accessError bool,
) {
	// Recursion guard
	if m == nil {
		m = map[Type]struct{}{st: (struct{}{})}
	} else if _, exists := m[st]; exists {
		return nil, false, nil, nil, false
	} else {
		m[st] = struct{}{}
	}
	// Search fields.
	for i := 0; i < len(st.Fields); i++ {
		sf := &st.Fields[i]
		// Maybe is a field of the struct.
		if sf.Name == n {
			// Ensure exposed or package match.
			if !isUpper(string(n)) && st.PkgPath != callerPath {
				return nil, false, nil, nil, true
			}
			vp := NewValuePathField(0, uint16(i), n)
			return []ValuePath{vp}, false, nil, sf.Type, false
		}
		// Maybe is embedded within a field.
		if sf.Embedded {
			st := sf.Type
			trail2, hasPtr2, rcvr2, field2, accessError2 := findEmbeddedFieldType(callerPath, st, n, m)
			if accessError2 {
				// XXX make test case and check against go
				return nil, false, nil, nil, true
			} else if trail2 != nil {
				if trail != nil {
					// conflict detected. return none.
					return nil, false, nil, nil, false
				} else {
					// remember.
					vp := NewValuePathField(0, uint16(i), sf.Name)
					trail, hasPtr, rcvr, field = append([]ValuePath{vp}, trail2...), hasPtr2, rcvr2, field2
				}
			}
		}
	}
	return // may be found or nil.
}

// ----------------------------------------
// Package type

// The package type holds no data.
// The PackageNode holds static declarations,
// and the PackageValue embeds a block.
var gPackageType = &PackageType{}

type PackageType struct {
	typeid TypeID
}

func (pt *PackageType) Kind() Kind {
	return PackageKind
}

func (pt *PackageType) TypeID() TypeID {
	if pt.typeid.IsZero() {
		// NOTE Different package types may have the same
		// TypeID if and only if neither have unexported fields.
		// pt.Path is only included in field names that are not
		// uppercase.
		pt.typeid = typeid("package{}")
	}
	return pt.typeid
}

func (pt *PackageType) String() string {
	return "package{}"
}

func (pt *PackageType) Elem() Type {
	panic("package types have no elements")
}

func (pt *PackageType) GetPkgPath() string {
	panic("package types has no package path (unlike package values)")
}

// ----------------------------------------
// Interface type

type InterfaceType struct {
	PkgPath string
	Methods []FieldType
	Generic Name // for uverse "generics"

	typeid TypeID
}

// General empty interface.
var gEmptyInterfaceType *InterfaceType = &InterfaceType{}

func (it *InterfaceType) IsEmptyInterface() bool {
	return len(it.Methods) == 0
}

func (it *InterfaceType) Kind() Kind {
	return InterfaceKind
}

func (it *InterfaceType) TypeID() TypeID {
	if debug {
		if it.Generic != "" {
			panic("generic type has no TypeID")
		}
	}
	if it.typeid.IsZero() {
		// NOTE Interface types expressed or declared in different
		// packages may have the same TypeID if and only if
		// neither have unexported fields.  pt.Path is only
		// included in field names that are not uppercase.
		ms := FieldTypeList(it.Methods)
		// XXX pre-sort.
		sort.Sort(ms)
		it.typeid = typeid("interface{" + ms.TypeIDForPackage(it.PkgPath).String() + "}")
	}
	return it.typeid
}

func (it *InterfaceType) String() string {
	if it.Generic != "" {
		return fmt.Sprintf("<%s>{%s}",
			it.Generic,
			FieldTypeList(it.Methods).String())
	} else {
		return fmt.Sprintf("interface{%s}",
			FieldTypeList(it.Methods).String())
	}
}

func (it *InterfaceType) Elem() Type {
	panic("interface types have no elements")
}

func (it *InterfaceType) GetPkgPath() string {
	return it.PkgPath
}

func (it *InterfaceType) FindEmbeddedFieldType(callerPath string, n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, ft Type, accessError bool,
) {
	// Recursion guard
	if m == nil {
		m = map[Type]struct{}{it: (struct{}{})}
	} else if _, exists := m[it]; exists {
		return nil, false, nil, nil, false
	} else {
		m[it] = struct{}{}
	}
	// ...
	for _, im := range it.Methods {
		if im.Name == n {
			// Ensure exposed or package match.
			if !isUpper(string(n)) && it.PkgPath != callerPath {
				return nil, false, nil, nil, true
			}
			// a matched name cannot be an embedded interface.
			if im.Type.Kind() == InterfaceKind {
				return nil, false, nil, nil, false
			}
			// match found.
			tr := []ValuePath{NewValuePathInterface(n)}
			hasPtr := false
			rcvr := Type(nil)
			ft := im.Type
			return tr, hasPtr, rcvr, ft, false
		}
		if et, ok := baseOf(im.Type).(*InterfaceType); ok {
			// embedded interfaces must be recursively searched.
			trail, hasPtr, rcvr, ft, accessError = et.FindEmbeddedFieldType(callerPath, n, m)
			if accessError {
				// XXX make test case and check against go
				return nil, false, nil, nil, true
			} else if trail != nil {
				if debug {
					if len(trail) != 1 || trail[0].Type != VPInterface {
						panic("should not happen")
					}
				}
				return trail, hasPtr, rcvr, ft, false
			} // else continue search.
		} // else continue search.
	}
	return nil, false, nil, nil, false
}

// For run-time type assertion.
// TODO: optimize somehow.
func (it *InterfaceType) IsImplementedBy(ot Type) (result bool) {
	for _, im := range it.Methods {
		if im.Type.Kind() == InterfaceKind {
			// field is embedded interface...
			im2 := baseOf(im.Type).(*InterfaceType)
			if !im2.IsImplementedBy(ot) {
				return false
			} else {
				continue
			}
		}
		// find method in field.
		tr, hp, rt, ft, _ := findEmbeddedFieldType(it.PkgPath, ot, im.Name, nil)
		if tr == nil { // not found.
			return false
		}
		if nft, ok := ft.(*NativeType); ok {
			// Treat native function types as autoNative calls.
			// ft: possibly gonative function type.
			// gnot: the corresponding gno type (GnoType()).
			// im.Type: the desired interface gno type.
			// ie, if each of ft's arg types can match
			// against the desired arg types in im.Types.
			if !gno2GoTypeMatches(im.Type, nft.Type) {
				return false
			}
		} else if mt, ok := ft.(*FuncType); ok {
			// if method is pointer receiver, check addressability:
			if _, ptrRcvr := rt.(*PointerType); ptrRcvr && !hp {
				return false // not addressable.
			}
			// check for func type equality.
			dmtid := mt.TypeID()
			imtid := im.Type.TypeID()
			if dmtid != imtid {
				return false
			}
		}
	}
	return true
}

func (it *InterfaceType) GetPathForName(n Name) ValuePath {
	return NewValuePathInterface(n)
}

// ----------------------------------------
// Chan type

type ChanType struct {
	Dir ChanDir
	Elt Type

	typeid TypeID
}

func (ct *ChanType) Kind() Kind {
	return ChanKind
}

func (ct *ChanType) TypeID() TypeID {
	if ct.typeid.IsZero() {
		switch ct.Dir {
		case SEND | RECV:
			ct.typeid = typeid("chan{%s}" + ct.Elt.TypeID().String())
		case SEND:
			ct.typeid = typeid("<-chan{%s}" + ct.Elt.TypeID().String())
		case RECV:
			ct.typeid = typeid("chan<-{%s}" + ct.Elt.TypeID().String())
		default:
			panic("should not happen")
		}
	}
	return ct.typeid
}

func (ct *ChanType) String() string {
	switch ct.Dir {
	case SEND | RECV:
		return "chan " + ct.Elt.String()
	case SEND:
		return "<-chan " + ct.Elt.String()
	case RECV:
		return "chan<- " + ct.Elt.String()
	default:
		panic("should not happen")
	}
}

func (ct *ChanType) Elem() Type {
	return ct.Elt
}

func (ct *ChanType) GetPkgPath() string {
	return ""
}

// ----------------------------------------
// Function type

type FuncType struct {
	Params  []FieldType
	Results []FieldType

	typeid TypeID
	bound  *FuncType
}

// if ft is a method, returns whether method takes a pointer receiver.
func (ft *FuncType) HasPointerReceiver() bool {
	if debug {
		if len(ft.Params) == 0 {
			panic("expected unbound method function type, but found no receiver parameter.")
		}
	}
	_, ok := ft.Params[0].Type.(*PointerType)
	return ok
	// return ft.Params[0].Type.Kind() == PointerKind
}

func (ft *FuncType) Kind() Kind {
	return FuncKind
}

// bound function type (if ft is a method).
func (ft *FuncType) BoundType() *FuncType {
	if ft.bound == nil {
		ft.bound = &FuncType{
			Params:  ft.Params[1:],
			Results: ft.Results,
		}
	}
	return ft.bound
}

// unbound function type
func (ft *FuncType) UnboundType(rft FieldType) *FuncType {
	return &FuncType{
		Params:  append([]FieldType{rft}, ft.Params...),
		Results: ft.Results,
	}
}

// given the call arg types (and whether is ...varg), specify any
// generic types to return the ultimate specified func type.
// Any untyped arg types are first converted to its default type.
// NOTE: if ft.HasVarg() and !isVarg, argTVs[len(ft.Params):]
// are ignored (since they are of the same type as
// argTVs[len(ft.Params)-1]).
func (ft *FuncType) Specify(store Store, argTVs []TypedValue, isVarg bool) *FuncType {
	hasGenericParams := false
	hasGenericResults := false
	for _, pf := range ft.Params {
		if isGeneric(pf.Type) {
			hasGenericParams = true
			break
		}
	}
	for _, rf := range ft.Results {
		if isGeneric(rf.Type) {
			hasGenericResults = true
			break
		}
	}
	if !hasGenericParams && hasGenericResults {
		panic("function with generic results require matching generic params")
	}
	if !hasGenericParams && !hasGenericResults {
		return ft // no changes.
	}
	lookup := map[Name]Type{}
	hasVarg := ft.HasVarg()
	if hasVarg && !isVarg {
		if isGeneric(ft.Params[len(ft.Params)-1].Type) {
			// consolidate vargs into slice.
			var nvarg int
			var vargt Type
			for i := len(ft.Params) - 1; i < len(argTVs); i++ {
				nvarg++
				varg := argTVs[i]
				if varg.T == nil {
					continue
				} else if vargt == nil {
					vargt = varg.T
				} else if isUntyped(varg.T) && vargt.TypeID() == defaultTypeOf(varg.T).TypeID() {
					vargt = defaultTypeOf(varg.T)
				} else if vargt.TypeID() != varg.T.TypeID() {
					panic(fmt.Sprintf(
						"incompatible varg types: expected %v, got %s",
						vargt.String(),
						varg.T.String()))
				}
			}
			if nvarg > 0 && vargt == nil {
				panic(fmt.Sprintf(
					"unspecified generic varg %s",
					ft.Params[len(ft.Params)-1].String()))
			}
			argTVs = argTVs[:len(ft.Params)-1]
			argTVs = append(argTVs, TypedValue{
				T: &SliceType{Elt: vargt, Vrd: true},
				V: nil,
			})
		} else {
			// just use already specific type.
			argTVs = argTVs[:len(ft.Params)-1]
			argTVs = append(argTVs, TypedValue{
				T: ft.Params[len(ft.Params)-1].Type,
				V: nil,
			})
		}
	}
	// specify generic types from args.
	for i, pf := range ft.Params {
		arg := &argTVs[i]
		if arg.T.Kind() == TypeKind {
			specifyType(store, lookup, pf.Type, arg.T, arg.GetType())
		} else {
			specifyType(store, lookup, pf.Type, arg.T, nil)
		}
	}
	// apply specifics to generic params and results.
	pfts := make([]FieldType, len(ft.Params))
	rfts := make([]FieldType, len(ft.Results))
	for i, pft := range ft.Params {
		// special case for maybenative, just take the native type.
		if mnpft, ok := pft.Type.(*MaybeNativeType); ok {
			if nt, ok := argTVs[i].T.(*NativeType); ok {
				pfts[i] = FieldType{
					Name: pft.Name,
					Type: nt,
				}
			} else {
				pt, _ := applySpecifics(lookup, mnpft.Type)
				pfts[i] = FieldType{
					Name: pft.Name,
					Type: pt,
				}
			}
			continue
		}
		// default case.
		pt, _ := applySpecifics(lookup, pft.Type)
		pfts[i] = FieldType{
			Name: pft.Name,
			Type: pt,
		}
	}
	for i, rft := range ft.Results {
		rt, _ := applySpecifics(lookup, rft.Type)
		rfts[i] = FieldType{
			Name: rft.Name,
			Type: rt,
		}
	}
	return &FuncType{
		Params:  pfts,
		Results: rfts,
	}
}

func (ft *FuncType) TypeID() TypeID {
	// Two functions of different realms can have the same
	// type, because the method signature doesn't change, and
	// this exchangeability is useful to denote type semantics.
	ps := FieldTypeList(ft.Params)
	rs := FieldTypeList(ft.Results)
	/*
		pp := ""
		if ps.HasUnexported() || rs.HasUnexported() {
			pp = fmt.Sprintf("@%q", ft.PkgPath)
		}
	*/
	if ft.typeid.IsZero() {
		ft.typeid = typeid(
			"func(%s)(%s)",
			// pp,
			ps.UnnamedTypeID(),
			rs.UnnamedTypeID(),
		)
	}
	return ft.typeid
}

func (ft *FuncType) String() string {
	return fmt.Sprintf("func(%s)(%s)",
		FieldTypeList(ft.Params).StringWithCommas(),
		FieldTypeList(ft.Results).StringWithCommas())
}

func (ft *FuncType) Elem() Type {
	panic("function types have no elements")
}

func (ft *FuncType) GetPkgPath() string {
	panic("function types have no package path")
}

func (ft *FuncType) HasVarg() bool {
	if numParams := len(ft.Params); numParams == 0 {
		return false
	} else {
		lpt := ft.Params[numParams-1].Type
		if lat, ok := lpt.(*SliceType); ok {
			return lat.Vrd
		} else if mnt, ok := lpt.(*MaybeNativeType); ok {
			if lat, ok := mnt.Type.(*SliceType); ok {
				return lat.Vrd
			} else {
				return false
			}
		} else {
			return false
		}
	}
}

// ----------------------------------------
// Map type

type MapType struct {
	Key   Type
	Value Type

	typeid TypeID
}

func (mt *MapType) Kind() Kind {
	return MapKind
}

func (mt *MapType) TypeID() TypeID {
	if mt.typeid.IsZero() {
		mt.typeid = typeid(
			"map[%s]%s",
			mt.Key.TypeID().String(),
			mt.Value.TypeID().String(),
		)
	}
	return mt.typeid
}

func (mt *MapType) String() string {
	return fmt.Sprintf("map[%s]%s",
		mt.Key.String(),
		mt.Value.String())
}

func (mt *MapType) Elem() Type {
	return mt.Value
}

func (mt *MapType) GetPkgPath() string {
	return ""
}

// ----------------------------------------
// Type (typeval) type

type TypeType struct { // nothing yet.
}

var gTypeType = &TypeType{}

func (tt *TypeType) Kind() Kind {
	return TypeKind
}

func (tt *TypeType) TypeID() TypeID {
	return typeid("type{}")
}

func (tt *TypeType) String() string {
	return string("type{}")
}

func (tt *TypeType) Elem() Type {
	panic("typeval types have no elements")
}

func (tt *TypeType) GetPkgPath() string {
	panic("typeval types have no package path")
}

// ----------------------------------------
// Declared type
// Declared types have a name, base (underlying) type,
// and associated methods.

type DeclaredType struct {
	PkgPath string
	Name    Name
	Base    Type         // not a DeclaredType
	Methods []TypedValue // {T:*FuncType,V:*FuncValue}...

	typeid TypeID
	sealed bool // for ensuring correctness with recursive types.
}

// returns an unsealed *DeclaredType.
// do not use for aliases.
func declareWith(pkgPath string, name Name, b Type) *DeclaredType {
	dt := &DeclaredType{
		PkgPath: pkgPath,
		Name:    name,
		Base:    baseOf(b),
		sealed:  false,
	}
	return dt
}

func BaseOf(t Type) Type {
	return baseOf(t)
}

func baseOf(t Type) Type {
	if dt, ok := t.(*DeclaredType); ok {
		return dt.Base
	} else {
		return t
	}
}

// NOTE: it may be faster to switch on baseOf().
func (dt *DeclaredType) Kind() Kind {
	return dt.Base.Kind()
}

func (dt *DeclaredType) Seal() {
	dt.checkSeal()
	dt.sealed = true
}

// NOTE: dt.sealed is only for recursive types support:
// it is unsealed until the recursion definition is complete.
// it is not used to prevent the updating of declared types,
// such as adding new method functions.
func (dt *DeclaredType) checkSeal() {
	if dt.sealed {
		panic(fmt.Sprintf(
			"*DeclaredType %s already sealed",
			dt.Name))
	}
}

func (dt *DeclaredType) TypeID() TypeID {
	if dt.typeid.IsZero() {
		dt.typeid = DeclaredTypeID(dt.PkgPath, dt.Name)
	}
	return dt.typeid
}

func DeclaredTypeID(pkgPath string, name Name) TypeID {
	return typeid("%s.%s", pkgPath, name)
}

func (dt *DeclaredType) String() string {
	return fmt.Sprintf("%s.%s", dt.PkgPath, dt.Name)
}

func (dt *DeclaredType) Elem() Type {
	return dt.Base.Elem()
}

func (dt *DeclaredType) GetPkgPath() string {
	return dt.PkgPath
}

func (dt *DeclaredType) DefineMethod(fv *FuncValue) {
	dt.Methods = append(dt.Methods, TypedValue{
		T: fv.Type,
		V: fv,
	})
}

func (dt *DeclaredType) GetPathForName(n Name) ValuePath {
	// May be a method.
	for i, tv := range dt.Methods {
		fv := tv.V.(*FuncValue)
		if fv.Name == n {
			if i > 2<<16-1 {
				panic("too many methods")
			}
			// NOTE: makes code simple but requires preprocessor's
			// Store to pre-load method types.
			if fv.GetType(nil).HasPointerReceiver() {
				return NewValuePathPtrMethod(uint16(i), n)
			} else {
				return NewValuePathValMethod(uint16(i), n)
			}
		}
	}
	// Otherwise it is underlying.
	path := dt.Base.(ValuePather).GetPathForName(n)
	path.Depth += 1
	return path
}

func (dt *DeclaredType) GetUnboundPathForName(n Name) ValuePath {
	for i, tv := range dt.Methods {
		fv := tv.V.(*FuncValue)
		if fv.Name == n {
			if i > 2<<16-1 {
				panic("too many methods")
			}
			return NewValuePathField(0, uint16(i), n)
		}
	}
	panic(fmt.Sprintf(
		"unknown *DeclaredType method named %s",
		n))
}

// Searches embedded fields to find matching field or method.
// This function is slow.
// TODO: consider memoizing for successful matches.
func (dt *DeclaredType) FindEmbeddedFieldType(callerPath string, n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, ft Type, accessError bool,
) {
	// Recursion guard
	if m == nil {
		m = map[Type]struct{}{dt: (struct{}{})}
	} else if _, exists := m[dt]; exists {
		return nil, false, nil, nil, false
	} else {
		m[dt] = struct{}{}
	}
	// Search direct methods.
	for i := 0; i < len(dt.Methods); i++ {
		mv := &dt.Methods[i]
		if fv := mv.GetFunc(); fv.Name == n {
			// Ensure exposed or package match.
			if !isUpper(string(n)) && dt.PkgPath != callerPath {
				return nil, false, nil, nil, true
			}
			// NOTE: makes code simple but requires preprocessor's
			// Store to pre-load method types.
			rt := fv.GetType(nil).Params[0].Type
			vp := ValuePath{}
			if _, ok := rt.(*PointerType); ok {
				vp = NewValuePathPtrMethod(uint16(i), n)
			} else {
				vp = NewValuePathValMethod(uint16(i), n)
			}
			// NOTE: makes code simple but requires preprocessor's
			// Store to pre-load method types.
			bt := fv.GetType(nil).BoundType()
			return []ValuePath{vp}, false, rt, bt, false
		}
	}
	// Otherwise, search base.
	trail, hasPtr, rcvr, ft, accessError = findEmbeddedFieldType(callerPath, dt.Base, n, m)
	if trail == nil {
		return nil, false, nil, nil, accessError
	}
	switch trail[0].Type {
	case VPInterface:
		return trail, hasPtr, rcvr, ft, false
	case VPField, VPDerefField:
		if debug {
			if trail[0].Depth != 0 && trail[0].Depth != 2 {
				panic("should not happen")
			}
		}
		trail[0].Depth += 1
		return trail, hasPtr, rcvr, ft, false
	default:
		panic("should not happen")
	}
}

// The Preprocesses uses *DT.FindEmbeddedFieldType() to set the path.
// OpSelector uses *TV.GetPointerTo(path), and for declared types, in turn
// uses *DT.GetValueAt(path) to find any methods (see values.go).
//
// i.e.,
//
//	preprocessor: *DT.FindEmbeddedFieldType(name)
//	              *DT.GetValueAt(path) // from op_type/evalTypeOf()
//
//	     runtime: *TV.GetPointerTo(path)
//	               -> *DT.GetValueAt(path)
func (dt *DeclaredType) GetValueAt(alloc *Allocator, store Store, path ValuePath) TypedValue {
	switch path.Type {
	case VPInterface:
		panic("should not happen")
		// should call *DT.FindEmbeddedFieldType(name) instead.
		// tr, hp, rt, ft := dt.FindEmbeddedFieldType(n)
	case VPValMethod, VPPtrMethod, VPField:
		if path.Depth == 0 {
			mtv := dt.Methods[path.Index]
			// Fill in *FV.Closure.
			ft := mtv.T
			fv := mtv.V.(*FuncValue).Copy(alloc)
			fv.Closure = fv.GetClosure(store)
			return TypedValue{T: ft, V: fv}
		} else {
			panic("DeclaredType.GetValueAt() expects depth == 0")
		}
	default:
		panic(fmt.Sprintf(
			"unexpected value path type %s",
			path.String()))
	}
}

// Like GetValueAt, but doesn't fill *FuncValue closures.
func (dt *DeclaredType) GetStaticValueAt(path ValuePath) TypedValue {
	switch path.Type {
	case VPInterface:
		panic("should not happen")
		// should call *DT.FindEmbeddedFieldType(name) instead.
		// tr, hp, rt, ft := dt.FindEmbeddedFieldType(n)
	case VPValMethod, VPPtrMethod, VPField:
		if path.Depth == 0 {
			return dt.Methods[path.Index]
		} else {
			panic("DeclaredType.GetStaticValueAt() expects depth == 0")
		}
	default:
		panic(fmt.Sprintf(
			"unexpected value path type %s",
			path.String()))
	}
}

// ----------------------------------------
// Native type

type NativeType struct {
	Type reflect.Type // Go "native" type

	typeid  TypeID
	gnoType Type // Gno converted type
}

func (nt *NativeType) Kind() Kind {
	switch nt.Type.Kind() {
	case reflect.Bool:
		return BoolKind
	case reflect.String:
		return StringKind
	case reflect.Int:
		return IntKind
	case reflect.Int8:
		return Int8Kind
	case reflect.Int16:
		return Int16Kind
	case reflect.Int32:
		return Int32Kind
	case reflect.Int64:
		return Int64Kind
	case reflect.Uint:
		return UintKind
	case reflect.Uint8:
		return Uint8Kind
	case reflect.Uint16:
		return Uint16Kind
	case reflect.Uint32:
		return Uint32Kind
	case reflect.Uint64:
		return Uint64Kind
	case reflect.Array:
		return ArrayKind
	case reflect.Chan:
		return ChanKind
	case reflect.Func:
		return FuncKind
	case reflect.Map:
		return MapKind
	case reflect.Ptr:
		return PointerKind
	case reflect.Slice:
		return SliceKind
	case reflect.Struct:
		return StructKind
	case reflect.Interface:
		return InterfaceKind
	default:
		panic(fmt.Sprintf(
			"unexpected native kind %v for type %v",
			nt.Type.Kind(), nt.Type))
	}
}

func (nt *NativeType) TypeID() TypeID {
	// like a DeclaredType, but different.
	if nt.typeid.IsZero() {
		if nt.Type.Name() == "" {
			// TODO try to derive better name specification,
			// current Golang one is undefined.
			// > String returns a string representation of the type.
			// > The string representation may use shortened package names
			// > (e.g., base64 instead of "encoding/base64") and is not
			// > guaranteed to be unique among types. To test for type identity,
			// > compare the Types directly.
			nt.typeid = typeid("go:%s.%s", nt.Type.PkgPath(), nt.Type.String())
		} else {
			nt.typeid = typeid("go:%s.%s", nt.Type.PkgPath(), nt.Type.Name())
		}
	}
	return nt.typeid
}

func (nt *NativeType) String() string {
	return fmt.Sprintf("gonative{%s}", nt.Type.String())
}

// TODO: memoize?
func (nt *NativeType) Key() Type {
	switch nt.Type.Kind() {
	case reflect.Map:
		return go2GnoType(nt.Type.Key())
	default:
		panic(fmt.Sprintf("unexpected native type %v for .Key",
			nt.Type.String()))
	}
}

// TODO: memoize?
func (nt *NativeType) Elem() Type {
	switch nt.Type.Kind() {
	case reflect.Ptr, reflect.Array, reflect.Slice, reflect.Map:
		return go2GnoType(nt.Type.Elem())
	default:
		panic(fmt.Sprintf("unexpected native type %v for .Elem",
			nt.Type.String()))
	}
}

func (nt *NativeType) GetPkgPath() string {
	return "go:" + nt.Type.PkgPath()
}

func (nt *NativeType) GnoType(store Store) Type {
	if nt.gnoType == nil {
		nt.gnoType = store.Go2GnoType(nt.Type)
	}
	return nt.gnoType
}

// TODO implement accessError return value.
func (nt *NativeType) FindEmbeddedFieldType(n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, field Type, accessError bool,
) {
	// Recursion guard
	if m == nil {
		m = map[Type]struct{}{nt: (struct{}{})}
	} else if _, exists := m[nt]; exists {
		return nil, false, nil, nil, false
	} else {
		m[nt] = struct{}{}
	}
	// special cases for pointer to struct and interface.
	var rt reflect.Type = nt.Type
	if rt.Kind() == reflect.Ptr {
		// match on pointer to field
		ert := rt.Elem()
		rft, ok := ert.FieldByName(string(n))
		if ok {
			trail = []ValuePath{NewValuePathNative(n)}
			hasPtr = true
			rcvr = nil
			field = go2GnoType(rft.Type)
			return
		} else {
			// deref and continue...
			hasPtr = true
		}
	} else if rt.Kind() == reflect.Interface {
		// match on interface.
		rmt, ok := rt.MethodByName(string(n))
		if ok {
			trail = []ValuePath{NewValuePathNative(n)}
			rcvr = nil
			field = go2GnoType(rmt.Type)
			return
		} else { // no match
			return nil, false, nil, nil, false
		}
	}
	// match method on non-interface type.
	rmt, ok := rt.MethodByName(string(n))
	if ok {
		trail = []ValuePath{NewValuePathNative(n)}
		if rmt.Type.In(0).Kind() == reflect.Ptr {
			if debug {
				if !hasPtr {
					panic("should not happen")
				}
			}
			rcvr = nt
		} else {
			if hasPtr {
				rcvr = &NativeType{
					Type: nt.Type.Elem(),
				} // XXX inefficient new/alloc.
			} else {
				rcvr = nt
			}
		}
		{ // construct bound function type
			// TODO cache.
			numIns := rmt.Type.NumIn() - 1 // bound
			ins := make([]reflect.Type, numIns)
			for i := 0; i < numIns; i++ {
				ins[i] = rmt.Type.In(i + 1) // skip receiver
			}
			numOuts := rmt.Type.NumOut()
			outs := make([]reflect.Type, numOuts)
			for i := 0; i < numOuts; i++ {
				outs[i] = rmt.Type.Out(i)
			}
			variadic := rmt.Type.IsVariadic()
			brmt := reflect.FuncOf(ins, outs, variadic) // bound reflect method type
			field = go2GnoType(brmt)
		}
		return
	}
	// match field on struct.
	if rt.Kind() == reflect.Struct {
		// match on field.
		rft, ok := rt.FieldByName(string(n))
		if ok {
			trail = []ValuePath{NewValuePathNative(n)}
			hasPtr = false
			rcvr = nil
			field = go2GnoType(rft.Type)
			return
		} else { // no match
			return nil, false, nil, nil, false
		}
	}
	// no match
	return nil, false, nil, nil, false
}

// ----------------------------------------
// blockType

type blockType struct{} // no data

func (bt blockType) Kind() Kind {
	return BlockKind
}

func (bt blockType) TypeID() TypeID {
	return typeid("block")
}

func (bt blockType) String() string {
	return "block"
}

func (bt blockType) Elem() Type {
	panic("blockType has no elem type")
}

func (bt blockType) GetPkgPath() string {
	panic("blockType has no package path")
}

// ----------------------------------------
// tupleType

type tupleType struct {
	Elts []Type

	typeid TypeID
}

func (tt *tupleType) Kind() Kind {
	return TupleKind
}

func (tt *tupleType) TypeID() TypeID {
	if tt.typeid.IsZero() {
		ell := len(tt.Elts)
		s := "("
		for i, et := range tt.Elts {
			s += et.TypeID().String()
			if i != ell-1 {
				s += ","
			}
		}
		s += ")"
		tt.typeid = typeid(s)
	}
	return tt.typeid
}

func (tt *tupleType) String() string {
	ell := len(tt.Elts)
	s := "("
	for i, et := range tt.Elts {
		s += et.String()
		if i != ell-1 {
			s += ","
		}
	}
	s += ")"
	return s
}

func (tt *tupleType) Elem() Type {
	panic("tupleType has no singular elem type")
}

func (tt *tupleType) GetPkgPath() string {
	panic("typleType has no package path")
}

// ----------------------------------------
// RefType

type RefType struct {
	ID TypeID
}

func (RefType) Kind() Kind {
	return RefTypeKind
}

func (rt RefType) TypeID() TypeID {
	return rt.ID
}

func (rt RefType) String() string {
	return fmt.Sprintf("RefType{%v}", rt.ID)
}

func (rt RefType) Elem() Type {
	panic("should not happen")
}

func (rt RefType) GetPkgPath() string {
	panic("should not happen")
}

// ----------------------------------------
// MaybeNativeType

// MaybeNativeType wraps an underlying gno type
// and allows the generic matching of spec to gno type,
// or go2GnoType2(spec) to gno type if spec is native.
type MaybeNativeType struct {
	Type
}

func (mn MaybeNativeType) Kind() Kind {
	return mn.Type.Kind()
}

func (mn MaybeNativeType) TypeID() TypeID {
	panic("MaybeNativeType type has no type id")
}

func (mn MaybeNativeType) String() string {
	return fmt.Sprintf("MaybeNativeType{%s}", mn.Type.String())
}

func (mn MaybeNativeType) Elem() Type {
	return mn.Type.Elem()
}

func (mn MaybeNativeType) GetPkgPath() string {
	return mn.Type.GetPkgPath()
}

// ----------------------------------------
// Kind

type Kind uint

const (
	InvalidKind Kind = iota
	BoolKind
	StringKind
	IntKind
	Int8Kind
	Int16Kind
	Int32Kind
	Int64Kind
	UintKind
	Uint8Kind
	Uint16Kind
	Uint32Kind
	Uint64Kind
	Float32Kind
	Float64Kind
	BigintKind // not in go.
	BigdecKind // not in go.
	// UintptrKind
	ArrayKind
	SliceKind
	PointerKind
	StructKind
	PackageKind // not in go.
	InterfaceKind
	ChanKind
	FuncKind
	MapKind
	TypeKind // not in go.
	// UnsafePointerKind
	BlockKind   // not in go.
	TupleKind   // not in go.
	RefTypeKind // not in go.
)

// This is generally slower than switching on baseOf(t).
func KindOf(t Type) Kind {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		switch t {
		case InvalidType:
			panic("invalid type has no kind")
		case BoolType, UntypedBoolType:
			return BoolKind
		case StringType, UntypedStringType:
			return StringKind
		case IntType:
			return IntKind
		case Int8Type:
			return Int8Kind
		case Int16Type:
			return Int16Kind
		case Int32Type, UntypedRuneType:
			return Int32Kind
		case Int64Type:
			return Int64Kind
		case UintType:
			return UintKind
		case Uint8Type, DataByteType:
			return Uint8Kind
		case Uint16Type:
			return Uint16Kind
		case Uint32Type:
			return Uint32Kind
		case Uint64Type:
			return Uint64Kind
		case Float32Type:
			return Float32Kind
		case Float64Type:
			return Float64Kind
		case BigintType, UntypedBigintType:
			return BigintKind
		case BigdecType, UntypedBigdecType:
			return BigdecKind
		default:
			panic(fmt.Sprintf("unexpected primitive type %s", t.String()))
		}
	case *DeclaredType:
		panic("unexpected nested DeclaredType")
	case FieldType:
		panic("FieldType is a pseudotype")
	case *ArrayType:
		return ArrayKind
	case *SliceType:
		return SliceKind
	case *PointerType:
		return PointerKind
	case *StructType:
		return StructKind
	case *PackageType:
		return PackageKind
	case *InterfaceType:
		return InterfaceKind
	case *ChanType:
		return ChanKind
	case *FuncType:
		return FuncKind
	case *MapType:
		return MapKind
	case *TypeType:
		return TypeKind
	case *NativeType:
		return t.Kind()
	case blockType:
		return BlockKind
	case *tupleType:
		return TupleKind
	case RefType:
		return RefTypeKind
	case MaybeNativeType:
		return t.Kind()
	default:
		panic(fmt.Sprintf("unexpected type %#v", t))
	}
}

// ----------------------------------------
// main type-assertion functions.

// TODO: document what class of problems its for.
// One of them can be nil, and this lets uninitialized primitives
// and others serve as empty values.  See doOpAdd()
// usage: if debug { assertSameTypes() }
func assertSameTypes(lt, rt Type) {
	if lt == nil && rt == nil {
		// both are nil.
	} else if lt == nil || rt == nil {
		// one is nil.  see function comment.
	} else if lt.Kind() == rt.Kind() &&
		isUntyped(lt) || isUntyped(rt) {
		// one is untyped of same kind.
	} else if lt.Kind() == rt.Kind() &&
		isDataByte(lt) {
		// left is databyte of same kind,
		// specifically for assignments.
		// TODO: make another function
		// and remove this case?
	} else if lt.TypeID() == rt.TypeID() {
		// non-nil types are identical.
	} else {
		debug.Errorf(
			"incompatible operands in binary expression: %s and %s",
			lt.String(),
			rt.String(),
		)
	}
}

// Like assertSameTypes(), but more relaxed, for == and !=.
func assertEqualityTypes(lt, rt Type) {
	if lt == nil && rt == nil {
		// both are nil.
	} else if lt == nil || rt == nil {
		// one is nil.  see function comment.
	} else if lt.Kind() == rt.Kind() &&
		isUntyped(lt) || isUntyped(rt) {
		// one is untyped of same kind.
	} else if lt.Kind() == InterfaceKind &&
		IsImplementedBy(lt, rt) {
		// rt implements lt (and lt is nil interface).
	} else if rt.Kind() == InterfaceKind &&
		IsImplementedBy(rt, lt) {
		// lt implements rt (and rt is nil interface).
	} else if lt.TypeID() == rt.TypeID() {
		// non-nil types are identical.
	} else {
		debug.Errorf(
			"incompatible operands in binary (eql/neq) expression: %s and %s",
			lt.String(),
			rt.String(),
		)
	}
}

// ----------------------------------------
// misc

func isUntyped(t Type) bool {
	switch t {
	case UntypedBoolType, UntypedRuneType, UntypedBigintType, UntypedBigdecType, UntypedStringType:
		return true
	default:
		return false
	}
}

func isDataByte(t Type) bool {
	switch t {
	case DataByteType:
		return true
	default:
		return false
	}
}

// TODO move untyped const stuff to preprocess.go.
// TODO associate with ConvertTo() in documentation.
func defaultTypeOf(t Type) Type {
	switch t {
	case UntypedBoolType:
		return BoolType
	case UntypedRuneType:
		return Int32Type
	case UntypedBigintType:
		return IntType
	case UntypedBigdecType:
		return Float64Type
	case UntypedStringType:
		return StringType
	default:
		panic("unexpected type for default untyped const conversion")
	}
}

func fillEmbeddedName(ft *FieldType) {
	if ft.Name != "" {
		return
	}
	switch ct := ft.Type.(type) {
	case *PointerType:
		// dereference one level
		switch ct := ct.Elt.(type) {
		case *DeclaredType:
			ft.Name = ct.Name
		case *NativeType:
			panic("native type cannot be embedded")
		default:
			panic("should not happen")
		}
	case *DeclaredType:
		ft.Name = ct.Name
	case PrimitiveType:
		switch ct {
		case BoolType:
			ft.Name = Name("bool")
		case StringType:
			ft.Name = Name("string")
		case IntType:
			ft.Name = Name("int")
		case Int8Type:
			ft.Name = Name("int8")
		case Int16Type:
			ft.Name = Name("int16")
		case Int32Type:
			ft.Name = Name("int32")
		case Int64Type:
			ft.Name = Name("int64")
		case UintType:
			ft.Name = Name("uint")
		case Uint8Type:
			ft.Name = Name("uint8")
		case Uint16Type:
			ft.Name = Name("uint16")
		case Uint32Type:
			ft.Name = Name("uint32")
		case Uint64Type:
			ft.Name = Name("uint64")
		case Float32Type:
			ft.Name = Name("float32")
		case Float64Type:
			ft.Name = Name("float64")
		case BigintType:
			ft.Name = Name("bigint")
		case BigdecType:
			ft.Name = Name("bigdec")
		default:
			panic("should not happen")
		}
	case *NativeType:
		panic("native type cannot be embedded")
	default:
		panic(fmt.Sprintf(
			"unexpected field type %s",
			ft.Type.String()))
	}
	ft.Embedded = true
}

func IsImplementedBy(it Type, ot Type) bool {
	switch cbt := baseOf(it).(type) {
	case *InterfaceType:
		return cbt.IsImplementedBy(ot)
	case *NativeType:
		return gno2GoTypeMatches(ot, cbt.Type)
	default:
		panic("should not happen")
	}
}

// Given a map of generic type names, match the tmpl type which
// might include generics with the spec type which is concrete
// with no generics, and update the lookup map or panic if error.
// specTypeval is Type if spec is TypeKind.
// NOTE: type-checking isn't strictly necessary here, as the resulting lookup
// map gets applied to produce the ultimate param and result types.
func specifyType(store Store, lookup map[Name]Type, tmpl Type, spec Type, specTypeval Type) {
	if isGeneric(spec) {
		panic("spec must not be generic")
	}
	if st, ok := spec.(*SliceType); ok && st.Vrd {
		spec = &SliceType{
			Elt: st.Elt,
			Vrd: false,
		}
	}
	switch ct := tmpl.(type) {
	case *PointerType:
		switch pt := baseOf(spec).(type) {
		case *PointerType:
			specifyType(store, lookup, ct.Elt, pt.Elt, nil)
		case *NativeType:
			// NOTE: see note about type-checking.
			et := pt.Elem()
			specifyType(store, lookup, ct.Elt, et, nil)
		default:
			panic(fmt.Sprintf(
				"expected pointer kind but got %s",
				spec.Kind()))
		}
	case *ArrayType:
		switch at := baseOf(spec).(type) {
		case *ArrayType:
			specifyType(store, lookup, ct.Elt, at.Elt, nil)
		case *NativeType:
			// NOTE: see note about type-checking.
			et := at.Elem()
			specifyType(store, lookup, ct.Elt, et, nil)
		default:
			panic(fmt.Sprintf(
				"expected array kind but got %s",
				spec.Kind()))
		}
	case *SliceType:
		switch st := baseOf(spec).(type) {
		case PrimitiveType:
			if isGeneric(ct.Elt) {
				if st.Kind() == StringKind {
					specifyType(store, lookup, ct.Elt, Uint8Type, nil)
				} else {
					panic(fmt.Sprintf(
						"expected slice kind but got %s",
						spec.Kind()))
				}
			} else if ct.Elt != Uint8Type {
				panic(fmt.Sprintf(
					"expected slice kind but got %s",
					spec.Kind()))
			} else if st != StringType {
				panic(fmt.Sprintf(
					"expected slice kind (or string type) but got %s",
					spec.Kind()))
			}
		case *SliceType:
			specifyType(store, lookup, ct.Elt, st.Elt, nil)
		case *NativeType:
			// NOTE: see note about type-checking.
			et := st.Elem()
			specifyType(store, lookup, ct.Elt, et, nil)
		default:
			panic(fmt.Sprintf(
				"expected slice kind but got %s",
				spec.Kind()))
		}
	case *MapType:
		switch mt := baseOf(spec).(type) {
		case *MapType:
			specifyType(store, lookup, ct.Key, mt.Key, nil)
			specifyType(store, lookup, ct.Value, mt.Value, nil)
		case *NativeType:
			// NOTE: see note about type-checking.
			kt := mt.Key()
			vt := mt.Elem()
			specifyType(store, lookup, ct.Key, kt, nil)
			specifyType(store, lookup, ct.Value, vt, nil)
		default:
			panic(fmt.Sprintf(
				"expected map kind but got %s",
				spec.Kind()))
		}
	case *InterfaceType:
		if ct.Generic != "" {
			// tmpl is generic, so replace from lookup.
			if strings.HasSuffix(string(ct.Generic), ".(type)") {
				if spec.Kind() != TypeKind {
					panic(fmt.Sprintf(
						"generic <%s> requires type kind, got %v",
						ct.Generic,
						spec.Kind()))
				}
				generic := ct.Generic[:len(ct.Generic)-len(".(type)")]
				match, ok := lookup[generic]
				if ok {
					if match.TypeID() != specTypeval.TypeID() {
						panic(fmt.Sprintf(
							"expected %s for <%s> but got %s",
							match.String(),
							ct.Generic,
							specTypeval.String()))
					} else {
						return // ok
					}
				} else {
					lookup[generic] = specTypeval
					return // ok
				}
			} else if strings.HasSuffix(string(ct.Generic), ".Elem()") {
				if spec.Kind() == TypeKind {
					panic("generic <%s> does not expect type kind")
				}
				generic := ct.Generic[:len(ct.Generic)-len(".Elem()")]
				match, ok := lookup[generic]
				if ok {
					checkType(spec, match.Elem(), false)
					return // ok
				} else {
					// Panic here, because we don't know whether T
					// should be native or gno yet.
					// It may be possible to allow lazy specification
					// with some changes, but it isn't obvious.
					panic("T.Elem generic must follow specification of T")
					// lookup[generic] = specTypeval
					// return // ok
				}
			} else {
				match, ok := lookup[ct.Generic]
				if ok {
					checkType(spec, match, false)
					return // ok
				} else {
					if isUntyped(spec) {
						spec = defaultTypeOf(spec)
					}
					lookup[ct.Generic] = spec
					return // ok
				}
			}
		} else {
			// TODO: handle generics in method signatures
			return // nothing to do
		}
	case *MaybeNativeType:
		switch cbt := baseOf(spec).(type) {
		case *NativeType:
			gnoType := store.Go2GnoType(cbt.Type)
			specifyType(store, lookup, ct.Type, gnoType, nil)
		default:
			specifyType(store, lookup, ct.Type, cbt, nil)
		}
	default:
		// ignore, no generics.
	}
}

// given the lookup map accumulated w/ specifyType(), apply the
// lookup map to derive the specific (composite) type from a
// generic template.  if the input tmpl has no generics, it is
// simply returned.  if a generic is not yet specified, panics.
func applySpecifics(lookup map[Name]Type, tmpl Type) (Type, bool) {
	switch ct := tmpl.(type) {
	case *PointerType:
		pte, ok := applySpecifics(lookup, ct.Elt)
		if !ok { // simply return
			return tmpl, false
		}
		return &PointerType{
			Elt: pte,
		}, true
	case *ArrayType:
		ate, ok := applySpecifics(lookup, ct.Elt)
		if !ok { // simply return
			return tmpl, false
		}
		return &ArrayType{
			Len: ct.Len,
			Elt: ate,
			Vrd: ct.Vrd,
		}, true
	case *SliceType:
		ste, ok := applySpecifics(lookup, ct.Elt)
		if !ok { // simply return
			return tmpl, false
		}
		return &SliceType{
			Elt: ste,
			Vrd: ct.Vrd,
		}, true
	case *MapType:
		mtk, okk := applySpecifics(lookup, ct.Key)
		mtv, okv := applySpecifics(lookup, ct.Value)
		if !okk && !okv { // simply return
			return tmpl, false
		}
		return &MapType{
			Key:   mtk,
			Value: mtv,
		}, true
	case *InterfaceType:
		if ct.Generic != "" {
			if strings.HasSuffix(string(ct.Generic), ".(type)") {
				// used for defining types by name via arg matching.
				return gTypeType, true
			} else {
				// used for capturing and distinguishing native vs gno
				// slice/array types, e.g. for "append".
				// TODO: implement .Elem() on TypeValues.
				generic := ct.Generic
				isElem := strings.HasSuffix(string(ct.Generic), ".Elem()")
				if isElem {
					generic = generic[:len(generic)-len(".Elem()")]
				}
				// Construct BlockStmt from map.
				// TODO: make arg type be this
				// to reduce redundant steps.
				pn := NewPackageNode("", "", nil)
				bs := new(BlockStmt)
				bs.InitStaticBlock(bs, pn)
				for n, t := range lookup {
					bs.Define(n, asValue(t))
				}
				// Parse generic to expr.
				gx := MustParseExpr(string(generic))
				gx = Preprocess(nil, bs, gx).(Expr)
				// Evaluate type from generic expression.
				m := NewMachine("", nil)
				tv := m.EvalStatic(bs, gx)
				m.Release()
				if isElem {
					return tv.GetType().Elem(), true
				} else {
					return tv.GetType(), true
				}
			}
		} else { // simply return
			// TODO: handle generics in method signatures
			return tmpl, false
		}
	default:
		// ignore, no generics.
		return tmpl, false
	}
}

// returns true if t is generic or has generic component.
func isGeneric(t Type) bool {
	switch ct := t.(type) {
	case FieldType:
		return isGeneric(ct.Type)
	case *PointerType:
		return isGeneric(ct.Elt)
	case *ArrayType:
		return isGeneric(ct.Elt)
	case *SliceType:
		return isGeneric(ct.Elt)
	case *MapType:
		return isGeneric(ct.Key) ||
			isGeneric(ct.Value)
	case *InterfaceType:
		// TODO: handle generics in method signatures
		return ct.Generic != ""
	case *MaybeNativeType:
		return isGeneric(ct.Type)
	default:
		return false
	}
}

// NOTE: runs at preprocess time but also runtime,
// for dynamic interface lookups. m can be nil,
// is used for recursion detection.
// TODO: could this be more optimized for the runtime?
// are Go-style itables the solution or?
// callerPath: the path of package where selector node was declared.
func findEmbeddedFieldType(callerPath string, t Type, n Name, m map[Type]struct{}) (
	trail []ValuePath, hasPtr bool, rcvr Type, ft Type, accessError bool,
) {
	switch ct := t.(type) {
	case *DeclaredType:
		return ct.FindEmbeddedFieldType(callerPath, n, m)
	case *PointerType:
		return ct.FindEmbeddedFieldType(callerPath, n, m)
	case *StructType:
		return ct.FindEmbeddedFieldType(callerPath, n, m)
	case *InterfaceType:
		return ct.FindEmbeddedFieldType(callerPath, n, m)
	case *NativeType:
		return ct.FindEmbeddedFieldType(n, m)
	default:
		return nil, false, nil, nil, false
	}
}

func isNative(t Type) bool {
	if _, ok := t.(*NativeType); ok {
		return true
	} else {
		return false
	}
}
