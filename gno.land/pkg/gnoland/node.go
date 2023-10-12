package gnoland

// func loadGenesisTxs(
// 	path string,
// 	chainID string,
// 	genesisRemote string,
// ) []std.Tx {
// 	txs := []std.Tx{}
// 	txsBz := osm.MustReadFile(path)
// 	txsLines := strings.Split(string(txsBz), "\n")
// 	for _, txLine := range txsLines {
// 		if txLine == "" {
// 			continue // skip empty line
// 		}

// 		// patch the TX
// 		txLine = strings.ReplaceAll(txLine, "%%CHAINID%%", chainID)
// 		txLine = strings.ReplaceAll(txLine, "%%REMOTE%%", genesisRemote)

// 		var tx std.Tx
// 		amino.MustUnmarshalJSON([]byte(txLine), &tx)
// 		txs = append(txs, tx)
// 	}

// 	return txs
// }

// func setupTestingGenesis(gnoDataDir string, cfg *config.Config, icfg *IntegrationConfig) error {
// 	genesisFilePath := filepath.Join(gnoDataDir, cfg.Genesis)
// 	osm.EnsureDir(filepath.Dir(genesisFilePath), 0o700)
// 	if !osm.FileExists(genesisFilePath) {
// 		genesisTxs := loadGenesisTxs(icfg.GenesisTxsFile, icfg.ChainID, icfg.GenesisRemote)
// 		pvPub := priv.GetPubKey()

// 		gen := &bft.GenesisDoc{
// 			GenesisTime: time.Now(),
// 			ChainID:     icfg.ChainID,
// 			ConsensusParams: abci.ConsensusParams{
// 				Block: &abci.BlockParams{
// 					// TODO: update limits.
// 					MaxTxBytes:   1000000,  // 1MB,
// 					MaxDataBytes: 2000000,  // 2MB,
// 					MaxGas:       10000000, // 10M gas
// 					TimeIotaMS:   100,      // 100ms
// 				},
// 			},
// 			Validators: []bft.GenesisValidator{
// 				{
// 					Address: pvPub.Address(),
// 					PubKey:  pvPub,
// 					Power:   10,
// 					Name:    "testvalidator",
// 				},
// 			},
// 		}

// 		// Load distribution.
// 		balances := loadGenesisBalances(icfg.GenesisBalancesFile)

// 		// Load initial packages from examples.
// 		// XXX: we should be able to config this
// 		test1 := crypto.MustAddressFromString(test1Addr)
// 		txs := []std.Tx{}

// 		// List initial packages to load from examples.
// 		// println(filepath.Join(gnoRootDir, "examples"))

// 		// load genesis txs from file.
// 		txs = append(txs, genesisTxs...)

// 		// construct genesis AppState.
// 		gen.AppState = GnoGenesisState{
// 			Balances: balances,
// 			Txs:      txs,
// 		}

// 		writeGenesisFile(gen, genesisFilePath)
// 	}

// 	return nil
// }

// func loadGenesisBalances(path string) []string {
// 	// each balance is in the form: g1xxxxxxxxxxxxxxxx=100000ugnot
// 	balances := []string{}
// 	content := osm.MustReadFile(path)
// 	lines := strings.Split(string(content), "\n")
// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)

// 		// remove comments.
// 		line = strings.Split(line, "#")[0]
// 		line = strings.TrimSpace(line)

// 		// skip empty lines.
// 		if line == "" {
// 			continue
// 		}

// 		parts := strings.Split(line, "=")
// 		if len(parts) != 2 {
// 			panic("invalid genesis_balance line: " + line)
// 		}

// 		balances = append(balances, line)
// 	}
// 	return balances
// }

// func writeGenesisFile(gen *bft.GenesisDoc, filePath string) {
// 	err := gen.SaveAs(filePath)
// 	if err != nil {
// 		panic(err)
// 	}
// }
