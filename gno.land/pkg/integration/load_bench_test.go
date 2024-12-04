package integration

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	tm2Log "github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func BenchmarkTestingNodeInit(b *testing.B) {
	b.StopTimer()

	gnoRootDir := gnoenv.RootDir()
	genesis := &gnoland.GnoGenesisState{
		Balances: LoadDefaultGenesisBalanceFile(b, gnoRootDir),
		Params:   LoadDefaultGenesisParamFile(b, gnoRootDir),
		Txs:      []gnoland.TxWithMetadata{},
	}
	logger := tm2Log.NewNoopLogger()
	pkgs := newPkgsLoader()

	b.StartTimer()

	creator := crypto.MustAddressFromString(DefaultAccount_Address) // test1
	defaultFee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

	// get packages
	pkgsTxs, err := pkgs.LoadPackages(creator, defaultFee, nil)
	if err != nil {
		b.Fatalf("unable to load packages txs: %s", err)
	}

	// Generate config and node
	cfg := TestingMinimalNodeConfig(b, gnoRootDir)
	genesis.Txs = pkgsTxs

	// setup genesis state
	cfg.Genesis.AppState = *genesis

	cfg.DB = memdb.NewMemDB() // so it can be reused when restarting.

	TestingInMemoryNode(b, logger, cfg)
}
