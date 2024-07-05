package types

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	typesver "github.com/gnolang/gno/tm2/pkg/bft/types/version"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// Block defines the atomic unit of a Tendermint blockchain.
type Block struct {
	mtx        sync.Mutex
	Header     `json:"header"`
	Data       `json:"data"`
	LastCommit *Commit `json:"last_commit"`
}

// ValidateBasic performs basic validation that doesn't involve state data.
// It checks the internal consistency of the block.
// Further validation is done using state#ValidateBlock.
func (b *Block) ValidateBasic() error {
	if b == nil {
		return errors.New("nil block")
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if len(b.ChainID) > MaxChainIDLen {
		return fmt.Errorf("ChainID is too long. Max is %d, got %d", MaxChainIDLen, len(b.ChainID))
	}

	if b.Height < 0 {
		return errors.New("Negative Header.Height")
	} else if b.Height == 0 {
		return errors.New("Zero Header.Height")
	}

	// NOTE: Timestamp validation is subtle and handled elsewhere.

	newTxs := int64(len(b.Data.Txs))
	if b.NumTxs != newTxs {
		return fmt.Errorf("wrong Header.NumTxs. Expected %v, got %v",
			newTxs,
			b.NumTxs,
		)
	}

	// TODO: fix tests so we can do this
	/*if b.TotalTxs < b.NumTxs {
		return fmt.Errorf("Header.TotalTxs (%d) is less than Header.NumTxs (%d)", b.TotalTxs, b.NumTxs)
	}*/
	if b.TotalTxs < 0 {
		return errors.New("Negative Header.TotalTxs")
	}

	if err := b.LastBlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong Header.LastBlockID: %w", err)
	}

	// Validate the last commit and its hash.
	if b.Header.Height > 1 {
		if b.LastCommit == nil {
			return errors.New("nil LastCommit")
		}
		if err := b.LastCommit.ValidateBasic(); err != nil {
			return fmt.Errorf("wrong LastCommit")
		}
	}
	if err := ValidateHash(b.LastCommitHash); err != nil {
		return fmt.Errorf("wrong Header.LastCommitHash: %w", err)
	}
	if !bytes.Equal(b.LastCommitHash, b.LastCommit.Hash()) {
		return fmt.Errorf("wrong Header.LastCommitHash. Expected %v, got %v",
			b.LastCommit.Hash(),
			b.LastCommitHash,
		)
	}

	// Validate the hash of the transactions.
	// NOTE: b.Data.Txs may be nil, but b.Data.Hash()
	// still works fine
	if err := ValidateHash(b.DataHash); err != nil {
		return fmt.Errorf("wrong Header.DataHash: %w", err)
	}
	if !bytes.Equal(b.DataHash, b.Data.Hash()) {
		return fmt.Errorf(
			"wrong Header.DataHash. Expected %v, got %v",
			b.Data.Hash(),
			b.DataHash,
		)
	}

	// Basic validation of hashes related to application data.
	// Will validate fully against state in state#ValidateBlock.
	if err := ValidateHash(b.ValidatorsHash); err != nil {
		return fmt.Errorf("wrong Header.ValidatorsHash: %w", err)
	}
	if err := ValidateHash(b.NextValidatorsHash); err != nil {
		return fmt.Errorf("wrong Header.NextValidatorsHash: %w", err)
	}
	if err := ValidateHash(b.ConsensusHash); err != nil {
		return fmt.Errorf("wrong Header.ConsensusHash: %w", err)
	}
	// NOTE: AppHash is arbitrary length
	if err := ValidateHash(b.LastResultsHash); err != nil {
		return fmt.Errorf("wrong Header.LastResultsHash: %w", err)
	}

	if len(b.ProposerAddress) != crypto.AddressSize {
		return fmt.Errorf("expected len(Header.ProposerAddress) to be %d, got %d",
			crypto.AddressSize, len(b.ProposerAddress))
	}

	return nil
}

