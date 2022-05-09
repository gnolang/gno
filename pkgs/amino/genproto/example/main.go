package main

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/genproto"
	"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule"
	"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule2"
)

// amino
type StructA struct {
	fieldA int
	fieldB int
	FieldC int
	FieldD uint32
}

// amino
type StructB struct {
	fieldA int
	fieldB int
	FieldC int
	FieldD uint32
	FieldE submodule.StructSM
	FieldF StructA
	FieldG interface{}
}

func main() {
	packages := []*amino.Package{
		Package,
		submodule.Package,
		submodule2.Package,
	}

	for _, pkg := range packages {
		// Defined in genproto.go.
		// These will generate .proto files next to
		// their .go origins.
		genproto.WriteProto3Schema(pkg)

		// Make proto folder for proto dependencies.
		genproto.MakeProtoFolder(pkg, "proto")

		// Generate Go code from .proto files generated above.
		genproto.RunProtoc(pkg, "proto")

		// Generate bindings.go for other methods.
		genproto.WriteProtoBindings(pkg)
	}
}
