package types

import (
	// it is ok to use math/rand here: we do not need a cryptographically secure random
	// number generator here and we can run the tests a bit faster
	"crypto/rand"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	typesver "github.com/gnolang/gno/tm2/pkg/bft/types/version"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/random"
)

func TestBlockValidateBasic(t *testing.T) {
	t.Parallel()

	require.Error(t, (*Block)(nil).ValidateBasic())

	txs := []Tx{Tx("foo"), Tx("bar")}
	lastID := makeBlockIDRandom()
	h := int64(3)

	voteSet, valSet, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	require.NoError(t, err)

	testCases := []struct {
		testName      string
		malleateBlock func(*Block)
		expErr        bool
	}{
		{"Make Block", func(blk *Block) {}, false},
		{"Make Block w/ proposer Addr", func(blk *Block) { blk.ProposerAddress = valSet.GetProposer().Address }, false},
		{"Negative Height", func(blk *Block) { blk.Height = -1 }, true},
		{"Increase NumTxs", func(blk *Block) { blk.NumTxs++ }, true},
		{"Remove 1/2 the commits", func(blk *Block) {
			blk.LastCommit.Precommits = commit.Precommits[:commit.Size()/2]
			blk.LastCommit.hash = nil // clear hash or change wont be noticed
		}, true},
		{"Remove LastCommitHash", func(blk *Block) { blk.LastCommitHash = []byte("something else") }, true},
		{"Tampered Data", func(blk *Block) {
			blk.Data.Txs[0] = Tx("something else")
			blk.Data.hash = nil // clear hash or change wont be noticed
		}, true},
		{"Tampered DataHash", func(blk *Block) {
			blk.DataHash = random.RandBytes(len(blk.DataHash))
		}, true},
	}
	for i, tc := range testCases {
		tc := tc
		i := i
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			block := MakeBlock(h, txs, commit)
			block.ProposerAddress = valSet.GetProposer().Address
			tc.malleateBlock(block)
			err = block.ValidateBasic()
			assert.Equal(t, tc.expErr, err != nil, "#%d: %v", i, err)
		})
	}
}

func TestBlockHash(t *testing.T) {
	t.Parallel()

	assert.Nil(t, (*Block)(nil).Hash())
	assert.Nil(t, MakeBlock(int64(3), []Tx{Tx("Hello World")}, nil).Hash())
}

func TestBlockMakePartSet(t *testing.T) {
	t.Parallel()

	assert.Nil(t, (*Block)(nil).MakePartSet(2))

	partSet := MakeBlock(int64(3), []Tx{Tx("Hello World")}, nil).MakePartSet(1024)
	assert.NotNil(t, partSet)
	assert.Equal(t, 1, partSet.Total())
}

func TestBlockHashesTo(t *testing.T) {
	t.Parallel()

	assert.False(t, (*Block)(nil).HashesTo(nil))

	lastID := makeBlockIDRandom()
	h := int64(3)
	voteSet, valSet, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	require.NoError(t, err)

	block := MakeBlock(h, []Tx{Tx("Hello World")}, commit)
	block.ValidatorsHash = valSet.Hash()
	assert.False(t, block.HashesTo([]byte{}))
	assert.False(t, block.HashesTo([]byte("something else")))
	assert.True(t, block.HashesTo(block.Hash()))
}

func TestBlockSize(t *testing.T) {
	t.Parallel()

	size := MakeBlock(int64(3), []Tx{Tx("Hello World")}, nil).Size()
	if size <= 0 {
		t.Fatal("Size of the block is zero or negative")
	}
}

func TestBlockString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "nil-Block", (*Block)(nil).String())
	assert.Equal(t, "nil-Block", (*Block)(nil).StringIndented(""))
	assert.Equal(t, "nil-Block", (*Block)(nil).StringShort())

	block := MakeBlock(int64(3), []Tx{Tx("Hello World")}, nil)
	assert.NotEqual(t, "nil-Block", block.String())
	assert.NotEqual(t, "nil-Block", block.StringIndented(""))
	assert.NotEqual(t, "nil-Block", block.StringShort())
}

func makeBlockIDRandom() BlockID {
	blockHash := make([]byte, tmhash.Size)
	partSetHash := make([]byte, tmhash.Size)
	rand.Read(blockHash)
	rand.Read(partSetHash)
	blockPartsHeader := PartSetHeader{123, partSetHash}
	return BlockID{blockHash, blockPartsHeader}
}

