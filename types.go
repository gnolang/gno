package gno

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

//----------------------------------------
// (runtime) Type

type Type interface {
	assertType()

	Kind() Kind     // penetrates *DeclaredType & *nativeType
	TypeID() TypeID // deterministic
	String() string // for dev/debugging
	Elem() Type     // for TODO... types
}

type TypeID Hashlet

func (tid TypeID) IsZero() bool {
	return tid == (TypeID{})
}

func (tid TypeID) Bytes() []byte {
	return tid[:]
}

func (tid TypeID) String() string {
	return fmt.Sprintf("%X", tid[:])
}

func typeid(f string, args ...interface{}) (tid TypeID) {
	fs := fmt.Sprintf(f, args...)
	hb := HashBytes([]byte(fs))
	return TypeID(hb)
}

// Complex types are pointers, but due to the design goal
// of the language to enable mass scale persistence, we
// cannot use pointer equality to test for type equality.
// Instead, for checking equalty use the TypeID.
func (PrimitiveType) assertType()  {}
func (PointerType) assertType()    {}
func (FieldType) assertType()      {}
func (*ArrayType) assertType()     {}
func (*SliceType) assertType()     {}
func (*StructType) assertType()    {}
func (*FuncType) assertType()      {}
func (*MapType) assertType()       {}
func (*InterfaceType) assertType() {}
func (*TypeType) assertType()      {}
func (*DeclaredType) assertType()  {}
func (*PackageType) assertType()   {}
func (*ChanType) assertType()      {}
func (*nativeType) assertType()    {}
func (blockType) assertType()      {}

//----------------------------------------
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
	//UintptrType
	UntypedBigintType
	BigintType
)

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
	case BigintType, UntypedBigintType:
		return BigintKind
	default:
		panic(fmt.Sprintf("unexpected primitive type %v", pt))
	}
}

func (pt PrimitiveType) TypeID() TypeID {
	switch pt {
	case InvalidType:
		panic("invalid type has no typeid")
	case UntypedBoolType:
		panic("untyped bool type has no typeid")
	case BoolType:
		return typeid("bool")
	case UntypedStringType:
		panic("untyped string type has no typeid")
	case StringType:
		return typeid("string")
	case IntType:
		return typeid("int")
	case Int8Type:
		return typeid("int8")
	case Int16Type:
		return typeid("int16")
	case UntypedRuneType:
		panic("untyped rune type has no typeid")
	case Int32Type:
		return typeid("int32")
	case Int64Type:
		return typeid("int64")
	case UintType:
		return typeid("uint")
	case Uint8Type:
		return typeid("uint8")
	case DataByteType:
		panic("untyped data byte type has no typeid")
	case Uint16Type:
		return typeid("uint16")
	case Uint32Type:
		return typeid("uint32")
	case Uint64Type:
		return typeid("uint64")
	case UntypedBigintType:
		panic("untyped bigint type has no typeid")
	case BigintType:
		return typeid("bigint")
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
	case UntypedRuneType, Int32Type:
		return string("int32")
	case Int64Type:
		return string("int64")
	case UintType:
		return string("uint")
	case Uint8Type, DataByteType:
		return string("uint8")
	case Uint16Type:
		return string("uint16")
	case Uint32Type:
		return string("uint32")
	case Uint64Type:
		return string("uint64")
	case UntypedBigintType:
		return string("<untyped> bigint")
	case BigintType:
		return string("bigint")
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

//----------------------------------------
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
	panic("see FieldTypeList.TypeID()")
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

//----------------------------------------
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
			s += string(ft.Name) + "#" + ft.Type.TypeID().String()
		}
		if i != ll-1 {
			s += ","
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
			s += string(fn) + "#" + ft.Type.TypeID().String()
		} else {
			s += pkgPath + "." + string(fn) + "#" + ft.Type.TypeID().String()
		}
		if i != ll-1 {
			s += ","
		}
	}
	return typeid(s)
}

func (l FieldTypeList) HasUnexported() bool {
	for _, ft := range l {
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
		s += string(ft.Name) + "#" + ft.Type.TypeID().String()
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
		s += string(ft.Name) + "#" + ft.Type.String()
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
			s += ","
		}
	}
	return typeid(s)
}

//----------------------------------------
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

//----------------------------------------
// Slice type

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

//----------------------------------------
// Pointer type

type PointerType struct {
	Elt Type

	typeid TypeID
}

func (pt PointerType) Kind() Kind {
	return PointerKind
}

