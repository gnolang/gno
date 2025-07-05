package state

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// Step is the step in the consensus process.
type Step uint8

const (
	StepPropose   Step = 1 // Only used in proposal signing
	StepPrevote   Step = 2 // Only used in vote signing
	StepPrecommit Step = 3 // Only used in vote signing
)

// A vote type is either stepPrevote or stepPrecommit.
func VoteTypeToStep(voteType types.SignedMsgType) Step {
	switch voteType {
	case types.PrevoteType:
		return StepPrevote
	case types.PrecommitType:
		return StepPrecommit
	default:
		panic("Unknown vote type")
	}
}

// FileState stores the state of the last signing operation in a file.
// NOTE: keep in sync with gno.land/cmd/gnoland/secrets.go
// NOTE: this was migrated from tm2/pkg/bft/privval/file.go
type FileState struct {
	Height    int64  `json:"height" comment:"the height of the last sign"`
	Round     int    `json:"round" comment:"the round of the last sign"`
	Step      Step   `json:"step" comment:"the step of the last sign"`
	SignBytes []byte `json:"signbytes,omitempty" comment:"the raw signature bytes of the last sign"`
	Signature []byte `json:"signature,omitempty" comment:"the signature of the last sign"`

	filePath string
}

// FileState type implements fmt.Stringer.
var _ fmt.Stringer = (*FileState)(nil)

// String implements fmt.Stringer.
func (fs *FileState) String() string {
	return fmt.Sprintf("{H: %d, R: %d, S: %d}", fs.Height, fs.Round, fs.Step)
}

// FileState HRS checking errors.
var (
	errHeightRegression = errors.New("height regression")
	errRoundRegression  = errors.New("round regression")
	errStepRegression   = errors.New("step regression")
	errNoSignBytes      = errors.New("no SignBytes set")
)

// checkHRS checks the given height, round, step (HRS) against the last state.
// It returns an error if the arguments constitute a regression, or if HRS match but
// the SignBytes are not set.
// The returned boolean indicates whether the last Signature should be reused or not.
// It will be true if the HRS match and the SignBytes and Signature are already set
// in the last state (indicating we have already signed for this HRS).
func (fs *FileState) CheckHRS(height int64, round int, step Step) (bool, error) {
	// Check if the height differs.
	if height < fs.Height { // Height regression
		return false, fmt.Errorf("%w: expected >= %d, got %d", errHeightRegression, fs.Height, height)
	}
	if height > fs.Height { // New height, we can't reuse the signature.
		return false, nil
	}

	// Height is the same, now check if the round differs.
	if round < fs.Round { // Round regression
		return false, fmt.Errorf("%w: expected >= %d, got %d (height %d)", errRoundRegression, fs.Round, round, height)
	}
	if round > fs.Round { // New round, we can't reuse the signature.
		return false, nil
	}

	// Round is the same, now check if the step differs.
	if step < fs.Step { // Step regression
		return false, fmt.Errorf("%w: expected >= %d, got %d (height %d, round %d)", errStepRegression, fs.Step, step, height, round)
	}
	if step > fs.Step { // New step, we can't reuse the signature.
		return false, nil
	}

	// If the HRS are the same, the SignBytes should be already set.
	if fs.SignBytes == nil {
		return false, errNoSignBytes
	}

	// If the SignBytes are set, the Signature should be set as well.
	if fs.Signature == nil {
		panic("FileState: Signature is nil but SignBytes is not!")
	}

	// Everything matches, we can reuse the signature.
	return true, nil
}

// checkVotesOnlyDifferByTimestamp returns the timestamp from the last state SignBytes
// and a boolean indicating if the only difference in the votes is their timestamp.
func (fs *FileState) CheckVotesOnlyDifferByTimestamp(signBytes []byte) (time.Time, bool) {
	// Unmarshal the last and new votes.
	var lastVote, newVote types.CanonicalVote
	if err := amino.UnmarshalSized(fs.SignBytes, &lastVote); err != nil {
		panic(fmt.Sprintf("state signBytes cannot be unmarshalled into vote: %v", err))
	}
	if err := amino.UnmarshalSized(signBytes, &newVote); err != nil {
		panic(fmt.Sprintf("parameter signBytes cannot be unmarshalled into vote: %v", err))
	}

	// Save the last timestamp before modifying the vote.
	lastTime := lastVote.Timestamp

	// Set the times to the same value then remarshal to check equality.
	// If the only difference is the timestamp, the marshalled bytes should be equal.
	now := tmtime.Now()
	lastVote.Timestamp = now
	newVote.Timestamp = now
	lastVoteBytes, _ := amino.MarshalJSON(lastVote)
	newVoteBytes, _ := amino.MarshalJSON(newVote)

	return lastTime, bytes.Equal(newVoteBytes, lastVoteBytes)
}

