package gnoland

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	bftcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

const (
	// migTmpName is the DB name used for the in-progress migration database.
	migTmpName = "gnolang-mig"

	// migBakName is the DB name used when the original database is backed up
	// after a successful migration.
	migBakName = "gnolang.bak"
)

// MigrationConfig holds the parameters for an in-place block-replay migration.
type MigrationConfig struct {
	// DataRootDir is the node data root directory (parent of the "db/" subdir).
	DataRootDir string

	// GenesisPath is the path to the genesis.json that will be used to
	// initialise the fresh application state before replaying blocks.
	// This is typically the *new* chain's genesis (possibly with a migration
	// overlay applied).
	GenesisPath string

	// DBBackend is the database backend used for the block store and state DB.
	// It must match the backend that was used when the chain was running.
	// Defaults to PebbleDB if empty.
	DBBackend dbm.BackendType

	// GenesisOverlay, if non-nil, is called with the loaded genesis document
	// and may return a modified version.  Use this to apply chain-ID renames,
	// param changes, or any other genesis-level migration before replay starts.
	GenesisOverlay func(*bft.GenesisDoc) *bft.GenesisDoc

	// NewApp is a factory that creates the gno.land ABCI application using the
	// provided database.  The caller supplies this so that migration.go does not
	// need to know about stdlib dirs, log formats, etc.
	NewApp func(db dbm.DB) (abci.Application, error)

	// Logger for migration progress messages.  Falls back to slog.Default().
	Logger *slog.Logger
}

// RunInPlaceMigration performs an in-place block-replay migration.
//
// The migration works as follows:
//
//  1. Open the existing block store and TM state DB (read-only references).
//  2. Load (and optionally overlay) the genesis document.
//  3. Create a fresh application database at "<dbDir>/gnolang-mig.db".
//  4. Initialise the fresh app via InitChain.
//  5. Replay every block from height 1 through blockStore.Height() using the
//     new binary's ABCI logic, preserving original timestamps and transactions.
//  6. On success, atomically replace the live database:
//     "<dbDir>/gnolang.db"     → "<dbDir>/gnolang.bak.db"  (backup)
//     "<dbDir>/gnolang-mig.db" → "<dbDir>/gnolang.db"      (new live)
//
// The backup is kept so operators can roll back if needed.  A subsequent
// successful migration will refuse to run if the backup still exists; remove it
// manually after verifying the migration.
//
// The TM state DB ("state.db") and the block store ("blockstore.db") are NOT
// touched; only the application DB ("gnolang.db") is replaced.
func RunInPlaceMigration(cfg MigrationConfig) error {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	backend := cfg.DBBackend
	if backend == "" {
		backend = dbm.PebbleDBBackend
	}

	dbDir := filepath.Join(cfg.DataRootDir, bftcfg.DefaultDBDir)

	// Refuse to overwrite an existing backup – it means either a previous
	// migration succeeded (and the backup was never removed) or a partial
	// migration left stale data.
	bakPath := filepath.Join(dbDir, migBakName+".db")
	if _, err := os.Stat(bakPath); err == nil {
		return fmt.Errorf(
			"migration backup already exists at %q; "+
				"if a previous migration succeeded, remove the backup directory and retry; "+
				"if it failed, also remove %q and retry",
			bakPath,
			filepath.Join(dbDir, migTmpName+".db"),
		)
	}

	logger.Info("Starting in-place block-replay migration", "dataDir", cfg.DataRootDir)

	// ------------------------------------------------------------------
	// Open existing stores (read-only for our purposes).
	// ------------------------------------------------------------------

	blockStoreDB, err := dbm.NewDB("blockstore", backend, dbDir)
	if err != nil {
		return fmt.Errorf("opening block store db: %w", err)
	}
	defer blockStoreDB.Close()

	blockStore := store.NewBlockStore(blockStoreDB)
	haltHeight := blockStore.Height()
	if haltHeight == 0 {
		return fmt.Errorf("block store is empty; nothing to migrate")
	}
	logger.Info("Will replay blocks", "from", 1, "to", haltHeight)

	// The TM state DB is needed only to supply validator-set information to
	// BeginBlock during replay; we never write to it.
	stateDB, err := dbm.NewDB("state", backend, dbDir)
	if err != nil {
		return fmt.Errorf("opening state db: %w", err)
	}
	defer stateDB.Close()

	// ------------------------------------------------------------------
	// Load and optionally overlay the genesis document.
	// ------------------------------------------------------------------

	genDoc, err := bft.GenesisDocFromFile(cfg.GenesisPath)
	if err != nil {
		return fmt.Errorf("loading genesis file %q: %w", cfg.GenesisPath, err)
	}
	if cfg.GenesisOverlay != nil {
		genDoc = cfg.GenesisOverlay(genDoc)
	}

	// ------------------------------------------------------------------
	// Create the migration (temp) application database.
	// ------------------------------------------------------------------

	tmpAppDB, err := dbm.NewDB(migTmpName, dbm.PebbleDBBackend, dbDir)
	if err != nil {
		return fmt.Errorf("creating migration temp db: %w", err)
	}

	// Run the replay; clean up on failure.
	replayErr := runMigrationReplay(cfg.NewApp, tmpAppDB, stateDB, blockStore, genDoc, haltHeight, logger)
	tmpAppDB.Close()

	if replayErr != nil {
		// Best-effort cleanup of the partial temp DB.
		_ = os.RemoveAll(filepath.Join(dbDir, migTmpName+".db"))
		return fmt.Errorf("block-replay migration failed: %w", replayErr)
	}

	// ------------------------------------------------------------------
	// Atomic swap: gnolang.db → gnolang.bak.db, gnolang-mig.db → gnolang.db
	// ------------------------------------------------------------------

	origPath := filepath.Join(dbDir, "gnolang.db")
	tmpPath := filepath.Join(dbDir, migTmpName+".db")

	logger.Info("Swapping application databases",
		"original", origPath,
		"backup", bakPath,
		"migration", tmpPath,
	)

	if err := os.Rename(origPath, bakPath); err != nil {
		return fmt.Errorf("backing up original db: %w", err)
	}
	if err := os.Rename(tmpPath, origPath); err != nil {
		// Attempt to restore the original so the node can still start.
		if restoreErr := os.Rename(bakPath, origPath); restoreErr != nil {
			logger.Error("CRITICAL: failed to restore backup after failed swap; manual intervention required",
				"backup", bakPath, "err", restoreErr)
		}
		return fmt.Errorf("promoting migration db: %w", err)
	}

	logger.Info("In-place block-replay migration completed",
		"replayedBlocks", haltHeight,
		"backup", bakPath,
	)
	return nil
}