func (pt PointerType) TypeID() TypeID {
	if pt.typeid.IsZero() {
		pt.typeid = typeid("*%s", pt.Elt.TypeID().String())
	}
	return pt.typeid
}

func (pt PointerType) String() string {
	return fmt.Sprintf("*%s", pt.Elt.String())
}

func (pt PointerType) Elem() Type {
	return pt.Elt
}

//----------------------------------------
// Struct type

// Struct fields are flattened.
// All non-pointer (embedded and named) inner struct fields get
// appended (flattened) to the outer struct's fields buffer.
// Each non-pointer inner-struct's fields are preceded by the
// type of inner-struct.
//
// Mapping contains the original field index to the translated
// index in Fields. The value is always greater than or equal
// to the key.
//
// Example:
// type Foo struct {
//   A int
//   *Foo
// }
// type Bar struct {
//   B int
//   X Foo
//   C int
// }
// StructType{Bar}.Fields = []FieldType{
//   {"B", IntType},
//   {"X", StructType{Foo}},
//   {"A", IntType},
//   {"Foo", PointerType{StructType{Foo}}},
//   {"C", IntType},
// }
// StructType{Bar}.Mapping = []int{0, 1, 4}
//
// The type of non-pointer inner struct fields have as their
// fields slices of the container struct type's field buffer.
// StructValues are similar in structure.  Mapping
// contains a mapping from the top-level declared field
// index to the corresponding entry in the flat buffer
// Fields.  i.e., len(Mapping) <= len(Fields) and Mapping[x] >= x.
type StructType struct {
	PkgPath string
	Fields  []FieldType // flattened
	Mapping []int       // map[Orig]:Flat

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
			"s{%s}",
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

func (st *StructType) GetPathForName(n Name) ValuePath {
	for i := 0; i < len(st.Fields); i++ {
		ft := st.Fields[i]
		if ft.Name == n {
			if i > 2<<16-1 {
				panic("too many fields")
			}
			return NewValuePathDefault(1, uint16(i), n)
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
		if path.Depth != 1 {
			panic("expected path.Depth of 1")
		}
	}
	return st.Fields[path.Index].Type
}

//----------------------------------------
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
		pt.typeid = typeid("pkg{}")
	}
	return pt.typeid
}

func (pt *PackageType) String() string {
	return "package{}"
}

func (pt *PackageType) Elem() Type {
	panic("package types have no elements")
}

//----------------------------------------
// Interface type

type InterfaceType struct {
	PkgPath   string
	Methods   []FieldType
	IsUntyped bool // for uverse "generics"

	typeid TypeID
}

// General empty interface.
var gEmptyInterfaceType *InterfaceType = &InterfaceType{}

// Special untyped type for uverse functions.
var gAnyInterfaceType *InterfaceType = &InterfaceType{IsUntyped: true}

func (it *InterfaceType) IsEmptyInterface() bool {
	return len(it.Methods) == 0
}

func (it *InterfaceType) Kind() Kind {
	return InterfaceKind
}

func (it *InterfaceType) TypeID() TypeID {
	if debug {
		if it.IsUntyped {
			panic("untyped interface type has no TypeID")
		}
	}
	if it.typeid.IsZero() {
		// NOTE Interface types expressed or declared in different
		// packages may have the same TypeID if and only if neither
		// have unexported fields.  pt.Path is only included in field
		// names that are not uppercase.
		ms := FieldTypeList(it.Methods)
		// XXX pre-sort.
		sort.Sort(ms)
		it.typeid = typeid("i{" + ms.TypeIDForPackage(it.PkgPath).String() + "}")
	}
	return it.typeid
}

func (it *InterfaceType) String() string {
	if it.IsUntyped {
		return fmt.Sprintf("any{%s}",
			FieldTypeList(it.Methods).String())
	} else {
		return fmt.Sprintf("interface{%s}",
			FieldTypeList(it.Methods).String())
	}
}

func (it *InterfaceType) Elem() Type {
	panic("interface types have no elements")
}

// TODO: optimize
func (it *InterfaceType) GetMethodType(n Name) *FuncType {
	for _, im := range it.Methods {
		if im.Name == n {
			return im.Type.(*FuncType)
		}
	}
	return nil
}

