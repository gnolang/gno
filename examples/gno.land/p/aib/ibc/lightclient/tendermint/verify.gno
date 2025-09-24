package tendermint

import (
	"bytes"
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

	// TODO remaining of verifyHeader
	if untrustedHeader.Height == trustedHeader.Height+1 {
		return tm.verifyAdjacent(
			trustedHeader, untrustedHeader, untrustedValset,
		)
	}
	return tm.verifyNonAdjacent(
		trustedHeader, trustedValset,
		untrustedHeader, untrustedValset,
	)
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
// b) untrustedHeader is valid
// c) untrustedHeader.ValidatorsHash equals trustedHeader.NextValidatorsHash
// d) more than 2/3 of new validators (untrustedVals) have signed h2
//
// maxClockDrift defines how much untrustedHeader.Time can drift into the
// future.
func (tm TMLightClient) verifyAdjacent(
	trustedHeader *Header, //height=X
	untrustedHeader *Header, //height=X+1
	untrustedValset *ValidatorSet, //height=X+1
) error {
	// TODO consider skip some checks that have been already done before
	if untrustedHeader.Height != trustedHeader.Height+1 {
		return errors.New("headers must be adjacent in height")
	}
	if err := headerExpired(trustedHeader, tm.ClientState.TrustingPeriod); err != nil {
		return err
	}
	// TODO remaining
	return nil
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
	return nil
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

func (tm *TMLightClient) verifyMisbehavior(msg *Misbehaviour) error {
	panic("TODO verifyMisbehavior unimplemented")
}
