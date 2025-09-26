package tendermint

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

// MsgHeader defines the Tendermint client consensus Header.
// It encapsulates all the information necessary to update from a trusted
// Tendermint ConsensusState. The inclusion of TrustedHeight and
// TrustedValidators allows this update to process correctly, so long as the
// ConsensusState for the TrustedHeight exists, this removes race conditions
// among relayers The SignedHeader and ValidatorSet are the new untrusted update
// fields for the client. The TrustedHeight is the height of a stored
// ConsensusState on the client that will be used to verify the new untrusted
// header. The Trusted ConsensusState must be within the unbonding period of
// current time in order to correctly verify, and the TrustedValidators must
// hash to TrustedConsensusState.NextValidatorsHash since that is the last
// trusted validator set at the TrustedHeight.
type MsgHeader struct {
	Header            *Header       `json:"header,omitempty"`
	Commit            *Commit       `json:"commit,omitempty"`
	ValidatorSet      *ValidatorSet `json:"validator_set,omitempty"`
	TrustedHeight     types.Height  `json:"trusted_height"`
	TrustedValidators *ValidatorSet `json:"trusted_validators,omitempty"`
}

// Implements lightclient.ClientMessage
func (*MsgHeader) ClientType() string {
	return lightclient.Tendermint
}

// Implements lightclient.ClientMessage
// NOTE: TrustedHeight and TrustedValidators may be empty when creating client
// with MsgCreateClient
func (h *MsgHeader) Validate() error {
	if h.Header == nil {
		return errors.New("missing header")
	}
	if h.Commit == nil {
		return errors.New("missing commit")
	}
	if err := h.Header.Validate(); err != nil {
		return ufmt.Errorf("invalid header: %v", err)
	}
	if err := h.Commit.Validate(); err != nil {
		return ufmt.Errorf("invalid commit: %v", err)
	}
	// Make sure the header is consistent with the commit.
	if h.Header.Height != h.Commit.Height {
		return ufmt.Errorf("header and commit height mismatch: %d vs %d", h.Header.Height, h.Commit.Height)
	}
	// TODO(tb): computing the block hash from header is difficult in Gno because
	// it requires to use protobuf and other merkle libraries, we need to ensure
	// that this check is really necessary in terms of security.
	// if hhash, chash := h.Header.Hash(), h.Commit.BlockID.Hash; !bytes.Equal(hhash, chash) {
	// cs, hs := hex.EncodeToString(chash), hex.EncodeToString(hhash)
	// return ufmt.Errorf("commit signs block %q, header is block %q", cs, hs)
	// }

	// TrustedHeight is less than Header for updates and misbehaviour
	if h.TrustedHeight.GTE(h.GetHeight()) {
		return ufmt.Errorf("trustedHeight %s must be less than header height %s",
			h.TrustedHeight, h.GetHeight())
	}

	if h.ValidatorSet == nil {
		return ufmt.Errorf("validator set is nil")
	}
	if h.TrustedValidators == nil {
		return ufmt.Errorf("trusted validator set is nil")
	}
	if !bytes.Equal(h.Header.ValidatorsHash, h.ValidatorSet.Hash()) {
		return ufmt.Errorf("validator set does not match hash")
	}
	if !bytes.Equal(h.Header.NextValidatorsHash, h.TrustedValidators.Hash()) {
		return ufmt.Errorf("trusted validator set does not match hash")
	}
	return nil
}

// GetHeight returns the current height. It returns 0 if the tendermint
// header is nil.
// NOTE: the header.Header is checked to be non nil in Validate.
func (h MsgHeader) GetHeight() types.Height {
	revision := ParseChainID(h.Header.ChainID)
	return types.NewHeight(revision, h.Header.Height)
}

// Misbehaviour is a wrapper over two conflicting Headers
// that implements Misbehaviour interface expected by ICS-02
type Misbehaviour struct {
	Header1 *MsgHeader `json:"header_1,omitempty"`
	Header2 *MsgHeader `json:"header_2,omitempty"`
}

// Implements lightclient.ClientMessage
func (Misbehaviour) ClientType() string {
	return lightclient.Tendermint
}

// Implements lightclient.ClientMessage
func (m Misbehaviour) Validate() error {
	// TODO
	return nil
}

type ValidatorSet struct {
	Validators       []*Validator `json:"validators,omitempty"`
	Proposer         *Validator   `json:"proposer,omitempty"`
	TotalVotingPower int64        `json:"total_voting_power,omitempty"`
}

func (vals ValidatorSet) Hash() []byte {
	var bz []byte
	for _, v := range vals.Validators {
		bz = append(bz, v.Bytes()...)
	}
	// TODO: ibc-go returns a merkle hash, ensure we can just use a simple hash
	// -> see p/demo/merkle
	// return merkle.HashFromByteSlices(bzs)
	h := sha256.Sum256(bz)
	return h[:]
}

