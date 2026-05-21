package fork

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	bstore "github.com/gnolang/gno/tm2/pkg/bft/store"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	mstore "github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

// dataDirTxsSource reads chain state directly from a gnoland node's data
// directory, with no RPC dependency. It expects the standard layout
// produced by `gnoland start`:
//
//	<dataDir>/
//	  db/blockstore.db        ← committed blocks (per-tx bytes + headers)
//	  db/state.db             ← ABCI responses (per-tx success/failure)
//	  db/gnolang.db           ← app multistore (auth accounts, etc.)
//
// PebbleDB is assumed (gnoland's default). The data directory must be a
// halted snapshot at the desired --halt-height: the source loads the
// latest committed app version, so any older state may have been pruned
// by the live node.
//
// The source's genesis is NOT read here — it lives behind GenesisSource
// (--source-genesis-* flags) and is supplied to FetchTxs as chainID.
type dataDirTxsSource struct {
	dataDir string

	bsDB    dbm.DB
	stateDB dbm.DB
	appDB   dbm.DB

	blockStore *bstore.BlockStore

	cms     mstore.CommitMultiStore
	mainKey mstore.StoreKey
	acck    auth.AccountKeeper
}

func newDataDirTxsSource(dataDir string) (s *dataDirTxsSource, err error) {
	dbDir := filepath.Join(dataDir, "db")
	for _, sub := range []string{"blockstore.db", "state.db", "gnolang.db"} {
		if _, statErr := os.Stat(filepath.Join(dbDir, sub)); statErr != nil {
			return nil, fmt.Errorf("%s not found under %s: %w", sub, dbDir, statErr)
		}
	}

	s = &dataDirTxsSource{dataDir: dataDir}
	// Cleanup defer: if any later step fails, close whatever DBs got opened
	// and zero out the returned source so the caller gets the conventional
	// (nil, err) shape. Error returns below MUST return s (not nil) so this
	// defer fires against a non-nil receiver; we nil it here after Close.
	defer func() {
		if err != nil {
			_ = s.Close()
			s = nil
		}
	}()

	if s.bsDB, err = dbm.NewDB("blockstore", dbm.PebbleDBBackend, dbDir); err != nil {
		return s, fmt.Errorf("opening blockstore.db: %w", err)
	}
	s.blockStore = bstore.NewBlockStore(s.bsDB)

	if s.stateDB, err = dbm.NewDB("state", dbm.PebbleDBBackend, dbDir); err != nil {
		return s, fmt.Errorf("opening state.db: %w", err)
	}

	if s.appDB, err = dbm.NewDB("gnolang", dbm.PebbleDBBackend, dbDir); err != nil {
		return s, fmt.Errorf("opening gnolang.db: %w", err)
	}

	// Set up a minimal auth keeper on the gnolang multistore so per-signer
	// (accNum, finalSeq) lookups can run without an RPC. Mirrors
	// gno.land/pkg/gnoland/app.go: same "main" + "base" store keys, same
	// constructors, same proto accounts.
	s.mainKey = mstore.NewStoreKey("main")
	baseKey := mstore.NewStoreKey("base")
	s.cms = mstore.NewCommitMultiStore(s.appDB)
	s.cms.MountStoreWithDB(s.mainKey, iavl.StoreConstructor, s.appDB)
	s.cms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, s.appDB)
	if err = s.cms.LoadLatestVersion(); err != nil {
		return s, fmt.Errorf("loading app multistore: %w", err)
	}
	// AccountKeeper.GetAccount reads directly from its store key and does
	// not consult the params keeper — we only need the keeper for its
	// constructor signature, so no prmk.Register here (the keeper's params
	// are never queried by the read-only flow we use).
	prmk := params.NewParamsKeeper(s.mainKey)
	s.acck = auth.NewAccountKeeper(
		s.mainKey,
		prmk.ForModule(auth.ModuleName),
		gnoland.ProtoGnoAccount,
		gnoland.ProtoGnoSessionAccount,
	)

	return s, nil
}