// fillHeader fills in any remaining header fields that are a function of the block data
func (b *Block) fillHeader() {
	if b.LastCommitHash == nil {
		b.LastCommitHash = b.LastCommit.Hash()
	}
	if b.DataHash == nil {
		b.DataHash = b.Data.Hash()
	}
}

// Hash computes and returns the block hash.
// If the block is incomplete, block hash is nil for safety.
func (b *Block) Hash() []byte {
	if b == nil {
		return nil
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if b.LastCommit == nil {
		return nil
	}
	b.fillHeader()
	return b.Header.Hash()
}

// MakePartSet returns a PartSet containing parts of a serialized block.
// This is the form in which the block is gossipped to peers.
// CONTRACT: partSize is greater than zero.
func (b *Block) MakePartSet(partSize int) *PartSet {
	if b == nil {
		return nil
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// We prefix the byte length, so that unmarshaling
	// can easily happen via a reader.
	bz, err := amino.MarshalSized(b)
	if err != nil {
		panic(err)
	}
	return NewPartSetFromData(bz, partSize)
}

// HashesTo is a convenience function that checks if a block hashes to the given argument.
// Returns false if the block is nil or the hash is empty.
func (b *Block) HashesTo(hash []byte) bool {
	if len(hash) == 0 {
		return false
	}
	if b == nil {
		return false
	}
	return bytes.Equal(b.Hash(), hash)
}

// Size returns size of the block in bytes.
func (b *Block) Size() int {
	bz, err := amino.Marshal(b)
	if err != nil {
		return 0
	}
	return len(bz)
}

// String returns a string representation of the block
func (b *Block) String() string {
	return b.StringIndented("")
}

// StringIndented returns a string representation of the block
func (b *Block) StringIndented(indent string) string {
	if b == nil {
		return "nil-Block"
	}
	return fmt.Sprintf(`Block{
%s  %v
%s  %v
%s  %v
%s}#%v`,
		indent, b.Header.StringIndented(indent+"  "),
		indent, b.Data.StringIndented(indent+"  "),
		indent, b.LastCommit.StringIndented(indent+"  "),
		indent, b.Hash())
}

// StringShort returns a shortened string representation of the block
func (b *Block) StringShort() string {
	if b == nil {
		return "nil-Block"
	}
	return fmt.Sprintf("Block#%v", b.Hash())
}

//-----------------------------------------------------------------------------

// Header defines the structure of a Tendermint block header.
// NOTE: changes to the Header should be duplicated in:
// - header.Hash()
// - abci.Header
// - /docs/spec/blockchain/blockchain.md
type Header struct {
	// basic block info
	Version    string    `json:"version"`
	ChainID    string    `json:"chain_id"`
	Height     int64     `json:"height"`
	Time       time.Time `json:"time"`
	NumTxs     int64     `json:"num_txs"`
	TotalTxs   int64     `json:"total_txs"`
	AppVersion string    `json:"app_version"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash []byte `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       []byte `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     []byte `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash []byte `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      []byte `json:"consensus_hash"`       // consensus params for current block
	AppHash            []byte `json:"app_hash"`             // state after txs from the previous block
	LastResultsHash    []byte `json:"last_results_hash"`    // root hash of all results from the txs from the previous block

	// consensus info
	ProposerAddress Address `json:"proposer_address"` // original proposer of the block
}

// Implements abci.Header
func (h *Header) AssertABCIHeader()  {}
func (h *Header) GetChainID() string { return h.ChainID }
func (h *Header) GetHeight() int64   { return h.Height }
func (h *Header) GetTime() time.Time { return h.Time }

// MakeBlock returns a new block with an empty header, except what can be
// computed from itself.
// It populates the same set of fields validated by ValidateBasic.
func MakeBlock(height int64, txs []Tx, lastCommit *Commit) *Block {
	block := &Block{
		Header: Header{
			Height: height,
			NumTxs: int64(len(txs)),
		},
		Data: Data{
			Txs: txs,
		},
		LastCommit: lastCommit,
	}
	block.fillHeader()
	return block
}

func (h *Header) Copy() *Header {
	return amino.DeepCopy(h).(*Header)
}

// Populate the Header with state-derived data.
// Call this after MakeBlock to complete the Header.
func (h *Header) Populate(
	chainID string,
	timestamp time.Time, lastBlockID BlockID, totalTxs int64,
	appVersion string,
	valHash, nextValHash []byte,
	consensusHash, appHash, lastResultsHash []byte,
	proposerAddress Address,
) {
	h.Version = typesver.BlockVersion
	h.ChainID = chainID
	h.Time = timestamp
	h.LastBlockID = lastBlockID
	h.TotalTxs = totalTxs
	h.AppVersion = appVersion
	h.ValidatorsHash = valHash
	h.NextValidatorsHash = nextValHash
	h.ConsensusHash = consensusHash
	h.AppHash = appHash
	h.LastResultsHash = lastResultsHash
	h.ProposerAddress = proposerAddress
}

// Hash returns the hash of the header.
// It computes a Merkle tree from the header fields
// ordered as they appear in the Header.
// Returns nil if ValidatorHash is missing,
// since a Header is not valid unless there is
// a ValidatorsHash (corresponding to the validator set).
func (h *Header) Hash() []byte {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	return merkle.SimpleHashFromByteSlices([][]byte{
		bytesOrNil(h.Version),
		bytesOrNil(h.ChainID),
		bytesOrNil(h.Height),
		bytesOrNil(h.Time),
		bytesOrNil(h.NumTxs),
		bytesOrNil(h.TotalTxs),
		bytesOrNil(h.AppVersion),
		bytesOrNil(h.LastBlockID),
		bytesOrNil(h.LastCommitHash),
		bytesOrNil(h.DataHash),
		bytesOrNil(h.ValidatorsHash),
		bytesOrNil(h.NextValidatorsHash),
		bytesOrNil(h.ConsensusHash),
		bytesOrNil(h.AppHash),
		bytesOrNil(h.LastResultsHash),
		bytesOrNil(h.ProposerAddress),
	})
}

// StringIndented returns a string representation of the header
func (h *Header) StringIndented(indent string) string {
	if h == nil {
		return "nil-Header"
	}
	return fmt.Sprintf(`Header{
%s  Version:        %v
%s  ChainID:        %v
%s  Height:         %v
%s  Time:           %v
%s  NumTxs:         %v
%s  TotalTxs:       %v
%s  AppVersion:     %v
%s  LastBlockID:    %v
%s  LastCommit:     %v
%s  Data:           %v
%s  Validators:     %v
%s  NextValidators: %v
%s  App:            %v
%s  Consensus:      %v
%s  Results:        %v
%s  Proposer:       %v
%s}#%v`,
		indent, h.Version,
		indent, h.ChainID,
		indent, h.Height,
		indent, h.Time,
		indent, h.NumTxs,
		indent, h.TotalTxs,
		indent, h.AppVersion,
		indent, h.LastBlockID,
		indent, h.LastCommitHash,
		indent, h.DataHash,
		indent, h.ValidatorsHash,
		indent, h.NextValidatorsHash,
		indent, h.AppHash,
		indent, h.ConsensusHash,
		indent, h.LastResultsHash,
		indent, h.ProposerAddress,
		indent, h.Hash())
}

//-------------------------------------

// CommitSig is a vote included in a Commit.
// For now, it is identical to a vote,
// but in the future it will contain fewer fields
// to eliminate the redundancy in commits.
// See https://github.com/tendermint/classic/issues/1648.
type CommitSig Vote

// String returns the underlying Vote.String()
func (cs *CommitSig) String() string {
	return cs.toVote().String()
}

// toVote converts the CommitSig to a vote.
// TODO: deprecate for #1648. Converting to Vote will require
// access to ValidatorSet.
func (cs *CommitSig) toVote() *Vote {
	if cs == nil {
		return nil
	}
	v := Vote(*cs)
	return &v
}

//-------------------------------------

// Commit contains the evidence that a block was committed by a set of validators.
// NOTE: Commit is empty for height 1, but never nil.
type Commit struct {
	// NOTE: The Precommits are in order of address to preserve the bonded ValidatorSet order.
	// Any peer with a block can gossip precommits by index with a peer without recalculating the
	// active ValidatorSet.
	BlockID    BlockID      `json:"block_id"`
	Precommits []*CommitSig `json:"precommits" amino:"nil_elements"`

	// memoized in first call to corresponding method
	// NOTE: can't memoize in constructor because constructor
	// isn't used for unmarshaling
	height   int64
	round    int
	hash     []byte
	bitArray *bitarray.BitArray
}

// NewCommit returns a new Commit with the given blockID and precommits.
// TODO: memoize ValidatorSet in constructor so votes can be easily reconstructed
// from CommitSig after #1648.
func NewCommit(blockID BlockID, precommits []*CommitSig) *Commit {
	return &Commit{
		BlockID:    blockID,
		Precommits: precommits,
	}
}

// Construct a VoteSet from the Commit and validator set. Panics
// if precommits from the commit can't be added to the voteset.
// Inverse of VoteSet.MakeCommit().
func CommitToVoteSet(chainID string, commit *Commit, vals *ValidatorSet) *VoteSet {
	height, round, typ := commit.Height(), commit.Round(), PrecommitType
	voteSet := NewVoteSet(chainID, height, round, typ, vals)
	for idx, precommit := range commit.Precommits {
		if precommit == nil {
			continue
		}
		added, err := voteSet.AddVote(commit.GetVote(idx))
		if !added || err != nil {
			panic(fmt.Sprintf("Failed to reconstruct LastCommit: %v", err))
		}
	}
	return voteSet
}

// GetVote converts the CommitSig for the given valIdx to a Vote.
// Returns nil if the precommit at valIdx is nil.
// Panics if valIdx >= commit.Size().
func (commit *Commit) GetVote(valIdx int) *Vote {
	commitSig := commit.Precommits[valIdx]
	if commitSig == nil {
		return nil
	}

	// NOTE: this commitSig might be for a nil blockID,
	// so we can't just use commit.BlockID here.
	// For #1648, CommitSig will need to indicate what BlockID it's for !
	blockID := commitSig.BlockID
	commit.memoizeHeightRound()
	return &Vote{
		Type:             PrecommitType,
		Height:           commit.height,
		Round:            commit.round,
		BlockID:          blockID,
		Timestamp:        commitSig.Timestamp,
		ValidatorAddress: commitSig.ValidatorAddress,
		ValidatorIndex:   valIdx,
		Signature:        commitSig.Signature,
	}
}

// VoteSignBytes constructs the SignBytes for the given CommitSig.
// The only unique part of the SignBytes is the Timestamp - all other fields
// signed over are otherwise the same for all validators.
// Panics if valIdx >= commit.Size().
func (commit *Commit) VoteSignBytes(chainID string, valIdx int) []byte {
	return commit.GetVote(valIdx).SignBytes(chainID)
}

// memoizeHeightRound memoizes the height and round of the commit using
// the first non-nil vote.
// Should be called before any attempt to access `commit.height` or `commit.round`.
func (commit *Commit) memoizeHeightRound() {
	if len(commit.Precommits) == 0 {
		return
	}
	if commit.height > 0 {
		return
	}
	for _, precommit := range commit.Precommits {
		if precommit != nil {
			commit.height = precommit.Height
			commit.round = precommit.Round
			return
		}
	}
}

// Height returns the height of the commit
func (commit *Commit) Height() int64 {
	commit.memoizeHeightRound()
	return commit.height
}

// Round returns the round of the commit
func (commit *Commit) Round() int {
	commit.memoizeHeightRound()
	return commit.round
}

// Type returns the vote type of the commit, which is always VoteTypePrecommit
func (commit *Commit) Type() byte {
	return byte(PrecommitType)
}

// Size returns the number of votes in the commit
func (commit *Commit) Size() int {
	if commit == nil {
		return 0
	}
	return len(commit.Precommits)
}

// BitArray returns a BitArray of which validators voted in this commit
func (commit *Commit) BitArray() *bitarray.BitArray {
	if commit.bitArray == nil {
		commit.bitArray = bitarray.NewBitArray(len(commit.Precommits))
		for i, precommit := range commit.Precommits {
			// TODO: need to check the BlockID otherwise we could be counting conflicts,
			// not just the one with +2/3 !
			commit.bitArray.SetIndex(i, precommit != nil)
		}
	}
	return commit.bitArray
}

// GetByIndex returns the vote corresponding to a given validator index.
// Panics if `index >= commit.Size()`.
// Implements VoteSetReader.
func (commit *Commit) GetByIndex(valIdx int) *Vote {
	return commit.GetVote(valIdx)
}

// IsCommit returns true if there is at least one vote.
func (commit *Commit) IsCommit() bool {
	return len(commit.Precommits) != 0
}

// ValidateBasic performs basic validation that doesn't involve state data.
// Does not actually check the cryptographic signatures.
func (commit *Commit) ValidateBasic() error {
	if commit.BlockID.IsZero() {
		return errors.New("Commit cannot be for nil block")
	}
	if len(commit.Precommits) == 0 {
		return errors.New("No precommits in commit")
	}
	height, round := commit.Height(), commit.Round()

	// Validate the precommits.
	for _, precommit := range commit.Precommits {
		// It's OK for precommits to be missing.
		if precommit == nil {
			continue
		}
		// Ensure that all votes are precommits.
		if precommit.Type != PrecommitType {
			return fmt.Errorf("invalid commit vote. Expected precommit, got %v",
				precommit.Type)
		}
		// Ensure that all heights are the same.
		if precommit.Height != height {
			return fmt.Errorf("invalid commit precommit height. Expected %v, got %v",
				height, precommit.Height)
		}
		// Ensure that all rounds are the same.
		if precommit.Round != round {
			return fmt.Errorf("invalid commit precommit round. Expected %v, got %v",
				round, precommit.Round)
		}
	}
	return nil
}

// Hash returns the hash of the commit
func (commit *Commit) Hash() []byte {
	if commit == nil {
		return nil
	}
	if commit.hash == nil {
		bs := make([][]byte, len(commit.Precommits))
		for i, precommit := range commit.Precommits {
			bs[i] = bytesOrNil(precommit)
		}
		commit.hash = merkle.SimpleHashFromByteSlices(bs)
	}
	return commit.hash
}

// StringIndented returns a string representation of the commit
func (commit *Commit) StringIndented(indent string) string {
	if commit == nil {
		return "nil-Commit"
	}
	precommitStrings := make([]string, len(commit.Precommits))
	for i, precommit := range commit.Precommits {
		precommitStrings[i] = precommit.String()
	}
	return fmt.Sprintf(`Commit{
%s  BlockID:    %v
%s  Precommits:
%s    %v
%s}#%v`,
		indent, commit.BlockID,
		indent,
		indent, strings.Join(precommitStrings, "\n"+indent+"    "),
		indent, commit.hash)
}

//-----------------------------------------------------------------------------

// SignedHeader is a header along with the commits that prove it.
// It is the basis of the lite client.
type SignedHeader struct {
	*Header `json:"header"`
	Commit  *Commit `json:"commit"`
}

// ValidateBasic does basic consistency checks and makes sure the header
// and commit are consistent.
//
// NOTE: This does not actually check the cryptographic signatures.  Make
// sure to use a Verifier to validate the signatures actually provide a
// significantly strong proof for this header's validity.
func (sh SignedHeader) ValidateBasic(chainID string) error {
	// Make sure the header is consistent with the commit.
	if sh.Header == nil {
		return errors.New("SignedHeader missing header.")
	}
	if sh.Commit == nil {
		return errors.New("SignedHeader missing commit (precommit votes).")
	}

	// Check ChainID.
	if sh.ChainID != chainID {
		return fmt.Errorf("Header belongs to another chain '%s' not '%s'",
			sh.ChainID, chainID)
	}
	// Check Height.
	if sh.Commit.Height() != sh.Height {
		return fmt.Errorf("SignedHeader header and commit height mismatch: %v vs %v",
			sh.Height, sh.Commit.Height())
	}
	// Check Hash.
	hhash := sh.Hash()
	chash := sh.Commit.BlockID.Hash
	if !bytes.Equal(hhash, chash) {
		return fmt.Errorf("SignedHeader commit signs block %X, header is block %X",
			chash, hhash)
	}
	// ValidateBasic on the Commit.
	err := sh.Commit.ValidateBasic()
	if err != nil {
		return errors.Wrap(err, "commit.ValidateBasic failed during SignedHeader.ValidateBasic")
	}
	return nil
}

func (sh SignedHeader) String() string {
	return sh.StringIndented("")
}

// StringIndented returns a string representation of the SignedHeader.
func (sh SignedHeader) StringIndented(indent string) string {
	return fmt.Sprintf(`SignedHeader{
%s  %v
%s  %v
%s}`,
		indent, sh.Header.StringIndented(indent+"  "),
		indent, sh.Commit.StringIndented(indent+"  "),
		indent)
}

//-----------------------------------------------------------------------------

// Data contains the set of transactions included in the block
type Data struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs Txs `json:"txs"`

	// Volatile
	hash []byte
}

