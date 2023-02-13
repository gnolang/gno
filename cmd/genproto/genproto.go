package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/genproto"
	"github.com/gnolang/gno/pkgs/commands"

	// TODO: move these out.
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/blockchain"
	"github.com/gnolang/gno/pkgs/bft/consensus"
	ctypes "github.com/gnolang/gno/pkgs/bft/consensus/types"
	"github.com/gnolang/gno/pkgs/bft/mempool"
	btypes "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/bitarray"
	"github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/crypto/hd"
	"github.com/gnolang/gno/pkgs/crypto/merkle"
	"github.com/gnolang/gno/pkgs/crypto/multisig"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			LongHelp: "Generates proto bindings for Amino packages",
		},
		nil,
		execGen,
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func execGen(_ context.Context, _ []string) error {
	pkgs := []*amino.Package{
		bitarray.Package,
		merkle.Package,
		abci.Package,
		btypes.Package,
		consensus.Package,
		ctypes.Package,
		mempool.Package,
		ed25519.Package,
		blockchain.Package,
		hd.Package,
		multisig.Package,
		std.Package,
		sdk.Package,
		bank.Package,
		vm.Package,
		gno.Package,
	}

	for _, pkg := range pkgs {
		genproto.WriteProto3Schema(pkg)
		genproto.WriteProtoBindings(pkg)
		genproto.MakeProtoFolder(pkg, "proto")
	}

	for _, pkg := range pkgs {
		genproto.RunProtoc(pkg, "proto")
	}

	return nil
}
