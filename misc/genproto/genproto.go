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
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "============================================================")
	fmt.Fprintln(os.Stderr, "  WARNING: misc/genproto (genproto1 / pbbindings) is DEPRECATED")
	fmt.Fprintln(os.Stderr, "  for the gno.land project.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Use misc/genproto2 instead. genproto1's pbbindings are")
	fmt.Fprintln(os.Stderr, "  not guaranteed to match the reflect codec byte-for-byte;")
	fmt.Fprintln(os.Stderr, "  genproto2 is the source-of-truth generator with a tested")
	fmt.Fprintln(os.Stderr, "  parity contract against the reflect codec.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  The underlying library `tm2/pkg/amino/genproto` is kept")
	fmt.Fprintln(os.Stderr, "  in-tree for future projects that need protobuf3-compatible")
	fmt.Fprintln(os.Stderr, "  generated code (genproto2 emits its own native wire format,")
	fmt.Fprintln(os.Stderr, "  not standard protobuf3). If you revive genproto1, follow")
	fmt.Fprintln(os.Stderr, "  the procedure in tm2/pkg/amino/genproto/HARDENING.md to")
	fmt.Fprintln(os.Stderr, "  bring it to wire-parity with the reflect codec first.")
	fmt.Fprintln(os.Stderr, "============================================================")
	fmt.Fprintln(os.Stderr, "")

	cmd := commands.NewCommand(
		commands.Metadata{
			LongHelp: "Generates proto bindings for Amino packages.\n\nDEPRECATED for the gno.land project: use misc/genproto2 instead. genproto1's pbbindings are not guaranteed to match the reflect codec byte-for-byte; genproto2 is the source-of-truth generator with a tested parity contract.\n\nThe underlying library tm2/pkg/amino/genproto is kept in-tree for future projects that need protobuf3-compatible generated code (genproto2 emits its own native wire format). If you revive genproto1, see tm2/pkg/amino/genproto/HARDENING.md for the parity-hardening procedure.",
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
