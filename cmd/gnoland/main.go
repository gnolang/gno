package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/bft/privval"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
)

func main() {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	rootDir := "testdir"
	cfg := config.LoadOrMakeDefaultConfig(rootDir)

	// create priv validator first.
	// need it to generate genesis.json
	newPrivValKey := cfg.PrivValidatorKeyFile()
	newPrivValState := cfg.PrivValidatorStateFile()
	priv := privval.LoadOrGenFilePV(newPrivValKey, newPrivValState)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	if !osm.FileExists(genesisFilePath) {
		genDoc := makeGenesisDoc(priv.GetPubKey())
		writeGenesisFile(genDoc, genesisFilePath)
	}

	// create application and node.
	gnoApp, err := gnoland.NewApp(rootDir, logger)
	if err != nil {
		panic(fmt.Sprintf("error in creating new app: %v", err))
	}
	cfg.LocalApp = gnoApp
	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		panic(fmt.Sprintf("error in creating node: %v", err))
	}
	if err := gnoNode.Start(); err != nil {
		panic(fmt.Sprintf("error in start node: %v", err))
	}

	// run forever
	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			_ = gnoNode.Stop()
		}
	})
	select {} // run forever
}

// Makes a local test genesis doc with local privValidator.
func makeGenesisDoc(pvPub crypto.PubKey) *bft.GenesisDoc {
	gen := &bft.GenesisDoc{}
	gen.GenesisTime = time.Now()
	gen.ChainID = "testchain"
	gen.Validators = []bft.GenesisValidator{
		bft.GenesisValidator{
			Address: pvPub.Address(),
			PubKey:  pvPub,
			Power:   10,
			Name:    "testvalidator",
		},
	}
	gen.AppState = gnoland.GnoGenesisState{
		Balances: []string{
			"g1luefaaj8sasunh3knlpr37wf0zlccz8n8ev2je=100gnot",
		},
	}
	return gen
}

func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
	err := gen.SaveAs(filePath)
	if err != nil {
		panic(err)
	}
}