// Hash returns the hash of the data
func (data *Data) Hash() []byte {
	if data == nil {
		return (Txs{}).Hash()
	}
	if data.hash == nil {
		data.hash = data.Txs.Hash() // NOTE: leaves of merkle tree are TxIDs
	}
	return data.hash
}

// StringIndented returns a string representation of the transactions
func (data *Data) StringIndented(indent string) string {
	if data == nil {
		return "nil-Data"
	}
	txStrings := make([]string, min(len(data.Txs), 21))
	for i, tx := range data.Txs {
		if i == 20 {
			txStrings[i] = fmt.Sprintf("... (%v total)", len(data.Txs))
			break
		}
		txStrings[i] = fmt.Sprintf("%X (%d bytes)", tx.Hash(), len(tx))
	}
	return fmt.Sprintf(`Data{
%s  %v
%s}#%v`,
		indent, strings.Join(txStrings, "\n"+indent+"  "),
		indent, data.hash)
}

//--------------------------------------------------------------------------------

// BlockID defines the unique ID of a block as its Hash and its PartSetHeader
type BlockID struct {
	Hash        []byte        `json:"hash"`
	PartsHeader PartSetHeader `json:"parts"`
}

// Equals returns true if the BlockID matches the given BlockID
func (blockID BlockID) Equals(other BlockID) bool {
	return bytes.Equal(blockID.Hash, other.Hash) &&
		blockID.PartsHeader.Equals(other.PartsHeader)
}

