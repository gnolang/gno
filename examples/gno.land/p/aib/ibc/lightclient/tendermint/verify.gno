package tendermint

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
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
	if err := checkTrustedHeader(msg, consState); err != nil {
		return err
	}

	var (
		now             = time.Now()
		untrustedHeader = msg.Header
		untrustedValset = msg.ValidatorSet
		trustedHeader   = &Header{
			ChainID:            tm.ClientState.ChainID,
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
	// Check the validator hashes are the same
	if !bytes.Equal(untrustedHeader.ValidatorsHash, trustedHeader.NextValidatorsHash) {
		h1, h2 := hex.EncodeToString(trustedHeader.NextValidatorsHash), hex.EncodeToString(untrustedHeader.ValidatorsHash)
		return ufmt.Errorf(
			"expected old header next validators (%s) to match those from new header (%s)",
			h1, h2,
		)
	}

	if err := trustedHeader.Expired(tm.ClientState.TrustingPeriod); err != nil {
		return err
	}

	if untrustedHeader.Height != trustedHeader.Height+1 {
		// Heights aren't adjacent, this requires to check if at least trustlevel
		// ([1/3, 1]) of trusted validators have signed correctly.
		var (
			tl = tm.ClientState.TrustLevel
			// TODO recompute total voting power ? how to ensure it is correct ?
			votingPowerNeeded = trustedValset.TotalVotingPower *
				int64(tl.Numerator) / int64(tl.Denominator)
		)
		if err := tm.verifyCommit(trustedValset, msg.Commit, votingPowerNeeded, true, false); err != nil {
			return err
		}
	}
	// Ensure that +2/3 of untrustedValset signed correctly.
	//
	// NOTE: this should always be the last check because untrustedValset can be
	// intentionally made very large to DOS the light client.
	votingPowerNeeded := untrustedValset.TotalVotingPower * 2 / 3
	return tm.verifyCommit(untrustedValset, msg.Commit, votingPowerNeeded, false, true)
}

func (tm *TMLightClient) verifyMisbehavior(msg *Misbehaviour) error {
	panic("TODO verifyMisbehavior unimplemented")
}

// checkTrustedHeader checks that consensus state matches trusted fields of Header
func checkTrustedHeader(msg *MsgHeader, consensusState *ConsensusState) error {
	tvalHash := msg.TrustedValidators.Hash()
	if !bytes.Equal(consensusState.NextValidatorsHash, tvalHash) {
		h1, h2 := hex.EncodeToString(consensusState.NextValidatorsHash), hex.EncodeToString(tvalHash)
		return ufmt.Errorf(
			"trusted validators does not hash to latest trusted validators. Expected: %q, got: %q",
			h1, h2,
		)
	}
	return nil
}

// verifyCommit verifies the signatures included in a commit. Returns an error
// if votingPowerNeeded is not reached with the verified signatures.
// CONTRACT: both commit and validator set should have passed validate basic.
func (tm TMLightClient) verifyCommit(
	vals *ValidatorSet,
	commit *Commit,
	votingPowerNeeded int64,
	countAllSignatures bool,
	lookUpByIndex bool,
) error {
	var (
		val                *Validator
		valIdx             int32
		seenVals           = make(map[int32]int, len(commit.Signatures))
		talliedVotingPower int64
	)
	for idx, commitSig := range commit.Signatures {
		if commitSig.BlockIDFlag != BlockIDFlagCommit {
			continue
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
				addr := hex.EncodeToString(val.Address)
				return ufmt.Errorf("double vote from %s (%d and %d)", addr, firstIndex, secondIndex)
			}
			seenVals[valIdx] = idx
		}

		if len(val.PubKey) == 0 {
			addr := hex.EncodeToString(val.Address)
			return ufmt.Errorf("validator %s has a nil PubKey at index %d", addr, idx)
		}

		voteSignBytes := commit.BytesToSign(tm.ClientState.ChainID, idx)

		// CONTRACT: all validators use ed25519 keys
		if !ed25519.Verify(val.PubKey, voteSignBytes, commitSig.Signature) {
			return ufmt.Errorf("verify signature fail (#%d)", idx)
		}

		talliedVotingPower += val.VotingPower

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
