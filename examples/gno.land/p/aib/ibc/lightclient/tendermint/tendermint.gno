package tendermint

import (
	"errors"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/avl"
)

type TMLightClient struct {
	ClientState            *ClientState
	ConsensusStateByHeight avl.Tree // key: height, value: *ConsensusState
}

var _ lightclient.LightClientModule = (*TMLightClient)(nil)

// Implements lightclient.LightClientModule
func (tm *TMLightClient) Initialize(clientState, consensusState any) error {
	cs := clientState.(ClientState)
	if err := cs.Validate(); err != nil {
		return err
	}
	tm.ClientState = &cs
	consState := consensusState.(ConsensusState)
	if err := consState.Validate(); err != nil {
		return err
	}
	tm.ConsensusStateByHeight.Set(cs.LatestHeight.String(), &consState)
	return nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) VerifyClientMessage(clientMsg any) error {
	switch msg := clientMsg.(type) {
	case *MsgHeader:
		return tm.verifyHeader(msg)
	case *Misbehaviour:
		return tm.verifyMisbehavior(msg)
	default:
		return errors.New("unknown client message type")
	}
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) CheckForMisbehaviour(clientID string, clientMsg any) bool {
	// TODO
	return false
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) UpdateStateOnMisbehaviour(clientID string, clientMsg any) {
	// TODO
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) UpdateState(clientMsg any) []types.Height {
	msg := clientMsg.(*MsgHeader)
	// check for duplicate update
	msgHeight := msg.GetHeight()
	if tm.ConsensusStateByHeight.Has(msgHeight.String()) {
		// perform no-op
		return []types.Height{msgHeight}
	}
	// Update latestHeight if required
	if msgHeight.GT(tm.ClientState.LatestHeight) {
		tm.ClientState.LatestHeight = msgHeight
	}
	// Build and store new consensus state from clientMsg
	consState := &ConsensusState{
		Timestamp:          msg.Header.Time,
		Root:               MerkleRoot{Hash: msg.Header.AppHash},
		NextValidatorsHash: msg.Header.NextValidatorsHash,
	}
	tm.ConsensusStateByHeight.Set(msgHeight.String(), consState)
	// TODO what about consensus metadata ?

	return []types.Height{msgHeight}
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) VerifyMembership(clientID string, height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte, path string,
	value []byte) error {
	// TODO
	return nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) VerifyNonMembership(clientID string, height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte, path string) error {
	// TODO
	return nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) Status(clientID string) lightclient.Status {
	// TODO
	return lightclient.Unknown
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) LatestHeight() types.Height {
	return tm.ClientState.LatestHeight
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) TimestampAtHeight(clientID string, height types.Height) (uint64, error) {
	// TODO
	return 0, nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) RecoverClient(clientID, substituteClientID string) error {
	// TODO
	return nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) VerifyUpgradeAndUpdateState(clientID string,
	newClient []byte, newConsState []byte, upgradeClientProof,
	upgradeConsensusStateProof []byte) error {
	// TODO
	return nil
}
