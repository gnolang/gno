package gno

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(

	//----------------------------------------
	// Values
	&TypedValue{}, "tv",
	StringValue(""), "vstr",
	BigintValue{}, "vbig",
	// DataByteValue{}
	PointerValue{}, "vptr",
	&ArrayValue{}, "varr",
	&SliceValue{}, "vsli",
	&StructValue{}, "vstt",
	&FuncValue{}, "vfun",
	&MapValue{}, "vmap",
	&MapList{}, "vmapl",
	&MapListItem{}, "vmapli",
	&BoundMethodValue{}, "vbnd",
	TypeValue{}, "vtyp",
	&PackageValue{}, "vpkg",
	// &NativeValue{},
	&Block{}, "vblk",
	RefValue{}, "vref",

	//----------------------------------------
	// Realm/Object
	&ObjectInfo{}, "obj",

	//----------------------------------------
	// Nodes
	RefNode{}, "nref",
	Location{}, "loc",

	//----------------------------------------
	// Types
	PrimitiveType(0), "tpri",
	&PointerType{}, "tptr",
	&ArrayType{}, "tarr",
	&SliceType{}, "tsli",
	&StructType{}, "tstt",
	FieldType{}, "tfld",
	&FuncType{}, "tfun",
	&MapType{}, "tmap",
	&InterfaceType{}, "tint",
	&TypeType{}, "ttyp",
	&DeclaredType{}, "tdec",
	&PackageType{}, "tpkg",
	&ChanType{}, "tchn",
	blockType{}, "tblk",
	&tupleType{}, "ttup",
	RefType{}, "tref",
))
