package store

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// A cleanupFunc cleans up any config / test files created for a particular
// test.
type cleanupFunc func()

// make a Commit with a single vote containing just the height and a timestamp
func makeTestCommit(height int64, timestamp time.Time) *types.Commit {
	commitSigs := []*types.CommitSig{{Height: height, Timestamp: timestamp}}
	return types.NewCommit(types.BlockID{}, commitSigs)
}

func makeTxs(height int64) (txs []types.Tx) {
	for i := 0; i < 10; i++ {
		txs = append(txs, types.Tx([]byte{byte(height), byte(i)}))
	}
	return txs
}

func makeBlock(height int64, state sm.State, lastCommit *types.Commit) *types.Block {
	block, _ := state.MakeBlock(height, makeTxs(height), lastCommit, state.Validators.GetProposer().Address)
	return block
}

func makeStateAndBlockStore(logger log.Logger) (sm.State, *BlockStore, cleanupFunc) {
	config := cfg.ResetTestRoot("blockchain_reactor_test")
	// blockDB := dbm.NewDebugDB("blockDB", dbm.NewMemDB())
	// stateDB := dbm.NewDebugDB("stateDB", dbm.NewMemDB())
	blockDB := dbm.NewMemDB()
	stateDB := dbm.NewMemDB()
	state, err := sm.LoadStateFromDBOrGenesisFile(stateDB, config.GenesisFile())
	if err != nil {
		panic(errors.Wrap(err, "error constructing state from genesis file"))
	}

	bs, err := NewBlockStore(blockDB)
	if err != nil {
		panic(fmt.Errorf("error creating BlockStore: %w", err))
	}
	return state, bs, func() { os.RemoveAll(config.RootDir) }
}

func TestLoadBlockStoreStateJSON(t *testing.T) {
	db := dbm.NewMemDB()

	bsj := &BlockStoreStateJSON{Height: 1000}
	bsj.Save(db)

	retrBSJ, err := LoadBlockStoreStateJSON(db)
	require.NoError(t, err)

	assert.Equal(t, *bsj, retrBSJ, "expected the retrieved DBs to match")
}

func TestNewBlockStore(t *testing.T) {
	db := dbm.NewMemDB()
	db.Set(blockStoreKey, []byte(`{"height": "10000"}`))
	bs, err := NewBlockStore(db)
	require.NoError(t, err)
	h, err := bs.Height()
	require.NoError(t, err)
	require.Equal(t, int64(10000), h, "failed to properly parse blockstore")

	errorCausers := []struct {
		data    []byte
		wantErr string
	}{
		{[]byte("artful-doger"), "could not unmarshall bytes 61727466756C2D646F676572: invalid character 'a' looking for beginning of value"},
		{[]byte(" "), "could not unmarshall bytes 20: unexpected end of JSON input"},
	}

	for i, tt := range errorCausers {
		// Expecting an error here on trying to parse an invalid blockStore
		db.Set(blockStoreKey, tt.data)
		_, err = NewBlockStore(db)
		require.Error(t, err)
		require.ErrorContains(t, err, tt.wantErr, "#%d data: %q", i, tt.data)
	}

	db.Set(blockStoreKey, nil)
	bs, err = NewBlockStore(db)
	require.NoError(t, err)
	h, err = bs.Height()
	require.NoError(t, err)
	assert.Equal(t, h, int64(0), "expecting nil bytes to be unmarshalled alright")
}

func freshBlockStore() (*BlockStore, dbm.DB) {
	db := dbm.NewMemDB()
	bs, err := NewBlockStore(db)
	if err != nil {
		panic(err)
	}
	return bs, db
}

var (
	state       sm.State
	block       *types.Block
	partSet     *types.PartSet
	part1       *types.Part
	part2       *types.Part
	seenCommit1 *types.Commit
)