// For run-time type assertion.
// TODO: optimize somehow.
// NOTE: also see *DeclaredType.Implements.
func (it *InterfaceType) Implements(ot *InterfaceType) bool {
	for _, om := range ot.Methods {
		if imt := it.GetMethodType(om.Name); imt != nil {
			imtid := imt.TypeID()
			omtid := om.Type.TypeID()
			if imtid != omtid {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (it *InterfaceType) GetPathForName(n Name) ValuePath {
	return NewValuePathInterface(n)
}

//----------------------------------------
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
	panic("not yet implemented")
}

func (ct *ChanType) String() string {
	panic("not yet implemented")
}

func (ct *ChanType) Elem() Type {
	return ct.Elt
}

//----------------------------------------
// Function type

type FuncType struct {
	PkgPath string // needed for realm enforcement.
	Params  []FieldType
	Results []FieldType

	typeid TypeID
	bound  *FuncType
}

func (ft *FuncType) Kind() Kind {
	return FuncKind
}

// bound function type (if ft is a method).
func (ft *FuncType) BoundType() *FuncType {
	if ft.bound == nil {
		ft.bound = &FuncType{
			PkgPath: ft.PkgPath,
			Params:  ft.Params[1:],
			Results: ft.Results,
		}
	}
	return ft.bound
}

// unbound function type
func (ft *FuncType) UnboundType(rft FieldType) *FuncType {
	return &FuncType{
		PkgPath: ft.PkgPath,
		Params:  append([]FieldType{rft}, ft.Params...),
		Results: ft.Results,
	}
}

func (ft *FuncType) TypeID() TypeID {
	// Two functions of different realms can have the same
	// type, because the method signature doesn't change, and
	// this exchangeability is useful to denote type semantics.
	ps := FieldTypeList(ft.Params)
	rs := FieldTypeList(ft.Results)
	pp := ""
	if ps.HasUnexported() || rs.HasUnexported() {
		pp = fmt.Sprintf("@%q", ft.PkgPath)
	}
	if ft.typeid.IsZero() {
		ft.typeid = typeid(
			"f%s(%s)(%s)",
			pp,
			ps.UnnamedTypeID(),
			rs.UnnamedTypeID(),
		)
	}
	return ft.typeid
}

func (ft *FuncType) String() string {
	return fmt.Sprintf("func@%s(%s)(%s)",
		ft.PkgPath,
		FieldTypeList(ft.Params).StringWithCommas(),
		FieldTypeList(ft.Results).StringWithCommas())
}

func (ft *FuncType) Elem() Type {
	panic("function types have no elements")
}

func (ft *FuncType) HasVarg() bool {
	if numParams := len(ft.Params); numParams == 0 {
		return false
	} else {
		lpt := ft.Params[numParams-1].Type
		if lat, ok := lpt.(*SliceType); ok {
			return lat.Vrd
		} else {
			return false
		}
	}
}

//----------------------------------------
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
			"m[%s]%s",
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

//----------------------------------------
// Type (typeval) type

type TypeType struct {
	// nothing yet.
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

//----------------------------------------
// Declared type
// Declared types have a name, base (underlying) type,
// and associated methods.

type DeclaredType struct {
	PkgPath string
	Name    Name
	Base    Type         // not a DeclaredType
	Methods []TypedValue // {T:*FuncType,V:*FuncValue}...

	typeid TypeID
}

var x = 0

func declareWith(pkgPath string, name Name, b Type) *DeclaredType {
	dt := &DeclaredType{
		PkgPath: pkgPath,
		Name:    name,
		Base:    baseOf(b),
	}
	x++
	if x == 3 {
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

func (dt *DeclaredType) TypeID() TypeID {
	if dt.typeid.IsZero() {
		dt.typeid = typeid("%s.%s=%s",
			dt.PkgPath, dt.Name, dt.Base.TypeID().String())
	}
	return dt.typeid
}

func (dt *DeclaredType) String() string {
	return fmt.Sprintf("%s.%s", dt.PkgPath, dt.Name)
}

func (dt *DeclaredType) Elem() Type {
	return dt.Base.Elem()
}

func (dt *DeclaredType) GetPathForName(n Name) ValuePath {
	// May be a method.
	for i, tv := range dt.Methods {
		fv := tv.V.(*FuncValue)
		if fv.Name == n {
			if i > 2<<16-1 {
				panic("too many methods")
			}
			return NewValuePathDefault(1, uint16(i), n)
		}
	}
	// Otherwise it is underlying.
	path := dt.Base.(ValuePather).GetPathForName(n)
	path.Depth += 1
	return path
}

func (dt *DeclaredType) GetValueRefAt(path ValuePath) *TypedValue {
	if path.Type == VPTypeInterface {
		mv := dt.GetValueRef(path.Name)
		return mv
	} else if path.Type == VPTypeDefault {
		if path.Depth == 0 {
			panic("*DeclaredType global fields not yet implemented")
		} else if path.Depth == 1 {
			return &dt.Methods[path.Index]
		} else {
			panic("DeclaredType.GetValueRefAt() expects generation <= 1")
		}
	} else {
		panic(fmt.Sprintf(
			"unexpected value path type %X",
			path.Type))
	}
}

// TODO: optimize
func (dt *DeclaredType) GetValueRef(n Name) *TypedValue {
	for i := 0; i < len(dt.Methods); i++ {
		mv := &dt.Methods[i]
		if fv := mv.V.(*FuncValue); fv.Name == n {
			return mv
		}
	}
	return nil
}

func (dt *DeclaredType) GetMethod(n Name) *FuncValue {
	mv := dt.GetValueRef(n)
	if mv != nil {
		return mv.GetFunc()
	} else {
		return nil
	}
}

func (dt *DeclaredType) DefineMethod(fv *FuncValue) {
	dt.Methods = append(dt.Methods, TypedValue{
		T: fv.Type,
		V: fv,
	})
}

// For run-time type assertion.
// TODO: optimize somehow.
// NOTE: also see *InterfaceType.Implements.
func (dt *DeclaredType) Implements(ot *InterfaceType) bool {
	for _, om := range ot.Methods {
		if dm := dt.GetMethod(om.Name); dm != nil {
			dmtid := dm.Type.BoundType().TypeID()
			omtid := om.Type.TypeID()
			if dmtid != omtid {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

//----------------------------------------
// Native type

type nativeType struct {
	Type reflect.Type // Go "native" type

	typeid  TypeID
	gnoType Type // Gno converted type
}

func (nt *nativeType) Kind() Kind {
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

func (nt *nativeType) TypeID() TypeID {
	// like a DeclaredType, but different.
	if nt.typeid.IsZero() {
		nt.typeid = typeid("go:%s.%s", nt.Type.PkgPath(), nt.Type.Name())
	}
	return nt.typeid
}

func (nt *nativeType) String() string {
	return fmt.Sprintf("gonative{%s}", nt.Type.String())
}

func (nt *nativeType) Elem() Type {
	return nt.GnoType().Elem()
}

func (nt *nativeType) GnoType() Type {
	if nt.gnoType == nil {
		nt.gnoType = go2GnoType2(nt.Type)
	}
	return nt.gnoType
}

//----------------------------------------
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

//----------------------------------------
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
	BigintKind // not in go.
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
	BlockKind // not in go.
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
		case BigintType, UntypedBigintType:
			return BigintKind
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
	case PointerType:
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
	case *nativeType:
		return t.Kind()
	case blockType:
		return BlockKind
	default:
		panic(fmt.Sprintf("unexpected type %#v", t))
	}
}

//----------------------------------------
// misc

func isUntyped(t Type) bool {
	switch t {
	case UntypedBoolType, UntypedRuneType, UntypedBigintType, UntypedStringType:
		return true
	default:
		return false
	}
}

// TODO move untyped const stuff to preprocess.go
func defaultTypeOf(t Type) Type {
	switch t {
	case UntypedBoolType:
		return BoolType
	case UntypedRuneType:
		return Int32Type
	case UntypedBigintType:
		return IntType
	case UntypedStringType:
		return StringType
	default:
		if debug {
			panic(fmt.Sprintf("unexpected type for default untyped const conversion: %s", t.String()))
		} else {
			panic("unexpected type for default untyped const conversion")
		}
	}
}

func fillEmbeddedName(ft *FieldType) {
	if ft.Name != "" {
		return
	}
	switch ct := ft.Type.(type) {
	case PointerType:
		// dereference one level
		switch ct := ct.Elt.(type) {
		case *DeclaredType:
			ft.Name = ct.Name
			return
		case *nativeType:
			panic("native type cannot be embedded")
		default:
			panic("should not happen")
		}
	case *DeclaredType:
		ft.Name = ct.Name
		return
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
		case BigintType:
			ft.Name = Name("bigint")
		default:
			panic("should not happen")
		}
	case *nativeType:
		panic("native type cannot be embedded")
	default:
		panic(fmt.Sprintf(
			"unexpected field type %s",
			ft.Type.String()))
	}
	ft.Embedded = true
}
