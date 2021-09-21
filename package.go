package gno

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(

	StringValue(""), "st",
	BigintValue{}, "big",
	// DataByteValue{}
	PointerValue{}, "ptr",
	&ArrayValue{}, "arr",
	&SliceValue{}, "sli",
	&StructValue{}, "str",
	&FuncValue{}, "fun",
	&MapValue{}, "map",
	&BoundMethodValue{}, "bnd",
	TypeValue{}, "typ",
	&PackageValue{}, "pkg",
	// nativeValue
	&Block{}, "blk",
	RefType{}, "rft",
	RefValue{}, "rfv",
))
