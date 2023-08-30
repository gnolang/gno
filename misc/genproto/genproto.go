package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"

	// TODO: move these out.
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/blockchain"
	"github.com/gnolang/gno/tm2/pkg/bft/consensus"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			LongHelp: "Generates proto bindings for Amino packages",
		},
		commands.NewEmptyConfig(),
		execGen,
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

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
		tests.Package,
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
