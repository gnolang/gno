// Command gnobr (gno block rollback) rolls back a gnoland node to a target
// height and patches the app hash in state.db so gnoland can replay blocks
// locally without any special flags or patches.
//
// Usage:
//
//	gnobr --data-dir gnoland-data --drop-after 352921 --app-hash 14BD8BB9...
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

func main() {
	var (
		dataDir   = flag.String("data-dir", "gnoland-data", "Path to gnoland data directory")
		dropAfter = flag.Int64("drop-after", 0, "Keep up to this height, drop everything after")
		appHash   = flag.String("app-hash", "", "Set this app hash in state.db (hex, required when replaying)")
		dryRun    = flag.Bool("dry-run", false, "Show what would be done without modifying anything")
	)
	flag.Parse()

	if *dropAfter == 0 {
		fmt.Fprintln(os.Stderr, "usage: gnobr --data-dir <path> --drop-after <height> [--app-hash <hex>]")
		os.Exit(1)
	}

	dbDir := filepath.Join(*dataDir, "db")
	targetHeight := *dropAfter

	var newAppHash []byte
	if *appHash != "" {
		h, err := hex.DecodeString(strings.TrimPrefix(*appHash, "0x"))
		if err != nil {
			log.Fatalf("invalid --app-hash: %v", err)
		}
		newAppHash = h
	}

	// 1. Trim blockstore
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
		fmt.Println("[dry-run] would trim blockstore, patch state, wipe app DB")
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

	// 2. Patch state.db: update AppHash so the Handshaker won't panic
	stDB, err := dbm.NewDB("state", dbm.PebbleDBBackend, dbDir)
	if err != nil {
		log.Fatalf("failed to open state.db: %v", err)
	}
	state := sm.LoadState(stDB)
	if state.IsEmpty() {
		fmt.Println("state.db: empty (nothing to patch)")
	} else {
		fmt.Printf("state.db: height=%d appHash=%X\n", state.LastBlockHeight, state.AppHash)
		if newAppHash != nil {
			state.AppHash = newAppHash
			sm.SaveState(stDB, state)
			fmt.Printf("state.db: appHash patched to %X\n", newAppHash)
		}
	}
	stDB.Close()

	// 3. Wipe gnolang.db (app state) → app replays from genesis
	appDBPath := filepath.Join(dbDir, "gnolang.db")
	fmt.Printf("removing %s\n", appDBPath)
	os.RemoveAll(appDBPath)

	// 4. Wipe WAL
	walPath := filepath.Join(*dataDir, "wal")
	fmt.Printf("removing %s\n", walPath)
	os.RemoveAll(walPath)

	// 5. Reset priv_validator_state.json
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
