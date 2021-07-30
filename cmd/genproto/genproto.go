package main

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/genproto"

	// TODO: move these out.
	// "github.com/gnolang/gno/pkgs/classic/types"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/crypto/merkle"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
	"github.com/gnolang/gno/pkgs/crypto/secp256k1"
)

func main() {
	pkgs := []*amino.Package{
		ed25519.Package,
		secp256k1.Package,
		multisig.Package,
		merkle.Package,
		abci.Package,
		//types.Package,
	}
	for _, pkg := range pkgs {
		genproto.WriteProto3Schema(pkg)
		genproto.WriteProtoBindings(pkg)
		genproto.MakeProtoFolder(pkg, "proto")
	}
	for _, pkg := range pkgs {
		genproto.RunProtoc(pkg, "proto")
	}
}