// GetByAddress returns an index of the validator with address and validator
// itself (copy) if found. Otherwise, -1 and nil are returned.
func (vals *ValidatorSet) GetByAddress(address []byte) (index int32, val *Validator) {
	for idx, val := range vals.Validators {
		if bytes.Equal(val.Address, address) {
			return int32(idx), val
		}
	}
	return -1, nil
}

type Validator struct {
	Address          []byte `json:"address,omitempty"`
	PubKey           []byte `json:"pub_key"`
	VotingPower      int64  `json:"voting_power,omitempty"`
	ProposerPriority int64  `json:"proposer_priority,omitempty"`
}

// Bytes returns a bytes representation of v, composed as follow:
// 32bytes: Pubkey
// varint: VotingPower
func (v Validator) Bytes() (bz []byte) {
	bz = append(bz, v.PubKey...)
	bz = binary.AppendVarint(bz, v.VotingPower)
	return
}

// Header defines the structure of a block header.
// TODO remove json tags everywhere
type Header struct {
	// basic block info
	Version Consensus `json:"version"`
	ChainID string    `json:"chain_id,omitempty"`
	Height  uint64    `json:"height,omitempty"`
	Time    time.Time `json:"time"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	// commit info from validators from the last block
	LastCommitHash []byte `json:"last_commit_hash,omitempty"`
	// transactions
	DataHash []byte `json:"data_hash,omitempty"`

	// hashes from the app output form the preb block
	// validators for the current block
	ValidatorsHash []byte `json:"validators_hash,omitempty"`
	// validators for the next block
	NextValidatorsHash []byte `json:"next_validators_hash,omitempty"`
	// consensus params for the current block
	ConsensusHash []byte `json:"consensus_hash,omitempty"`
	// state afterx txs from the previous block
	AppHash []byte `json:"app_hash,omitempty"`
	// root hash of all results from the txs from the previous block
	LastResultsHash []byte `json:"last_results_hash,omitempty"`

	// consensus info
	// evidence included in the block
	EvidenceHash []byte `json:"evidence_hash,omitempty"`
	//original proposer of the block
	ProposerAddress []byte `json:"proposer_address,omitempty"`
}

const (
	// BlockProtocol versions all block data structures and processing.
	// This includes validity of blocks and state updates.
	BlockProtocol uint64 = 11
	AddressSize          = 20
	// MaxSignatureSize is a maximum allowed signature size for the Proposal
	// and Vote.
	// XXX: secp256k1 does not have Size nor MaxSize defined.
	MaxSignatureSize = 64
)

// Validate performs stateless validation on a Header returning an error if any
// validation fails.
//
// NOTE: Timestamp validation is subtle and handled elsewhere.
func (h Header) Validate() error {
	if h.Version.Block != BlockProtocol {
		return ufmt.Errorf("block protocol is incorrect: got: %d, want: %d ", h.Version.Block, BlockProtocol)
	}
	if len(h.ChainID) > MaxChainIDLen {
		return ufmt.Errorf("chainID is too long; got: %d, max: %d", len(h.ChainID), MaxChainIDLen)
	}

	if h.Height == 0 {
		return errors.New("zero Height")
	}

	if err := h.LastBlockID.Validate(); err != nil {
		return ufmt.Errorf("wrong LastBlockID: %v", err)
	}

	if err := validateHash(h.LastCommitHash); err != nil {
		return ufmt.Errorf("wrong LastCommitHash: %v", err)
	}

	if err := validateHash(h.DataHash); err != nil {
		return ufmt.Errorf("wrong DataHash: %v", err)
	}

	if err := validateHash(h.EvidenceHash); err != nil {
		return ufmt.Errorf("wrong EvidenceHash: %v", err)
	}

	if len(h.ProposerAddress) != AddressSize {
		return ufmt.Errorf(
			"invalid ProposerAddress length; got: %d, expected: %d",
			len(h.ProposerAddress), AddressSize,
		)
	}

	// Basic validation of hashes related to application data.
	// Will validate fully against state in state#ValidateBlock.
	if err := validateHash(h.ValidatorsHash); err != nil {
		return ufmt.Errorf("wrong ValidatorsHash: %v", err)
	}
	if err := validateHash(h.NextValidatorsHash); err != nil {
		return ufmt.Errorf("wrong NextValidatorsHash: %v", err)
	}
	if err := validateHash(h.ConsensusHash); err != nil {
		return ufmt.Errorf("wrong ConsensusHash: %v", err)
	}
	// NOTE: AppHash is arbitrary length
	if err := validateHash(h.LastResultsHash); err != nil {
		return ufmt.Errorf("wrong LastResultsHash: %v", err)
	}

	return nil
}

func (h *Header) Hash() []byte {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	// TODO Use a source of bytes that is not protobuf
	return nil
}

// Commit contains the evidence that a block was committed by a set of validators.
type Commit struct {
	// TODO revert to int64
	Height     uint64      `json:"height,omitempty"`
	Round      int32       `json:"round,omitempty"`
	BlockID    BlockID     `json:"block_id"`
	Signatures []CommitSig `json:"signatures"`
}

// Validate performs basic validation that doesn't involve state data.
// Does not actually check the cryptographic signatures.
func (commit *Commit) Validate() error {
	if commit.Round < 0 {
		return errors.New("negative Round")
	}

	if commit.Height >= 1 {
		if commit.BlockID.IsZero() {
			return errors.New("commit cannot be for nil block")
		}

		if len(commit.Signatures) == 0 {
			return errors.New("no signatures in commit")
		}
		for i, commitSig := range commit.Signatures {
			if err := commitSig.Validate(); err != nil {
				return ufmt.Errorf("wrong CommitSig #%d: %v", i, err)
			}
		}
	}
	return nil
}

// CommitSig is a part of the Vote included in a Commit.
type CommitSig struct {
	BlockIDFlag      BlockIDFlag `json:"block_id_flag,omitempty"`
	ValidatorAddress []byte      `json:"validator_address,omitempty"`
	Timestamp        time.Time   `json:"timestamp"`
	Signature        []byte      `json:"signature,omitempty"`
}

// Validate performs basic validation.
func (cs CommitSig) Validate() error {
	switch cs.BlockIDFlag {
	case BlockIDFlagAbsent:
	case BlockIDFlagCommit:
	case BlockIDFlagNil:
	default:
		return ufmt.Errorf("unknown BlockIDFlag: %v", cs.BlockIDFlag)
	}

	switch cs.BlockIDFlag {
	case BlockIDFlagAbsent:
		if len(cs.ValidatorAddress) != 0 {
			return errors.New("validator address is present")
		}
		if !cs.Timestamp.IsZero() {
			return errors.New("time is present")
		}
		if len(cs.Signature) != 0 {
			return errors.New("signature is present")
		}
	default:
		if len(cs.ValidatorAddress) != AddressSize {
			return ufmt.Errorf("expected ValidatorAddress size to be %d bytes, got %d bytes",
				AddressSize,
				len(cs.ValidatorAddress),
			)
		}
		// NOTE: Timestamp validation is subtle and handled elsewhere.
		if len(cs.Signature) == 0 {
			return errors.New("signature is missing")
		}
		if len(cs.Signature) > MaxSignatureSize {
			return ufmt.Errorf("signature is too big (max: %d)", MaxSignatureSize)
		}
	}

	return nil
}

// BlockIdFlag indicates which BlockID the signature is for
type BlockIDFlag byte

const (
	BlockIDFlagUnknown BlockIDFlag = 0
	BlockIDFlagAbsent  BlockIDFlag = 1
	BlockIDFlagCommit  BlockIDFlag = 2
	BlockIDFlagNil     BlockIDFlag = 3
)

// BlockID
type BlockID struct {
	Hash          []byte        `json:"hash,omitempty"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}

