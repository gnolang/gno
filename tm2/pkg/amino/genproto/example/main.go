package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule2"
)

// amino
type StructA struct {
	FieldC int
	FieldD uint32
}

// amino
type StructB struct {
	FieldC int
	FieldD uint32
	FieldE submodule.StructSM
	FieldF StructA
	FieldG any
}

func main() {
	packages := []*amino.Package{
		submodule2.Package,
		submodule.Package,
		Package,
	}

	for i, pkg := range packages {
		fmt.Println("#", i, "package path", pkg.GoPkgPath)
		// Defined in genproto.go.
		// These will generate .proto files next to
		// their .go origins.
		fmt.Println("#", i, "write proto3 schema")
		genproto.WriteProto3Schema(pkg)

		// Make proto folder for proto dependencies.
		fmt.Println("#", i, "make proto folder")
		genproto.MakeProtoFolder(pkg, "proto")

		// Generate Go code from .proto files generated above.
		fmt.Println("#", i, "run protoc")
		genproto.RunProtoc(pkg, "proto")

		// Generate bindings.go for other methods.
		fmt.Println("#", i, "write proto bindings")
		genproto.WriteProtoBindings(pkg)
	}
}