func makeBlockID(hash []byte, partSetSize int, partSetHash []byte) BlockID {
	return BlockID{
		Hash: hash,
		PartsHeader: PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

var nilBytes []byte

func TestNilHeaderHashDoesntCrash(t *testing.T) {
	t.Parallel()

	assert.Equal(t, (*Header)(nil).Hash(), nilBytes)
	assert.Equal(t, (new(Header)).Hash(), nilBytes)
}

func TestNilDataHashDoesntCrash(t *testing.T) {
	t.Parallel()

	assert.Equal(t, (*Data)(nil).Hash(), nilBytes)
	assert.Equal(t, new(Data).Hash(), nilBytes)
}

func TestCommit(t *testing.T) {
	t.Parallel()

	lastID := makeBlockIDRandom()
	h := int64(3)
	voteSet, _, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	require.NoError(t, err)

	assert.Equal(t, h-1, commit.Height())
	assert.Equal(t, 1, commit.Round())
	assert.Equal(t, PrecommitType, SignedMsgType(commit.Type()))
	if commit.Size() <= 0 {
		t.Fatalf("commit %v has a zero or negative size: %d", commit, commit.Size())
	}

	require.NotNil(t, commit.BitArray())
	assert.Equal(t, bitarray.NewBitArray(10).Size(), commit.BitArray().Size())

	assert.Equal(t, voteSet.GetByIndex(0), commit.GetByIndex(0))
	assert.True(t, commit.IsCommit())
}

func TestCommitValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName       string
		malleateCommit func(*Commit)
		expectErr      bool
	}{
		{"Random Commit", func(com *Commit) {}, false},
		{"Nil precommit", func(com *Commit) { com.Precommits[0] = nil }, false},
		{"Incorrect signature", func(com *Commit) { com.Precommits[0].Signature = []byte{0} }, false},
		{"Incorrect type", func(com *Commit) { com.Precommits[0].Type = PrevoteType }, true},
		{"Incorrect height", func(com *Commit) { com.Precommits[0].Height = int64(100) }, true},
		{"Incorrect round", func(com *Commit) { com.Precommits[0].Round = 100 }, true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			com := randCommit()
			tc.malleateCommit(com)
			assert.Equal(t, tc.expectErr, com.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestHeaderByteSize(t *testing.T) {
	t.Parallel()

	// Construct a UTF-8 string of MaxChainIDLen length using the supplementary
	// characters.
	// Each supplementary character takes 4 bytes.
	// http://www.i18nguy.com/unicode/supplementary-test.html
	maxChainID := ""
	for range MaxChainIDLen {
		maxChainID += "𠜎"
	}

	// time is varint encoded so need to pick the max.
	// year int, month Month, day, hour, min, sec, nsec int, loc *Location
	timestamp := time.Date(math.MaxInt64, 0, 0, 0, 0, 0, math.MaxInt64, time.UTC)

	h := Header{
		Version:            typesver.BlockVersion,
		ChainID:            maxChainID,
		Height:             math.MaxInt64,
		Time:               timestamp,
		NumTxs:             math.MaxInt64,
		TotalTxs:           math.MaxInt64,
		AppVersion:         "v0.0.0-test",
		LastBlockID:        makeBlockID(make([]byte, tmhash.Size), math.MaxInt64, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     tmhash.Sum([]byte("validators_hash")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators_hash")),
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		ProposerAddress:    crypto.AddressFromPreimage([]byte("proposer_address")),
	}

	bz, err := amino.MarshalSized(h)
	require.NoError(t, err)

	assert.EqualValues(t, 647, len(bz))
}

func randCommit() *Commit {
	lastID := makeBlockIDRandom()
	h := int64(3)
	voteSet, _, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	if err != nil {
		panic(err)
	}
	return commit
}

func TestCommitToVoteSet(t *testing.T) {
	t.Parallel()

	lastID := makeBlockIDRandom()
	h := int64(3)

	voteSet, valSet, vals := randVoteSet(h-1, 1, PrecommitType, 10, 1)
	commit, err := MakeCommit(lastID, h-1, 1, voteSet, vals)
	assert.NoError(t, err)

	chainID := voteSet.ChainID()
	voteSet2 := CommitToVoteSet(chainID, commit, valSet)

	for i := range vals {
		vote1 := voteSet.GetByIndex(i)
		vote2 := voteSet2.GetByIndex(i)
		vote3 := commit.GetVote(i)

		vote1bz := amino.MustMarshal(vote1)
		vote2bz := amino.MustMarshal(vote2)
		vote3bz := amino.MustMarshal(vote3)
		assert.Equal(t, vote1bz, vote2bz)
		assert.Equal(t, vote1bz, vote3bz)
	}
}

func TestCommitToVoteSetWithVotesForAnotherBlockOrNilBlock(t *testing.T) {
	t.Parallel()

	blockID := makeBlockID([]byte("blockhash"), 1000, []byte("partshash"))
	blockID2 := makeBlockID([]byte("blockhash2"), 1000, []byte("partshash"))
	blockID3 := makeBlockID([]byte("blockhash3"), 10000, []byte("partshash"))

	height := int64(3)
	round := 1

	type commitVoteTest struct {
		blockIDs      []BlockID
		numVotes      []int // must sum to numValidators
		numValidators int
		valid         bool
	}

	testCases := []commitVoteTest{
		{[]BlockID{blockID, blockID2, blockID3}, []int{8, 1, 1}, 10, true},
		{[]BlockID{blockID, blockID2, blockID3}, []int{67, 20, 13}, 100, true},
		{[]BlockID{blockID, blockID2, blockID3}, []int{1, 1, 1}, 3, false},
		{[]BlockID{blockID, blockID2, blockID3}, []int{3, 1, 1}, 5, false},
		{[]BlockID{blockID, {}}, []int{67, 33}, 100, true},
		{[]BlockID{blockID, blockID2, {}}, []int{10, 5, 5}, 20, false},
	}

	for _, tc := range testCases {
		voteSet, valSet, vals := randVoteSet(height-1, 1, PrecommitType, tc.numValidators, 1)

		vi := 0
		for n := range tc.blockIDs {
			for range tc.numVotes[n] {
				addr := vals[vi].PubKey().Address()
				vote := &Vote{
					ValidatorAddress: addr,
					ValidatorIndex:   vi,
					Height:           height - 1,
					Round:            round,
					Type:             PrecommitType,
					BlockID:          tc.blockIDs[n],
					Timestamp:        tmtime.Now(),
				}

				_, err := signAddVote(vals[vi], vote, voteSet)
				assert.NoError(t, err)
				vi++
			}
		}
		if tc.valid {
			commit := voteSet.MakeCommit() // panics without > 2/3 valid votes
			assert.NotNil(t, commit)
			err := valSet.VerifyCommit(voteSet.ChainID(), blockID, height-1, commit)
			assert.Nil(t, err)
		} else {
			assert.Panics(t, func() { voteSet.MakeCommit() })
		}
	}
}

func TestSignedHeaderValidateBasic(t *testing.T) {
	t.Parallel()

	commit := randCommit()
	chainID := "𠜎"
	timestamp := time.Date(math.MaxInt64, 0, 0, 0, 0, 0, math.MaxInt64, time.UTC)
	h := Header{
		Version:            typesver.BlockVersion,
		ChainID:            chainID,
		Height:             commit.Height(),
		Time:               timestamp,
		NumTxs:             math.MaxInt64,
		TotalTxs:           math.MaxInt64,
		AppVersion:         "v0.0.0-test",
		LastBlockID:        commit.BlockID,
		LastCommitHash:     commit.Hash(),
		DataHash:           commit.Hash(),
		ValidatorsHash:     commit.Hash(),
		NextValidatorsHash: commit.Hash(),
		ConsensusHash:      commit.Hash(),
		AppHash:            commit.Hash(),
		LastResultsHash:    commit.Hash(),
		ProposerAddress:    crypto.AddressFromPreimage([]byte("proposer_address")),
	}

	validSignedHeader := SignedHeader{Header: &h, Commit: commit}
	validSignedHeader.Commit.BlockID.Hash = validSignedHeader.Hash()
	invalidSignedHeader := SignedHeader{}

	testCases := []struct {
		testName  string
		shHeader  *Header
		shCommit  *Commit
		expectErr bool
	}{
		{"Valid Signed Header", validSignedHeader.Header, validSignedHeader.Commit, false},
		{"Invalid Signed Header", invalidSignedHeader.Header, validSignedHeader.Commit, true},
		{"Invalid Signed Header", validSignedHeader.Header, invalidSignedHeader.Commit, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			sh := SignedHeader{
				Header: tc.shHeader,
				Commit: tc.shCommit,
			}
			assert.Equal(t, tc.expectErr, sh.ValidateBasic(validSignedHeader.Header.ChainID) != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestBlockIDValidateBasic(t *testing.T) {
	t.Parallel()

	validBlockID := BlockID{
		Hash: []byte{},
		PartsHeader: PartSetHeader{
			Total: 1,
			Hash:  []byte{},
		},
	}

	invalidBlockID := BlockID{
		Hash: []byte{0},
		PartsHeader: PartSetHeader{
			Total: -1,
			Hash:  []byte{},
		},
	}

	testCases := []struct {
		testName           string
		blockIDHash        []byte
		blockIDPartsHeader PartSetHeader
		expectErr          bool
	}{
		{"Valid BlockID", validBlockID.Hash, validBlockID.PartsHeader, false},
		{"Invalid BlockID", invalidBlockID.Hash, validBlockID.PartsHeader, true},
		{"Invalid BlockID", validBlockID.Hash, invalidBlockID.PartsHeader, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			blockID := BlockID{
				Hash:        tc.blockIDHash,
				PartsHeader: tc.blockIDPartsHeader,
			}
			assert.Equal(t, tc.expectErr, blockID.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}
