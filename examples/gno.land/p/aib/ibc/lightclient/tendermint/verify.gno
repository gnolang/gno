package tendermint

import (
	"bytes"

	"gno.land/p/nt/ufmt"
)

func (tm *TMLightClient) verifyHeader(header *MsgHeader) error {
	if err := header.ValidateBasic(); err != nil {
		return ufmt.Errorf("header validate: %v", err)
	}
	// Retrieve trusted consensus states for trusted height
	x, found := tm.ConsensusStateByHeight.Get(header.TrustedHeight.String())
	if !found {
		return ufmt.Errorf("could not get trusted consensus state for Header at TrustedHeight: %s", header.TrustedHeight.String())
	}
	consState := x.(*ConsensusState)

	if err := checkTrustedHeader(header, consState); err != nil {
		return err
	}

	// UpdateClient only accepts updates with a header at the same revision
	// as the trusted consensus state
	// TODO deal with that
	// if header.GetHeight().GetRevisionNumber() != header.TrustedHeight.RevisionNumber {
	// if header.Header.Height != header.TrustedHeight.RevisionNumber {
	// return ufmt.Errorf(
	// "header height revision %d does not match trusted header revision %d",
	// header.GetHeight().GetRevisionNumber(), header.TrustedHeight.RevisionNumber,
	// )
	// }

	// TODO remaining of verifyHeader

	return nil
}

// checkTrustedHeader checks that consensus state matches trusted fields of Header
func checkTrustedHeader(header *MsgHeader, consState *ConsensusState) error {
	// TODO define a way to hash validator set that is affordable for gnovm
	// TODO check ibc-go checkTrustedHeader to implements all the checks

	// assert that trustedVals is NextValidators of last trusted header
	// to do this, we check that trustedVals.Hash() == consState.NextValidatorsHash
	tvalHash := []byte("XXX") // header.TrustedValidators.Hash()
	if !bytes.Equal(consState.NextValidatorsHash, tvalHash) {
		return ufmt.Errorf(
			"trusted validators %s, does not hash to latest trusted validators. Expected: %X, got: %X",
			header.TrustedValidators, consState.NextValidatorsHash, tvalHash,
		)
	}
	return nil
}

func (tm *TMLightClient) verifyMisbehavior(msg *Misbehaviour) error {
	panic("TODO verifyMisbehavior unimplemented")
}
