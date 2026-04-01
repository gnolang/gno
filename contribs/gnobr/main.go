// Command gnobr (gno block rollback) rolls back a gnoland node to a target
// height by trimming the blockstore and wiping app state. On restart, gnoland's
// Handshaker replays all blocks from genesis through the local blockstore.
// No network access needed.
//
// Usage:
//
//	gnobr --data-dir /path/to/gnoland-data --drop-after 352921
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/amino"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

func main() {
	var (
		dataDir   = flag.String("data-dir", "gnoland-data", "Path to gnoland data directory")
		dropAfter = flag.Int64("drop-after", 0, "Keep up to this height, drop everything after")
		dryRun    = flag.Bool("dry-run", false, "Show what would be done without modifying anything")
	)
	flag.Parse()

	if *dropAfter == 0 {
		fmt.Fprintln(os.Stderr, "usage: gnobr --data-dir <path> --drop-after <height>")
		os.Exit(1)
	}

	dbDir := filepath.Join(*dataDir, "db")
	targetHeight := *dropAfter

	// 1. Trim blockstore to target height
	bsDB, err := dbm.NewDB("blockstore", dbm.PebbleDBBackend, dbDir)
	if err != nil {
		log.Fatalf("failed to open blockstore.db: %v", err)
	}
	bsHeight := loadBlockStoreState(bsDB).Height
	fmt.Printf("blockstore: %d\n", bsHeight)

	if targetHeight >= bsHeight {
		fmt.Printf("target %d >= blockstore %d, nothing to trim\n", targetHeight, bsHeight)
		bsDB.Close()
		return
	}

	fmt.Printf("target: %d (dropping blocks %d..%d)\n", targetHeight, targetHeight+1, bsHeight)

	if *dryRun {
		fmt.Println("[dry-run] would trim blockstore, wipe app state, reset validator state")
		bsDB.Close()
		return
	}

	for h := targetHeight + 1; h <= bsHeight; h++ {
		if h%10000 == 0 {
			fmt.Printf("  deleting block %d...\n", h)
		}
		deleteBlock(bsDB, h)
	}
	saveBlockStoreState(bsDB, targetHeight)
	bsDB.Close()
	fmt.Printf("blockstore trimmed to %d\n", targetHeight)

	// 2. Wipe gnolang.db (app state) — app reports height 0 on startup.
	//    state.db is kept intact — Handshaker needs storeHeight == stateHeight.
	//    With app=0, store=N, state=N, the Handshaker replays blocks 1..N.
	appDBPath := filepath.Join(dbDir, "gnolang.db")
	fmt.Printf("removing %s\n", appDBPath)
	os.RemoveAll(appDBPath)

	// 3. Wipe WAL
	walPath := filepath.Join(*dataDir, "wal")
	fmt.Printf("removing %s\n", walPath)
	os.RemoveAll(walPath)

	// 4. Reset priv_validator_state.json
	pvsPath := filepath.Join(*dataDir, "secrets", "priv_validator_state.json")
	pvs := map[string]interface{}{"height": "0", "round": "0", "step": 0}
	pvsBytes, _ := json.MarshalIndent(pvs, "", "  ")
	if err := os.WriteFile(pvsPath, pvsBytes, 0o644); err != nil {
		log.Printf("warning: could not reset %s: %v", pvsPath, err)
	} else {
		fmt.Printf("reset %s\n", pvsPath)
	}

	fmt.Printf("\ndone. restart gnoland — it will replay blocks 1..%d from local blockstore.\n", targetHeight)
}

func deleteBlock(db dbm.DB, height int64) {
	db.Delete([]byte(fmt.Sprintf("H:%v", height)))
	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("P:%v:%v", height, i))
		if has, _ := db.Has(key); !has {
			break
		}
		db.Delete(key)
	}
	db.Delete([]byte(fmt.Sprintf("C:%v", height)))
	db.Delete([]byte(fmt.Sprintf("SC:%v", height)))
}

type blockStoreState struct {
	Height int64 `json:"height"`
}

func loadBlockStoreState(db dbm.DB) blockStoreState {
	var bss blockStoreState
	buf, err := db.Get([]byte("blockStore"))
	if err != nil || len(buf) == 0 {
		return bss
	}
	amino.MustUnmarshalJSON(buf, &bss)
	return bss
}

func saveBlockStoreState(db dbm.DB, height int64) {
	buf, err := amino.MarshalJSON(blockStoreState{Height: height})
	if err != nil {
		log.Fatalf("failed to marshal blockstore state: %v", err)
	}
	db.SetSync([]byte("blockStore"), buf)
}
