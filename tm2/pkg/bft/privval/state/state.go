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
	StepNone      Step = 0 // Undefined step
	StepPropose   Step = 1 // Only used in proposal signing
	StepPrevote   Step = 2 // Only used in vote signing
	StepPrecommit Step = 3 // Only used in vote signing
)

// A vote is either stepPrevote or stepPrecommit.
func VoteToStep(vote *types.Vote) Step {
	switch vote.Type {
	case types.PrevoteType:
		return StepPrevote
	case types.PrecommitType:
		return StepPrecommit
	default:
		panic("Unknown vote type")
	}
}

// FileState stores the state of the last signing operation in a file.
type FileState struct {
	Height    int64  `json:"height" comment:"the height of the last sign"`
	Round     int    `json:"round" comment:"the round of the last sign"`
	Step      Step   `json:"step" comment:"the step of the last sign"`
	SignBytes []byte `json:"signbytes,omitempty" comment:"the raw signature bytes of the last sign"`
	Signature []byte `json:"signature,omitempty" comment:"the signature of the last sign"`

	filePath string
}

// checkHRS checks the given height, round, step (HRS) against the last state.
// It returns an error if the arguments constitute a regression, or if HRS match but
// the SignBytes are not set.
// The returned boolean indicates whether the last Signature should be reused or not.
// It will be true if the HRS match and the SignBytes and Signature are already set
// in the last state (indicating we have already signed for this HRS).
func (fs *FileState) CheckHRS(height int64, round int, step Step) (bool, error) {
	// Check if the height differs.
	if height < fs.Height { // Height regression
		return false, fmt.Errorf("height regression. Got %v, last height %v", height, fs.Height)
	}
	if height > fs.Height { // New height, we can't reuse the signature.
		return false, nil
	}

	// Height is the same, now check if the round differs.
	if round < fs.Round { // Round regression
		return false, fmt.Errorf("round regression at height %v. Got %v, last round %v", height, round, fs.Round)
	}
	if round > fs.Round { // New round, we can't reuse the signature.
		return false, nil
	}

	// Round is the same, now check if the step differs.
	if step < fs.Step { // Step regression
		return false, fmt.Errorf("step regression at height %v round %v. Got %v, last step %v", height, round, step, fs.Step)
	}
	if step > fs.Step { // New step, we can't reuse the signature.
		return false, nil
	}

	// If the HRS are the same, the SignBytes should be already set.
	if fs.SignBytes == nil {
		return false, errors.New("no SignBytes found")
	}

	// If the SignBytes are set, the Signature should be set as well.
	if fs.Signature == nil {
		panic("FileeState: Signature is nil but SignBytes is not!")
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
func (fs *FileState) Update(
	height int64,
	round int,
	step Step,
	signBytes []byte,
	signature []byte,
) error {
	fs.Height = height
	fs.Round = round
	fs.Step = step
	fs.SignBytes = signBytes
	fs.Signature = signature

	return fs.save()
}

// save persists the FileState to its file path.
func (fs *FileState) save() error {
	// Check if the file path is set.
	if fs.filePath == "" {
		return errors.New("cannot save FileState: filePath not set")
	}

	// Marshal the FileState to JSON bytes using amino.
	jsonBytes, err := amino.MarshalJSONIndent(fs, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON bytes to the file.
	err = osm.WriteFileAtomic(fs.filePath, jsonBytes, 0o600)
	if err != nil {
		return err
	}

	return nil
}

// loadFileState reads a FileState from the given file path.
func loadFileState(filePath string) (*FileState, error) {
	// Read the JSON bytes from the file.
	rawJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON bytes into a FileState using amino.
	fs := &FileState{}
	err = amino.UnmarshalJSON(rawJSONBytes, &fs)
	if err != nil {
		return nil, fmt.Errorf("Error reading FileState from %v: %v\n", filePath, err)
	}

	// Set the file path for the FileState.
	fs.filePath = filePath

	return fs, nil
}

// generateFileState generate a new FileState and persists it to disk.
func generateFileState(filePath string) (*FileState, error) {
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

// NewFileState returns a new FileState instance from the given file path.
// If the file does not exist, a new FileState is generated and persisted to disk.
func NewFileState(filePath string) (*FileState, error) {
	// If the file exists, load the FileState from the file.
	if osm.FileExists(filePath) {
		return loadFileState(filePath)
	}

	// If the file does not exist, generate a new FileState and persist it to disk.
	return generateFileState(filePath)
}