// Equals returns true if the BlockID matches the given BlockID
func (blockID BlockID) Equals(other BlockID) bool {
	return bytes.Equal(blockID.Hash, other.Hash) &&
		blockID.PartSetHeader.Equals(other.PartSetHeader)
}

// IsZero returns true if this is the BlockID of a nil block.
func (blockID BlockID) IsZero() bool {
	return len(blockID.Hash) == 0 &&
		blockID.PartSetHeader.IsZero()
}

// Validate performs basic validation.
func (blockID BlockID) Validate() error {
	// Hash can be empty in case of POLBlockID in Proposal.
	if err := validateHash(blockID.Hash); err != nil {
		return ufmt.Errorf("wrong blockID Hash: %v", err)
	}
	if err := blockID.PartSetHeader.Validate(); err != nil {
		return ufmt.Errorf("wrong PartSetHeader: %v", err)
	}
	return nil
}

// PartsetHeader
type PartSetHeader struct {
	Total uint32 `json:"total,omitempty"`
	Hash  []byte `json:"hash,omitempty"`
}

func (psh PartSetHeader) IsZero() bool {
	return psh.Total == 0 && len(psh.Hash) == 0
}

func (psh PartSetHeader) Equals(other PartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash)
}

// Validate performs basic validation.
func (psh PartSetHeader) Validate() error {
	// Hash can be empty in case of POLBlockID.PartSetHeader in Proposal.
	if err := validateHash(psh.Hash); err != nil {
		return ufmt.Errorf("wrong parSetHeader Hash: %v", err)
	}
	return nil
}

// Consensus captures the consensus rules for processing a block in the blockchain,
// including all blockchain data structures and the rules of the application's
// state transition machine.
type Consensus struct {
	Block uint64 `json:"block,omitempty"`
	App   uint64 `json:"app,omitempty"`
}
