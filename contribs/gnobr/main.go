// Command gnobr (gno block rollback) rolls back a gnoland node to a target
// height and patches state.db so gnoland can replay blocks locally on restart.
//
// Usage:
//
//	gnobr --data-dir gnoland-data --drop-after 352921 [--app-hash 311BB985...]
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
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

func main() {
	var (
		dataDir   = flag.String("data-dir", "gnoland-data", "Path to gnoland data directory")
		dropAfter = flag.Int64("drop-after", 0, "Keep up to this height, drop everything after")
		appHash   = flag.String("app-hash", "", "Override app hash in state.db (hex); auto-detected from block header if omitted")
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

	// 1. Open blockstore
	bsDB, err := dbm.NewDB("blockstore", dbm.PebbleDBBackend, dbDir)
	if err != nil {
		log.Fatalf("failed to open blockstore.db: %v", err)
	}
	bs := store.NewBlockStore(bsDB)
	bsHeight := bs.Height()
	fmt.Printf("blockstore: %d\n", bsHeight)

	if targetHeight > bsHeight {
		fmt.Printf("target %d > blockstore %d, nothing to do\n", targetHeight, bsHeight)
		bsDB.Close()
		return
	}

	if *dryRun {
		fmt.Printf("target: %d\n[dry-run] would trim blockstore, patch state, wipe app DB\n", targetHeight)
		bsDB.Close()
		return
	}

	// Read block meta for target height and target+1 (needed to patch state)
	targetMeta := bs.LoadBlockMeta(targetHeight)
	if targetMeta == nil {
		log.Fatalf("block meta for height %d not found", targetHeight)
	}
	// Block at target+1 has LastResultsHash for target in its header.
	// Must read before trimming since trim deletes blocks above target.
	nextMeta := bs.LoadBlockMeta(targetHeight + 1)

	// Trim blocks above target
	if targetHeight < bsHeight {
		fmt.Printf("trimming blocks %d..%d\n", targetHeight+1, bsHeight)
		for h := targetHeight + 1; h <= bsHeight; h++ {
			if h%10000 == 0 {
				fmt.Printf("  deleting block %d...\n", h)
			}
			deleteBlock(bsDB, h)
		}
		saveBlockStoreState(bsDB, targetHeight)
		fmt.Printf("blockstore trimmed to %d\n", targetHeight)
	} else {
		fmt.Println("blockstore already at target, no trimming needed")
	}
	bsDB.Close()

	// 2. Patch state.db
	stDB, err := dbm.NewDB("state", dbm.PebbleDBBackend, dbDir)
	if err != nil {
		log.Fatalf("failed to open state.db: %v", err)
	}
	state := sm.LoadState(stDB)
	if state.IsEmpty() {
		fmt.Println("state.db: empty (nothing to patch)")
	} else {
		fmt.Printf("state.db: height=%d appHash=%X\n", state.LastBlockHeight, state.AppHash)

		if state.LastBlockHeight >= targetHeight {
			state.LastBlockHeight = targetHeight
			state.LastBlockID = targetMeta.BlockID
			state.LastBlockTime = targetMeta.Header.Time
			state.LastBlockTotalTx = targetMeta.Header.TotalTxs
		}

		// Get LastResultsHash from block at targetHeight+1 (its header stores the
		// results hash for targetHeight). This avoids loading ABCI responses which
		// calls osm.Exit() if amino can't unmarshal chain-specific event types.
		if nextMeta != nil {
			state.LastResultsHash = nextMeta.Header.LastResultsHash
			fmt.Printf("state.db: LastResultsHash set from block %d header\n", targetHeight+1)
		} else {
			fmt.Printf("state.db: WARNING: block %d not available, LastResultsHash not updated\n", targetHeight+1)
		}

		if appHash := resolveAppHash(newAppHash, nextMeta); appHash != nil {
			state.AppHash = appHash
			if newAppHash == nil {
				fmt.Printf("state.db: AppHash auto-detected from block %d header: %X\n", targetHeight+1, appHash)
			}
		} else if newAppHash == nil {
			fmt.Printf("state.db: WARNING: AppHash not updated (block %d not found; use --app-hash to set manually)\n", targetHeight+1)
		}

		sm.SaveState(stDB, state)
		fmt.Printf("state.db: patched height=%d blockID=%v time=%v totalTx=%d appHash=%X\n",
			state.LastBlockHeight, state.LastBlockID, state.LastBlockTime,
			state.LastBlockTotalTx, state.AppHash)
	}
	stDB.Close()

	// 3. Wipe gnolang.db (app state)
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

func saveBlockStoreState(db dbm.DB, height int64) {
	type bss struct {
		Height int64 `json:"height"`
	}
	buf, err := amino.MarshalJSON(bss{Height: height})
	if err != nil {
		log.Fatalf("failed to marshal blockstore state: %v", err)
	}
	db.SetSync([]byte("blockStore"), buf)
}

// resolveAppHash returns the app hash to set in state.db.
// If explicit is non-nil (from --app-hash), it takes precedence.
// Otherwise, if nextMeta is available, its Header.AppHash is returned (auto-detect:
// block N+1's header carries the committed app hash after block N).
// Returning nil means the current app hash in state.db should not be changed.
func resolveAppHash(explicit []byte, nextMeta *types.BlockMeta) []byte {
	if explicit != nil {
		return explicit
	}
	if nextMeta != nil {
		return nextMeta.Header.AppHash
	}
	return nil
}

// Ensure types is used (for BlockID in state patching).
var _ = types.BlockID{}
