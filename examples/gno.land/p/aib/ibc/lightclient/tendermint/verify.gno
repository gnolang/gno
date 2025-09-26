package tendermint

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"time"

	"gno.land/p/nt/ufmt"
)

func (tm *TMLightClient) verifyHeader(msg *MsgHeader) error {
	// Retrieve trusted consensus states for trusted height
	consState, found := tm.GetConsensusState(msg.TrustedHeight)
	if !found {
		return ufmt.Errorf("could not get trusted consensus state for Header at TrustedHeight: %s", msg.TrustedHeight)
	}

	// UpdateClient only accepts updates with a header at the same revision
	// as the trusted consensus state
	if msg.GetHeight().RevisionNumber != msg.TrustedHeight.RevisionNumber {
		return ufmt.Errorf(
			"header height revision %d does not match trusted header revision %d",
			msg.GetHeight().RevisionNumber, msg.TrustedHeight.RevisionNumber,
		)
	}

	// TODO remaining of verifyHeader
	var (
		now             = time.Now()
		untrustedHeader = msg.Header
		untrustedValset = msg.ValidatorSet
		trustedHeader   = &Header{
			ChainID:            tm.ClientState.ChainId,
			Height:             msg.TrustedHeight.RevisionHeight,
			Time:               consState.Timestamp,
			NextValidatorsHash: consState.NextValidatorsHash,
		}
		trustedValset = msg.TrustedValidators
	)

	// UpdateClient only accepts updates with a header at the same revision
	// as the trusted consensus state
	if untrustedHeader.ChainID != trustedHeader.ChainID {
		return ufmt.Errorf("header belongs to another chain %q, not %q",
			untrustedHeader.ChainID, trustedHeader.ChainID)
	}

	if !untrustedHeader.Time.After(trustedHeader.Time) {
		return ufmt.Errorf(
			"expected new header time %v to be after old header time %v",
			untrustedHeader.Time, trustedHeader.Time,
		)
	}

	if !untrustedHeader.Time.Before(now.Add(tm.ClientState.MaxClockDrift)) {
		return ufmt.Errorf(
			"new header has a time from the future %v (now: %v; max clock drift: %v)",
			untrustedHeader.Time, now, tm.ClientState.MaxClockDrift,
		)
	}

	if err := checkTrustedHeader(trustedHeader, trustedValset); err != nil {
		return err
	}

	if untrustedHeader.Height == trustedHeader.Height+1 {
		return tm.verifyAdjacent(
			trustedHeader, untrustedHeader, untrustedValset, msg.Commit,
		)
	}
	return tm.verifyNonAdjacent(
		trustedHeader, trustedValset,
		untrustedHeader, untrustedValset,
	)
}

func (tm *TMLightClient) verifyMisbehavior(msg *Misbehaviour) error {
	panic("TODO verifyMisbehavior unimplemented")
}

// checkTrustedHeader checks that trustedHeader.NextValidatorsHash matches
// trusted validator set hash.
func checkTrustedHeader(trustedHeader *Header, trustedValset *ValidatorSet) error {
	// assert that trustedVals is NextValidators of last trusted header
	// to do this, we check that
	// trustedVals.Hash() == consState.NextValidatorsHash
	tvalHash := trustedValset.Hash()
	if !bytes.Equal(trustedHeader.NextValidatorsHash, tvalHash) {
		h1, h2 := hex.EncodeToString(trustedHeader.NextValidatorsHash), hex.EncodeToString(tvalHash)
		return ufmt.Errorf(
			"trusted validators does not hash to latest trusted validators. Expected: %q, got: %q",
			h1, h2,
		)
	}
	return nil
}

// verifyAdjacent verifies directly adjacent untrustedHeader against
// trustedHeader. It ensures that:
//
// a) headers are adjacent.
// b) trustedHeader is not expired
// c) untrustedHeader.ValidatorsHash equals trustedHeader.NextValidatorsHash
// d) more than 2/3 of new validators (untrustedVals) have signed h2
func (tm TMLightClient) verifyAdjacent(
	trustedHeader *Header, //height=X
	untrustedHeader *Header, //height=X+1
	untrustedValset *ValidatorSet, //height=X+1
	commit *Commit, //height=X+1
) error {
	// TODO remove, already done earlier
	if untrustedHeader.Height != trustedHeader.Height+1 {
		return errors.New("headers must be adjacent in height")
	}
	// TODO could be done earlier too
	if err := headerExpired(trustedHeader, tm.ClientState.TrustingPeriod); err != nil {
		return err
	}
	// Check the validator hashes are the same
	// TODO isn't it already done somewhere else?
	if !bytes.Equal(untrustedHeader.ValidatorsHash, trustedHeader.NextValidatorsHash) {
		h1, h2 := hex.EncodeToString(trustedHeader.NextValidatorsHash), hex.EncodeToString(untrustedHeader.ValidatorsHash)
		return ufmt.Errorf(
			"expected old header next validators (%s) to match those from new header (%s)",
			h1, h2,
		)
	}

	// Ensure that +2/3 of new validators signed correctly.
	return tm.verifyCommit(untrustedValset, commit)
}

