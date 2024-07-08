package tests

import (
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// Creates one much like amino.RegisterPackage, but without registration.
// This is needed due to circular dependency issues for dependencies of Amino.
// Another reason to strive for many independent modules.
// NOTE: Register new repr types here as well.
// NOTE: This package registration is independent of test registration.
// See tests/common.go StructTypes etc to add to tests.
var Package = pkg.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/amino/tests",
	"tests",
	pkg.GetCallersDirname(),
).WithDependencies().WithTypes(
	EmptyStruct{},
	PrimitivesStruct{},
	ShortArraysStruct{},
	ArraysStruct{},
	ArraysArraysStruct{},
	SlicesStruct{},
	SlicesSlicesStruct{},
	PointersStruct{},
	PointerSlicesStruct{},
	// NestedPointersStruct{},
	ComplexSt{},
	EmbeddedSt1{},
	EmbeddedSt2{},
	EmbeddedSt3{},
	EmbeddedSt4{},
	pkg.Type{ // example of overriding type name.
		Type:             reflect.TypeOf(EmbeddedSt5{}),
		Name:             "EmbeddedSt5NameOverride",
		PointerPreferred: false,
	},
	AminoMarshalerStruct1{},
	ReprStruct1{},
	AminoMarshalerStruct2{},
	ReprElem2{},
	AminoMarshalerStruct3{},
	AminoMarshalerInt4(0),
	AminoMarshalerInt5(0),
	AminoMarshalerStruct6{},
	AminoMarshalerStruct7{},
	ReprElem7{},
	IntDef(0),
	IntAr{},
	IntSl(nil),
	ByteAr{},
	ByteSl(nil),
	PrimitivesStructDef{},
	PrimitivesStructSl(nil),
	PrimitivesStructAr{},
	Concrete1{},
	Concrete2{},
	ConcreteTypeDef{},
	ConcreteWrappedBytes{},
	&InterfaceFieldsStruct{},
)
