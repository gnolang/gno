package tendermint

import (
	"bytes"
	"encoding/hex"

	"gno.land/p/nt/ufmt"
)

func (tm *TMLightClient) verifyHeader(header *MsgHeader) error {
	if err := header.ValidateBasic(); err != nil {
		return ufmt.Errorf("header validate: %v", err)
	}
	// Retrieve trusted consensus states for trusted height
	consState, found := tm.GetConsensusState(header.TrustedHeight)
	if !found {
		return ufmt.Errorf("could not get trusted consensus state for Header at TrustedHeight: %s", header.TrustedHeight)
	}

	if err := checkTrustedHeader(header, consState); err != nil {
		return err
	}

	// UpdateClient only accepts updates with a header at the same revision
	// as the trusted consensus state
	if header.GetHeight().RevisionNumber != header.TrustedHeight.RevisionNumber {
		return ufmt.Errorf(
			"header height revision %d does not match trusted header revision %d",
			header.GetHeight().RevisionNumber, header.TrustedHeight.RevisionNumber,
		)
	}

	// assert header height is newer than consensus state
	if header.GetHeight().LTE(header.TrustedHeight) {
		return ufmt.Errorf(
			"header height ≤ consensus state height (%s ≤ %s)",
			header.GetHeight(), header.TrustedHeight,
		)
	}

	// TODO remaining of verifyHeader

	return nil
}

// checkTrustedHeader checks that consensus state matches trusted fields of
// Header
func checkTrustedHeader(header *MsgHeader, consState *ConsensusState) error {
	// assert that trustedVals is NextValidators of last trusted header
	// to do this, we check that
	// trustedVals.Hash() == consState.NextValidatorsHash
	tvalHash := header.TrustedValidators.Hash()
	if !bytes.Equal(consState.NextValidatorsHash, tvalHash) {
		h1, h2 := hex.EncodeToString(consState.NextValidatorsHash), hex.EncodeToString(tvalHash)
		return ufmt.Errorf(
			"trusted validators does not hash to latest trusted validators. Expected: %q, got: %q",
			h1, h2,
		)
	}
	return nil
}

func (tm *TMLightClient) verifyMisbehavior(msg *Misbehaviour) error {
	panic("TODO verifyMisbehavior unimplemented")
}