// runMigrationReplay creates the migration application, initialises genesis,
// and replays all blocks from height 1 to haltHeight using ExecCommitBlock.
func runMigrationReplay(
	newApp func(db dbm.DB) (abci.Application, error),
	tmpAppDB dbm.DB,
	stateDB dbm.DB,
	blockStore *store.BlockStore,
	genDoc *bft.GenesisDoc,
	haltHeight int64,
	logger *slog.Logger,
) error {
	// Create the fresh application backed by the temp DB.
	migApp, err := newApp(tmpAppDB)
	if err != nil {
		return fmt.Errorf("creating migration app: %w", err)
	}

	// Wire up proxy app connections (local, in-process).
	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(migApp))
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return fmt.Errorf("starting proxy app connections: %w", err)
	}
	defer proxyApp.Stop()

	// ------------------------------------------------------------------
	// InitChain: feed genesis state to the fresh application.
	// ------------------------------------------------------------------

	validators := make([]*bft.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		validators[i] = bft.NewValidator(val.PubKey, val.Power)
	}
	csParams := genDoc.ConsensusParams
	initRes, err := proxyApp.Consensus().InitChainSync(abci.RequestInitChain{
		Time:            genDoc.GenesisTime,
		ChainID:         genDoc.ChainID,
		ConsensusParams: &csParams,
		Validators:      bft.NewValidatorSet(validators).ABCIValidatorUpdates(),
		AppState:        genDoc.AppState,
	})
	if err != nil {
		return fmt.Errorf("InitChain: %w", err)
	}
	logger.Info("InitChain complete",
		"chainID", genDoc.ChainID,
		"genesisTxs", len(initRes.TxResponses),
	)

	// ------------------------------------------------------------------
	// Replay blocks 1 .. haltHeight.
	//
	// ExecCommitBlock replays a single block against the ABCI app without
	// mutating the TM state.  It reads validator information from stateDB
	// (the original state DB) so it can populate BeginBlock.LastCommitInfo
	// correctly.
	// ------------------------------------------------------------------

	for h := int64(1); h <= haltHeight; h++ {
		block := blockStore.LoadBlock(h)
		if block == nil {
			return fmt.Errorf("block %d not found in block store", h)
		}

		if _, err := sm.ExecCommitBlock(proxyApp.Consensus(), block, logger, stateDB); err != nil {
			return fmt.Errorf("replaying block %d: %w", h, err)
		}

		if h%1000 == 0 || h == haltHeight {
			logger.Info("Migration replay progress", "height", h, "total", haltHeight)
		}
	}

	return nil
}