func TestMain(m *testing.M) {
	var cleanup cleanupFunc
	state, _, cleanup = makeStateAndBlockStore(log.NewTMLogger(new(bytes.Buffer)))
	block = makeBlock(1, state, new(types.Commit))
	partSet = block.MakePartSet(2)
	part1 = partSet.GetPart(0)
	part2 = partSet.GetPart(1)
	seenCommit1 = makeTestCommit(10, tmtime.Now())
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// TODO: This test should be simplified ...

func TestBlockStoreSaveLoadBlock(t *testing.T) {
	state, bs, cleanup := makeStateAndBlockStore(log.NewTMLogger(new(bytes.Buffer)))
	defer cleanup()
	h, err := bs.Height()
	require.NoError(t, err)
	require.Equal(t, h, int64(0), "initially the height should be zero")

	// check there are no blocks at various heights
	noBlockHeights := []int64{0, -1, 100, 1000, 2}
	for i, height := range noBlockHeights {
		g, err := bs.LoadBlock(height)
		require.NoError(t, err)
		if g != nil {
			t.Errorf("#%d: height(%d) got a block; want nil", i, height)
		}
	}

	// save a block
	block := makeBlock(h+1, state, new(types.Commit))
	validPartSet := block.MakePartSet(2)
	seenCommit := makeTestCommit(10, tmtime.Now())
	bs.SaveBlock(block, partSet, seenCommit)

	h, err = bs.Height()
	require.NoError(t, err)

	require.Equal(t, h, block.Header.Height, "expecting the new height to be changed")

	incompletePartSet := types.NewPartSetFromHeader(types.PartSetHeader{Total: 2})
	uncontiguousPartSet := types.NewPartSetFromHeader(types.PartSetHeader{Total: 0})
	uncontiguousPartSet.AddPart(part2)

	header1 := types.Header{
		Height:  1,
		NumTxs:  100,
		ChainID: "block_test",
		Time:    tmtime.Now(),
	}
	header2 := header1
	header2.Height = 4

	// End of setup, test data

	commitAtH10 := makeTestCommit(10, tmtime.Now())
	tuples := []struct {
		block      *types.Block
		parts      *types.PartSet
		seenCommit *types.Commit
		wantErr    string
		wantPanic  string

		corruptBlockInDB      bool
		corruptCommitInDB     bool
		corruptSeenCommitInDB bool
		eraseCommitInDB       bool
		eraseSeenCommitInDB   bool
	}{
		{
			block:      newBlock(header1, commitAtH10),
			parts:      validPartSet,
			seenCommit: seenCommit1,
		},

		{
			block:   nil,
			wantErr: "only save a non-nil block",
		},

		{
			block:   newBlock(header2, commitAtH10),
			parts:   uncontiguousPartSet,
			wantErr: "only save contiguous blocks", // and incomplete and uncontiguous parts
		},

		{
			block:   newBlock(header1, commitAtH10),
			parts:   incompletePartSet,
			wantErr: "only save complete block", // incomplete parts
		},

		{
			block:             newBlock(header1, commitAtH10),
			parts:             validPartSet,
			seenCommit:        seenCommit1,
			corruptCommitInDB: true, // Corrupt the DB's commit entry
			wantErr:           "unmarshal to types.Commit failed",
		},

		{
			block:            newBlock(header1, commitAtH10),
			parts:            validPartSet,
			seenCommit:       seenCommit1,
			wantErr:          "unmarshal to types.BlockMeta failed",
			corruptBlockInDB: true, // Corrupt the DB's block entry
		},

		{
			block:      newBlock(header1, commitAtH10),
			parts:      validPartSet,
			seenCommit: seenCommit1,

			// Expecting no error and we want a nil back
			eraseSeenCommitInDB: true,
		},

		{
			block:      newBlock(header1, commitAtH10),
			parts:      validPartSet,
			seenCommit: seenCommit1,

			corruptSeenCommitInDB: true,
			wantErr:               "unmarshal to types.Commit failed",
		},

		{
			block:      newBlock(header1, commitAtH10),
			parts:      validPartSet,
			seenCommit: seenCommit1,

			// Expecting no error and we want a nil back
			eraseCommitInDB: true,
		},
	}

	type quad struct {
		block  *types.Block
		commit *types.Commit
		meta   *types.BlockMeta

		seenCommit *types.Commit
	}

	for i, tuple := range tuples {
		tuple := tuple
		bs, db := freshBlockStore()
		// SaveBlock
		res, err, panicErr := doFn(func() (interface{}, error) {
			err := bs.SaveBlock(tuple.block, tuple.parts, tuple.seenCommit)
			if err != nil { // TODO refactor this tests to check errors, not panics.
				return nil, err
			}

			if tuple.block == nil {
				return nil, nil
			}

			if tuple.corruptBlockInDB {
				db.Set(calcBlockMetaKey(tuple.block.Height), []byte("block-bogus"))
			}
			bBlock, err := bs.LoadBlock(tuple.block.Height)
			if err != nil { // TODO refactor this tests to check errors, not panics.
				return nil, err
			}
			bBlockMeta, err := bs.LoadBlockMeta(tuple.block.Height)
			if err != nil { // TODO refactor this tests to check errors, not panics.
				return nil, err
			}
			if tuple.eraseSeenCommitInDB {
				db.Delete(calcSeenCommitKey(tuple.block.Height))
			}
			if tuple.corruptSeenCommitInDB {
				db.Set(calcSeenCommitKey(tuple.block.Height), []byte("bogus-seen-commit"))
			}
			bSeenCommit, err := bs.LoadSeenCommit(tuple.block.Height)
			if err != nil { // TODO refactor this tests to check errors, not panics.
				return nil, err
			}

			commitHeight := tuple.block.Height - 1
			if tuple.eraseCommitInDB {
				db.Delete(calcBlockCommitKey(commitHeight))
			}
			if tuple.corruptCommitInDB {
				db.Set(calcBlockCommitKey(commitHeight), []byte("foo-bogus"))
			}
			bCommit, err := bs.LoadBlockCommit(commitHeight)
			if err != nil { // TODO refactor this tests to check errors, not panics.
				return nil, err
			}

			return &quad{
				block: bBlock, seenCommit: bSeenCommit, commit: bCommit,
				meta: bBlockMeta,
			}, nil
		})

		if subStr := tuple.wantPanic; subStr != "" {
			if panicErr == nil {
				t.Errorf("#%d: want a non-nil panic", i)
			} else if got := fmt.Sprintf("%#v\n%v", panicErr, panicErr); !strings.Contains(got, subStr) {
				t.Errorf("#%d:\n\tgotErr: %q\nwant substring: %q", i, got, subStr)
			}
			continue
		}

		if tuple.wantErr != "" {
			require.ErrorContains(t, err, tuple.wantErr)
			continue
		}

		assert.Nil(t, panicErr, "#%d: unexpected panic", i)
		assert.Nil(t, err, "#%d: expecting a non-nil error", i)
		qua, ok := res.(*quad)
		if !ok || qua == nil {
			t.Errorf("#%d: got nil quad back; gotType=%T", i, res)
			continue
		}
		if tuple.eraseSeenCommitInDB {
			assert.Nil(t, qua.seenCommit,
				"erased the seenCommit in the DB hence we should get back a nil seenCommit")
		}
		if tuple.eraseCommitInDB {
			assert.Nil(t, qua.commit,
				"erased the commit in the DB hence we should get back a nil commit")
		}
	}
}

func TestLoadBlockPart(t *testing.T) {
	bs, db := freshBlockStore()
	height, index := int64(10), 1
	loadPart := func() (interface{}, error) {
		return bs.LoadBlockPart(height, index)
	}

	// Initially no contents.
	// 1. Requesting for a non-existent block shouldn't fail
	res, err, panicErr := doFn(loadPart)
	require.Nil(t, panicErr, "a non-existent block part shouldn't cause a panic")
	require.Nil(t, err, "a non-existent block part shouldn't cause a panic")

	require.Nil(t, res, "a non-existent block part should return nil")

	// 2. Next save a corrupted block then try to load it
	db.Set(calcBlockPartKey(height, index), []byte("Tendermint"))
	res, err, panicErr = doFn(loadPart)
	require.NotNil(t, err, "expecting a non-nil panic")
	require.Nil(t, panicErr, "expecting a nil panic")

	require.Contains(t, err.Error(), "unmarshal to types.Part failed")

	// 3. A good block serialized and saved to the DB should be retrievable
	db.Set(calcBlockPartKey(height, index), amino.MustMarshal(part1))
	gotPart, _, panicErr := doFn(loadPart)
	require.Nil(t, panicErr, "an existent and proper block should not panic")
	require.Nil(t, res, "a properly saved block should return a proper block")
	require.Equal(t, gotPart.(*types.Part), part1,
		"expecting successful retrieval of previously saved block")
}

func TestLoadBlockMeta(t *testing.T) {
	bs, db := freshBlockStore()
	height := int64(10)
	loadMeta := func() (interface{}, error) {
		return bs.LoadBlockMeta(height)
	}

	// Initially no contents.
	// 1. Requesting for a non-existent blockMeta shouldn't fail
	res, err, panicErr := doFn(loadMeta)
	require.Nil(t, panicErr, "a non-existent blockMeta shouldn't cause a panic")
	require.Nil(t, err, "a non-existent blockMeta shouldn't cause an error")
	require.Nil(t, res, "a non-existent blockMeta should return nil")

	// 2. Next save a corrupted blockMeta then try to load it
	db.Set(calcBlockMetaKey(height), []byte("Tendermint-Meta"))
	res, err, panicErr = doFn(loadMeta)
	require.Nil(t, panicErr, "expecting a nil panic")
	require.NotNil(t, err, "expecting a non-nil error")
	require.Contains(t, err.Error(), "unmarshal to types.BlockMeta")

	// 3. A good blockMeta serialized and saved to the DB should be retrievable
	meta := &types.BlockMeta{}
	db.Set(calcBlockMetaKey(height), amino.MustMarshal(meta))
	gotMeta, _, panicErr := doFn(loadMeta)
	require.Nil(t, panicErr, "an existent and proper block should not panic")
	require.Nil(t, res, "a properly saved blockMeta should return a proper blocMeta ")
	require.Equal(t, amino.MustMarshal(meta), amino.MustMarshal(gotMeta),
		"expecting successful retrieval of previously saved blockMeta")
}

func TestBlockFetchAtHeight(t *testing.T) {
	state, bs, cleanup := makeStateAndBlockStore(log.NewTMLogger(new(bytes.Buffer)))
	defer cleanup()
	h, err := bs.Height()
	require.NoError(t, err)
	require.Equal(t, h, int64(0), "initially the height should be zero")
	block := makeBlock(h+1, state, new(types.Commit))

	partSet := block.MakePartSet(2)
	seenCommit := makeTestCommit(10, tmtime.Now())
	bs.SaveBlock(block, partSet, seenCommit)
	h, err = bs.Height()
	require.NoError(t, err)
	require.Equal(t, h, block.Header.Height, "expecting the new height to be changed")

	blockAtHeight, err := bs.LoadBlock(h)
	require.NoError(t, err)
	bz1 := amino.MustMarshal(block)
	bz2 := amino.MustMarshal(blockAtHeight)
	require.Equal(t, bz1, bz2)
	require.Equal(t, block.Hash(), blockAtHeight.Hash(),
		"expecting a successful load of the last saved block")

	blockAtHeightPlus1, err := bs.LoadBlock(h + 1)
	require.NoError(t, err)
	require.Nil(t, blockAtHeightPlus1, "expecting an unsuccessful load of Height()+1")
	blockAtHeightPlus2, err := bs.LoadBlock(h + 2)
	require.NoError(t, err)
	require.Nil(t, blockAtHeightPlus2, "expecting an unsuccessful load of Height()+2")
}

func doFn(fn func() (interface{}, error)) (res interface{}, err error, panicErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case error:
				panicErr = e
			case string:
				panicErr = fmt.Errorf("%s", e)
			default:
				if st, ok := r.(fmt.Stringer); ok {
					panicErr = fmt.Errorf("%s", st)
				} else {
					panicErr = fmt.Errorf("%s", debug.Stack())
				}
			}
		}
	}()

	res, err = fn()
	return res, err, panicErr
}

func newBlock(hdr types.Header, lastCommit *types.Commit) *types.Block {
	return &types.Block{
		Header:     hdr,
		LastCommit: lastCommit,
	}
}
