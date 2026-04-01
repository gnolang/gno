// Command gnobr (gno block rollback) trims the block store to a target height
// and wipes state/app DBs so gnoland replays all blocks locally on restart.
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

	bsDB, err := dbm.NewDB("blockstore", dbm.PebbleDBBackend, dbDir)
	if err != nil {
		log.Fatalf("failed to open blockstore.db: %v", err)
	}

	currentHeight := loadBlockStoreState(bsDB).Height
	fmt.Printf("blockstore height: %d\n", currentHeight)

	targetHeight := *dropAfter
	if targetHeight <= 0 {
		log.Fatalf("target height %d is invalid", targetHeight)
	}
	if targetHeight >= currentHeight {
		fmt.Printf("target height %d >= current height %d, nothing to do\n", targetHeight, currentHeight)
		bsDB.Close()
		return
	}

	fmt.Printf("target height: %d (dropping blocks %d..%d)\n", targetHeight, targetHeight+1, currentHeight)

	if *dryRun {
		fmt.Println("[dry-run] would delete blocks and wipe state.db + gnolang.db")
		bsDB.Close()
		return
	}

	for h := targetHeight + 1; h <= currentHeight; h++ {
		if h%10000 == 0 {
			fmt.Printf("  deleting block %d...\n", h)
		}
		deleteBlock(bsDB, h)
	}

	saveBlockStoreState(bsDB, targetHeight)
	fmt.Printf("blockstore height set to %d\n", targetHeight)
	bsDB.Close()

	// Wipe state.db and gnolang.db so the app replays from genesis through blockstore
	for _, name := range []string{"state.db", "gnolang.db"} {
		p := filepath.Join(dbDir, name)
		fmt.Printf("removing %s\n", p)
		os.RemoveAll(p)
	}

	walPath := filepath.Join(*dataDir, "wal")
	fmt.Printf("removing %s\n", walPath)
	os.RemoveAll(walPath)

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
	bss := blockStoreState{Height: height}
	buf, err := amino.MarshalJSON(bss)
	if err != nil {
		log.Fatalf("failed to marshal blockstore state: %v", err)
	}
	db.SetSync([]byte("blockStore"), buf)
}
