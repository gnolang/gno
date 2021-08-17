package gno

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	RefImage{}, "RefImg",
	PrimitiveValueImage{}, "PriImg",
	PointerValueImage{}, "PtrImg",
	ArrayValueImage{}, "ArrImg",
	SliceValueImage{}, "SliImg",
	StructValueImage{}, "StrImg",
	FuncValueImage{}, "FunImg",
	BoundMethodValueImage{}, "BndImg",
	MapValueImage{}, "MapImg",
	TypeValueImage{}, "TypImg",
	PackageValueImage{}, "PkgImg",
	BlockValueImage{}, "BlkImg",
))