// Key returns a machine-readable string representation of the BlockID
func (blockID BlockID) Key() string {
	bz, err := amino.Marshal(blockID.PartsHeader)
	if err != nil {
		panic(err)
	}
	return string(blockID.Hash) + string(bz)
}

// ValidateBasic performs basic validation.
func (blockID BlockID) ValidateBasic() error {
	// Hash can be empty in case of POLBlockID in Proposal.
	if err := ValidateHash(blockID.Hash); err != nil {
		return fmt.Errorf("wrong Hash")
	}
	if err := blockID.PartsHeader.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong PartsHeader: %w", err)
	}
	return nil
}

// IsZero returns true if this is the BlockID of a nil block.
func (blockID BlockID) IsZero() bool {
	return len(blockID.Hash) == 0 &&
		blockID.PartsHeader.IsZero()
}

// IsComplete returns true if this is a valid BlockID of a non-nil block.
func (blockID BlockID) IsComplete() bool {
	return len(blockID.Hash) == tmhash.Size &&
		blockID.PartsHeader.Total > 0 &&
		len(blockID.PartsHeader.Hash) == tmhash.Size
}

// String returns a human readable string representation of the BlockID
func (blockID BlockID) String() string {
	return fmt.Sprintf(`%X:%v`, blockID.Hash, blockID.PartsHeader)
}