// verifyNonAdjacent verifies non-adjacent untrustedHeader against
// trustedHeader. It ensures that:
//
// a) trustedHeader can still be trusted
// b) untrustedHeader is valid
// c) trustLevel ([1/3, 1]) of trustedHeaderVals (or trustedHeaderNextVals)
//
//	signed correctly
//
// d) more than 2/3 of untrustedVals have signed h2
// e) headers are non-adjacent.
//
// maxClockDrift defines how much untrustedHeader.Time can drift into the
// future.
func (tm TMLightClient) verifyNonAdjacent(
	trustedHeader *Header, //height=X
	trustedValset *ValidatorSet, //height=X or height=X+1
	untrustedHeader *Header, //height=Y
	untrustedValset *ValidatorSet, //height=Y
) error {
	// TODO
	return errors.New("not implemented")
}

// headerExpired return true if the given header expired.
func headerExpired(h *Header, trustingPeriod time.Duration) error {
	var (
		expirationTime = h.Time.Add(trustingPeriod)
		now            = time.Now()
	)
	if !expirationTime.After(now) {
		return ufmt.Errorf("header %d has expired at %v (now: %v)", h.Height, expirationTime, now)
	}
	return nil
}

func (tm TMLightClient) verifyCommit(
	vals *ValidatorSet,
	commit *Commit,
) error {
	// run a basic validation of the arguments
	if err := verifyBasicValsAndCommit(vals, commit); err != nil {
		return err
	}

	var (
		// calculate voting power needed
		votingPowerNeeded = vals.TotalVotingPower * 2 / 3

		// ignore all commit signatures that are not for the block
		ignore = func(c CommitSig) bool { return c.BlockIDFlag != BlockIDFlagCommit }

		// count all the remaining signatures
		count = func(c CommitSig) bool { return true }

		chainID = tm.ClientState.ChainId
	)

	return verifyCommitSingle(chainID, vals, commit, votingPowerNeeded,
		ignore, count, false, true)
}

func verifyBasicValsAndCommit(vals *ValidatorSet, commit *Commit) error {
	// TODO move to MsgHeader.Validate
	if len(vals.Validators) != len(commit.Signatures) {
		return ufmt.Errorf(
			"Invalid commit -- wrong set size: %v vs %v",
			len(vals.Validators), len(commit.Signatures),
		)
	}
	return nil
}

// Single Verification

// verifyCommitSingle single verifies commits.
// If a key does not support batch verification, or batch verification fails this will be used
// This method is used to check all the signatures included in a commit.
// It is used in consensus for validating a block LastCommit.
// CONTRACT: both commit and validator set should have passed validate basic
func verifyCommitSingle(
	chainID string,
	vals *ValidatorSet,
	commit *Commit,
	// TODO check if we can remove the following params
	votingPowerNeeded int64,
	ignoreSig func(CommitSig) bool,
	countSig func(CommitSig) bool,
	countAllSignatures bool,
	lookUpByIndex bool,
) error {
	var (
		val                *Validator
		valIdx             int32
		seenVals           = make(map[int32]int, len(commit.Signatures))
		talliedVotingPower int64
		voteSignBytes      []byte
	)
	for idx, commitSig := range commit.Signatures {
		if ignoreSig(commitSig) {
			continue
		}

		// TODO check should be done before
		if commitSig.Validate() != nil {
			return ufmt.Errorf("invalid signatures from %v at index %d", val, idx)
		}

		// If the vals and commit have a 1-to-1 correspondence we can retrieve
		// them by index else we need to retrieve them by address
		if lookUpByIndex {
			val = vals.Validators[idx]
		} else {
			valIdx, val = vals.GetByAddress(commitSig.ValidatorAddress)

			// if the signature doesn't belong to anyone in the validator set
			// then we just skip over it
			if val == nil {
				continue
			}

			// because we are getting validators by address we need to make sure
			// that the same validator doesn't commit twice
			if firstIndex, ok := seenVals[valIdx]; ok {
				secondIndex := idx
				return ufmt.Errorf("double vote from %v (%d and %d)", val, firstIndex, secondIndex)
			}
			seenVals[valIdx] = idx
		}

		// TODO move to validate
		if val.PubKey == nil {
			return ufmt.Errorf("validator %v has a nil PubKey at index %d", val, idx)
		}

		voteSignBytes = VoteBytesToSign(
			chainID, int64(commit.Height), int64(commit.Round), commit.BlockID.Hash,
			commit.BlockID.PartSetHeader.Total, commit.BlockID.PartSetHeader.Hash,
			commit.Signatures[idx].BlockIDFlag, commit.Signatures[idx].Timestamp,
		)

		// CONTRACT: all validators have ed25519 keys
		if !ed25519.Verify(val.PubKey, voteSignBytes, commitSig.Signature) {
			return ufmt.Errorf("verify signature fail (#%d)", idx)
		}

		// If this signature counts then add the voting power of the validator
		// to the tally
		if countSig(commitSig) {
			talliedVotingPower += val.VotingPower
		}

		// check if we have enough signatures and can thus exit early
		if !countAllSignatures && talliedVotingPower > votingPowerNeeded {
			return nil
		}
	}

	if got, needed := talliedVotingPower, votingPowerNeeded; got <= needed {
		return ufmt.Errorf("not enough voting power, got %v, needed %v", got, needed)
	}

	return nil
}