func (s *dataDirTxsSource) Description() string { return "gnoland data directory" }

func (s *dataDirTxsSource) Close() error {
	var errs []error
	for _, db := range []dbm.DB{s.bsDB, s.stateDB, s.appDB} {
		if db == nil {
			continue
		}
		if closeErr := db.Close(); closeErr != nil {
			errs = append(errs, closeErr)
		}
	}
	return errors.Join(errs...)
}

// LatestHeight returns the highest committed block height in the local
// blockstore. For a halted source chain that is the halt height.
func (s *dataDirTxsSource) LatestHeight(_ context.Context) (int64, error) {
	h := s.blockStore.Height()
	if h <= 0 {
		return 0, fmt.Errorf("blockstore is empty (height=%d) in %s", h, s.dataDir)
	}
	return h, nil
}

// FetchTxs walks every block in [fromHeight, toHeight] from blockstore.db,
// reads the matching ABCIResponses from state.db for per-tx success/failure,
// and runs the same sequence-brute-forcing pipeline rpcTxsSource uses, using
// the local auth keeper for the (accNum, finalSeq) lookup.
func (s *dataDirTxsSource) FetchTxs(ctx context.Context, chainID string, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	var txs []gnoland.TxWithMetadata

	bsHeight := s.blockStore.Height()
	if toHeight > bsHeight {
		return nil, fmt.Errorf("requested toHeight=%d exceeds local blockstore height %d", toHeight, bsHeight)
	}
	// queryAccountAtHeight reads the multistore at its latest committed
	// version (older versions may have been pruned). If --halt-height is set
	// below the snapshot's tip, signer (accNum, finalSeq) are read from a
	// later state than requested. accNum is stable post-creation and
	// finalSeq is monotonic, so brute-force still converges for accounts
	// that existed at toHeight — but accounts created after toHeight will
	// be visible at the (newer) latest version, mistreated as if existing
	// at toHeight. Flag this loudly so the operator knows what they're
	// getting.
	if toHeight < bsHeight {
		io.Printf("  WARNING: --halt-height=%d is below the local blockstore tip %d.\n"+
			"           Account state is read at the snapshot's latest committed version, not at halt-height.\n"+
			"           For a guaranteed faithful snapshot, truncate the data dir first (e.g. via gnobr).\n",
			toHeight, bsHeight)
	}

	signerStates := map[crypto.Address]*signerState{}
	getOrCreateSignerState := func(addr crypto.Address) *signerState {
		if ss, ok := signerStates[addr]; ok {
			return ss
		}
		acc := s.queryAccountAtHeight(addr, toHeight, chainID, io)
		ss := &signerState{}
		if acc != nil {
			ss.accNum = acc.GetAccountNumber()
			ss.finalSeq = acc.GetSequence()
		}
		signerStates[addr] = ss
		return ss
	}

	total := toHeight - fromHeight + 1
	var processed, txCount int64

	for h := fromHeight; h <= toHeight; h++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processed++
		if processed%1000 == 0 || processed == total {
			io.Printf("\r  Blocks: %d/%d  Txs: %d", processed, total, txCount)
		}

		block := s.blockStore.LoadBlock(h)
		if block == nil {
			return nil, fmt.Errorf("loading block %d: not found in blockstore", h)
		}

		if len(block.Data.Txs) == 0 {
			continue
		}

		responses, err := sm.LoadABCIResponses(s.stateDB, h)
		if err != nil {
			return nil, fmt.Errorf("loading ABCIResponses for block %d: %w", h, err)
		}

		timestamp := block.Header.Time.Unix()

		for i, rawTx := range block.Data.Txs {
			var stdTx std.Tx
			if err := amino.Unmarshal(rawTx, &stdTx); err != nil {
				io.Printf("\n  WARNING: could not decode tx at height %d index %d: %v\n", h, i, err)
				continue
			}

			failed := false
			if i < len(responses.DeliverTxs) && responses.DeliverTxs[i].IsErr() {
				failed = true
			}

			signers := stdTx.GetSigners()
			sigs := stdTx.GetSignatures()
			txIdx := len(txs)

			signerInfos := make([]gnoland.SignerAccountInfo, len(signers))
			for j, signer := range signers {
				ss := getOrCreateSignerState(signer)
				signerInfos[j] = gnoland.SignerAccountInfo{
					Address:    signer,
					AccountNum: ss.accNum,
					Sequence:   0,
				}
			}

			if !failed {
				for j, signer := range signers {
					ss := signerStates[signer]

					if !ss.initialized || len(ss.pendingFails) > 0 {
						lo := ss.seq
						hi := ss.seq + uint64(len(ss.pendingFails))
						if !ss.initialized {
							lo = 0
							hi = ss.finalSeq
						}

						var sig std.Signature
						if j < len(sigs) {
							sig = sigs[j]
						}

						resolvedSeq, err := bruteForceSignerSequence(
							stdTx, sig, ss.accNum, lo, hi, chainID)
						if err != nil {
							io.Printf("\n  WARNING: brute-force failed for signer %s at height %d: %v (using counter %d)\n",
								signer, h, err, ss.seq)
							resolvedSeq = ss.seq
						}

						assignFailedTxSequences(txs, ss.pendingFails, ss.seq, resolvedSeq)
						ss.pendingFails = nil
						ss.seq = resolvedSeq
						ss.initialized = true
					}

					signerInfos[j].Sequence = ss.seq
					ss.seq++
				}
			} else {
				for j, signer := range signers {
					ss := signerStates[signer]
					ss.pendingFails = append(ss.pendingFails, &pendingFailedTx{
						txIndex: txIdx,
						signerI: j,
					})
					signerInfos[j].Sequence = ss.seq
				}
			}

			txs = append(txs, gnoland.TxWithMetadata{
				Tx: stdTx,
				Metadata: &gnoland.GnoTxMetadata{
					Timestamp:   timestamp,
					BlockHeight: h,
					ChainID:     chainID,
					Failed:      failed,
					SignerInfo:  signerInfos,
				},
			})
			txCount++
		}
	}

	for _, ss := range signerStates {
		if len(ss.pendingFails) == 0 {
			continue
		}
		if !ss.initialized {
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			if consumed > uint64(len(ss.pendingFails)) {
				ss.seq = ss.finalSeq - uint64(len(ss.pendingFails))
				consumed = uint64(len(ss.pendingFails))
			}
			assignTrailingFailedTxSequences(txs, ss.pendingFails, ss.seq, consumed)
		} else {
			var consumed uint64
			if ss.finalSeq > ss.seq {
				consumed = ss.finalSeq - ss.seq
			}
			assignTrailingFailedTxSequences(txs, ss.pendingFails, ss.seq, consumed)
		}
	}

	io.Printf("\r  Blocks: %d/%d  Txs: %d\n", processed, total, txCount)
	return txs, nil
}

// queryAccountAtHeight returns the (accNum, sequence) state for addr as
// committed in the local app multistore. The multistore is loaded at the
// latest version, which for a halted snapshot equals the halt height —
// older versions may have been pruned and are not re-loaded.
func (s *dataDirTxsSource) queryAccountAtHeight(
	addr crypto.Address, height int64, chainID string, io commands.IO,
) std.Account {
	ctx := sdk.NewContext(
		sdk.RunTxModeCheck,
		s.cms,
		&bft.Header{Height: height, ChainID: chainID},
		log.NewNoopLogger(),
	)
	acc := s.acck.GetAccount(ctx, addr)
	if acc == nil {
		io.Printf("\n  NOTE: account %s not found in local app state at height %d (treating as accNum=0, finalSeq=0)\n",
			addr, height)
	}
	return acc
}
