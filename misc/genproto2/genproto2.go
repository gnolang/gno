package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto2"
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
			LongHelp: "Generates pb3 direct-encoding methods for Amino packages",
		},
		commands.NewEmptyConfig(),
		execGen,
	)

	cmd.Execute(context.Background(), os.Args[1:])
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

	cdc := amino.NewCodec()
	for _, pkg := range pkgs {
		cdc.RegisterPackage(pkg)
	}
	cdc.Seal()

	ctx := genproto2.NewP3Context2(cdc)
	for _, pkg := range pkgs {
		ctx.WriteProto3Schema(pkg)
		ctx.WriteProtobuf3(pkg)
	}

	return nil
}