// checkProposalsOnlyDifferByTimestamp returns the timestamp from the last state SignBytes
// and a boolean indicating if the only difference in the proposals is their timestamp.
func (fs *FileState) CheckProposalsOnlyDifferByTimestamp(signBytes []byte) (time.Time, bool) {
	// Unmarshal the last and new proposals.
	var lastProposal, newProposal types.CanonicalProposal
	if err := amino.UnmarshalSized(fs.SignBytes, &lastProposal); err != nil {
		panic(fmt.Sprintf("state signBytes cannot be unmarshalled into proposal: %v", err))
	}
	if err := amino.UnmarshalSized(signBytes, &newProposal); err != nil {
		panic(fmt.Sprintf("parameter signBytes cannot be unmarshalled into proposal: %v", err))
	}

	// Save the last timestamp before modifying the proposal.
	lastTime := lastProposal.Timestamp

	// Set the times to the same value then remarshal to check equality.
	// If the only difference is the timestamp, the marshalled bytes should be equal.
	now := tmtime.Now()
	lastProposal.Timestamp = now
	newProposal.Timestamp = now
	lastProposalBytes, _ := amino.MarshalSized(lastProposal)
	newProposalBytes, _ := amino.MarshalSized(newProposal)

	return lastTime, bytes.Equal(newProposalBytes, lastProposalBytes)
}

// update updates the FileState then persists it to disk.
func (fs *FileState) Update(height int64, round int, step Step, signBytes, signature []byte) error {
	fs.Height = height
	fs.Round = round
	fs.Step = step
	fs.SignBytes = signBytes
	fs.Signature = signature

	return fs.save()
}

// FileState validation errors.
var (
	errInvalidSignStateStep      = errors.New("invalid sign state step value")
	errInvalidSignStateHeight    = errors.New("invalid sign state height value")
	errInvalidSignStateRound     = errors.New("invalid sign state round value")
	errInvalidSignStateSignBytes = errors.New("invalid sign state sign bytes")
	errSignatureShouldNotBeSet   = errors.New("signature should not be set")
	errFilePathNotSet            = errors.New("filePath not set")
)

// validate validates the FileState.
func (fs *FileState) validate() error {
	// Make sure the height is valid.
	if fs.Height < 0 {
		return errInvalidSignStateHeight
	}

	// Make sure the round is valid.
	if fs.Round < 0 {
		return errInvalidSignStateRound
	}

	// Make sure the sign step is valid.
	if fs.Step > StepPrecommit {
		return errInvalidSignStateStep
	}

	// Make sure the sign bytes are valid if set.
	if fs.SignBytes != nil {
		checkSignBytesHRS := func(height int64, round int, step Step) error {
			if height != fs.Height {
				return fmt.Errorf("%w: height mismatch", errInvalidSignStateSignBytes)
			}
			if round != fs.Round {
				return fmt.Errorf("%w: round mismatch", errInvalidSignStateSignBytes)
			}
			if step != fs.Step {
				return fmt.Errorf("%w: step mismatch", errInvalidSignStateSignBytes)
			}
			return nil
		}

		switch fs.Step {
		case StepPrevote, StepPrecommit:
			// Try to unmarshal as a canonical vote.
			var vote types.CanonicalVote
			if err := amino.UnmarshalSized(fs.SignBytes, &vote); err != nil {
				return errInvalidSignStateSignBytes
				// If SignBytes unmarshalled to vote, check if it contains the same HRS values.
			} else if err := checkSignBytesHRS(vote.Height, int(vote.Round), VoteTypeToStep(vote.Type)); err != nil {
				return err
			}

		case StepPropose:
			// Try to unmarshal as a canonical proposal.
			var proposal types.CanonicalProposal
			if err := amino.UnmarshalSized(fs.SignBytes, &proposal); err != nil {
				return errInvalidSignStateSignBytes
				// If SignBytes unmarshalled to proposal, check if it contains the same HRS values.
			} else if err := checkSignBytesHRS(proposal.Height, int(proposal.Round), StepPropose); err != nil {
				return err
			}

		default:
			// Invalid Step.
			return errInvalidSignStateSignBytes
		}
	}

	// Make sure the signature is not set if the sign bytes are not set.
	if fs.SignBytes == nil && fs.Signature != nil {
		return errSignatureShouldNotBeSet
	}

	// Check if the file path is set.
	if fs.filePath == "" {
		return errFilePathNotSet
	}

	return nil
}

// save persists the FileState to its file path.
func (fs *FileState) save() error {
	// Check if the FileState is valid.
	if err := fs.validate(); err != nil {
		return err
	}

	// Marshal the FileState to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(fs, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal FileState to JSON: %w", err)
	}

	// Write the JSON bytes to the file.
	if err := osm.WriteFileAtomic(fs.filePath, jsonBytes, 0o600); err != nil {
		return err
	}

	return nil
}

// LoadFileState reads a FileState from the given file path.
func LoadFileState(filePath string) (*FileState, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a FileState using amino.
	fs := &FileState{}
	err = amino.UnmarshalJSON(rawJSONBytes, fs)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal FileState from %s: %w", filePath, err)
	}

	// Manually set the private file path.
	fs.filePath = filePath

	// Validate the FileState.
	if err := fs.validate(); err != nil {
		return nil, err
	}

	return fs, nil
}

// GeneratePersistedFileState generates a new FileState persisted to disk.
func GeneratePersistedFileState(filePath string) (*FileState, error) {
	// Create a new FileState instance.
	fs := &FileState{
		filePath: filePath,
	}

	// Persist the FileState to disk.
	if err := fs.save(); err != nil {
		return nil, err
	}

	return fs, nil
}

// LoadOrMakeFileState returns a new FileState instance from the given file path.
// If the file does not exist, a new FileState is generated and persisted to disk.
func LoadOrMakeFileState(filePath string) (*FileState, error) {
	// If the file exists, load the FileState from the file.
	if osm.FileExists(filePath) {
		return LoadFileState(filePath)
	}

	// If the file does not exist, generate a new FileState and persist it to disk.
	return GeneratePersistedFileState(filePath)
}
