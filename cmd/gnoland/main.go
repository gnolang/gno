package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
)

func main() {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	cfg := config.DefaultConfig()
	rootDir := "testdir"
	config.EnsureRoot(rootDir)
	cfg.SetRootDir(rootDir)

	// write genesis file if missing.
	genesisFilePath := filepath.Join(rootDir, cfg.Genesis)
	if !osm.FileExists(genesisFilePath) {
		writeGenesisFile(genesisFilePath)
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

func writeGenesisFile(filePath string) {
	osm.MustWriteFile(filePath, []byte(genesisJSON), 0644)
}

const genesisJSON = `{
  "genesis_time": "2018-10-10T08:20:13.695936996Z",
  "chain_id": "%s",
  "validators": [
    {
      "pub_key": {
        "@type": "/tm.PubKeyEd25519",
        "value":"AT/+aaL1eB0477Mud9JMm8Sh8BIvOYlPGC9KkIUmFaE="
      },
      "power": "10",
      "name": ""
    }
  ],
  "app_hash": "",
  "app_state": {
    "@type": "/gno.GenesisState",
    "balances": [
	  "g1luefaaj8sasunh3knlpr37wf0zlccz8n8ev2je=100gnot"
	]
  }
}`
