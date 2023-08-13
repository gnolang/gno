package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Makes a local test genesis doc with local privValidator.
func makeGenesisDoc(
	pvPub crypto.PubKey,
	chainID string,
	genesisBalancesFile string,
	genesisTxs []std.Tx,
) *bft.GenesisDoc {
	gen := &bft.GenesisDoc{}

	gen.GenesisTime = time.Now()
	gen.ChainID = chainID
	gen.ConsensusParams = abci.ConsensusParams{
		Block: &abci.BlockParams{
			// TODO: update limits.
			MaxTxBytes:   1000000,  // 1MB,
			MaxDataBytes: 2000000,  // 2MB,
			MaxGas:       10000000, // 10M gas
			TimeIotaMS:   100,      // 100ms
		},
	}
	gen.Validators = []bft.GenesisValidator{
		{
			Address: pvPub.Address(),
			PubKey:  pvPub,
			Power:   10,
			Name:    "testvalidator",
		},
	}

	// Load distribution.
	balances := loadGenesisBalances(genesisBalancesFile)
	// debug: for _, balance := range balances { fmt.Println(balance) }

	// Load initial packages from examples.
	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	txs := []std.Tx{}

	// List initial packages to load from examples.
	pkgs, err := gnomod.ListPkgs(filepath.Join("..", "examples"))
	if err != nil {
		panic(fmt.Errorf("listing gno packages: %w", err))
	}

	// Sort packages by dependencies.
	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		panic(fmt.Errorf("sorting packages: %w", err))
	}

	// Filter out draft packages.
	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()

	for _, pkg := range nonDraftPkgs {
		// open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)

		var tx std.Tx
		tx.Msgs = []std.Msg{
			vmm.MsgAddPackage{
				Creator: test1,
				Package: memPkg,
				Deposit: nil,
			},
		}
		tx.Fee = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	// load genesis txs from file.
	txs = append(txs, genesisTxs...)

	// construct genesis AppState.
	gen.AppState = gnoland.GnoGenesisState{
		Balances: balances,
		Txs:      txs,
	}
	return gen
}

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}

func loadGenesisTxs(
	path string,
	chainID string,
	genesisRemote string,
) []std.Tx {
	txs := []std.Tx{}
	txsBz := osm.MustReadFile(path)
	txsLines := strings.Split(string(txsBz), "\n")
	for _, txLine := range txsLines {
		if txLine == "" {
			continue // skip empty line
		}

		// patch the TX
		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

		var tx std.Tx
		amino.MustUnmarshalJSON([]byte(txLine), &tx)
		txs = append(txs, tx)
	}

	return txs
}

func loadGenesisBalances(path string) []string {
	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
	balances := []string{}
	content := osm.MustReadFile(path)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// remove comments.
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		// skip empty lines.
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			panic("invalid genesis_balance line: " + line)
		}

		balances = append(balances, line)
	}
	return balances
}
