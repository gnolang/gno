package tendermint

import (
	"errors"
	"time"

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
func (tm *TMLightClient) Initialize(clientState lightclient.ClientState, consensusState lightclient.ConsensusState) error {
	cs := clientState.(ClientState)
	tm.ClientState = &cs
	consState := consensusState.(ConsensusState)
	tm.ConsensusStateByHeight.Set(cs.LatestHeight.String(), &consState)
	return nil
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) VerifyClientMessage(clientMsg lightclient.ClientMessage) error {
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
func (tm *TMLightClient) CheckForMisbehaviour(clientID string, clientMsg lightclient.ClientMessage) bool {
	// TODO
	return false
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) UpdateStateOnMisbehaviour(clientID string, clientMsg lightclient.ClientMessage) {
	// TODO
}

// Implements lightclient.LightClientModule
func (tm *TMLightClient) UpdateState(clientMsg lightclient.ClientMessage) []types.Height {
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
		Root:               NewMerkleRoot(msg.Header.AppHash),
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
func (tm *TMLightClient) Status() lightclient.Status {
	if !tm.ClientState.FrozenHeight.IsZero() {
		return lightclient.Frozen
	}
	// get latest consensus state to check for expiry
	lastConsState, found := tm.GetConsensusState(tm.LatestHeight())
	if !found {
		// if the client state does not have an associated consensus state for its
		// latest height then it must be expired
		return lightclient.Expired
	}
	if tm.IsExpired(lastConsState.Timestamp, time.Now()) {
		return lightclient.Expired
	}
	return lightclient.Active
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

// GetConsensusState returns the consensus state mapped at height, if any.
func (tm *TMLightClient) GetConsensusState(height types.Height) (*ConsensusState, bool) {
	x, found := tm.ConsensusStateByHeight.Get(height.String())
	if !found {
		return nil, false
	}
	return x.(*ConsensusState), true
}

// IsExpired returns whether or not the client has passed the trusting period
// since the last update (in which case no headers are considered valid).
func (tm *TMLightClient) IsExpired(latestTimestamp, now time.Time) bool {
	expirationTime := latestTimestamp.Add(tm.ClientState.TrustingPeriod)
	return !expirationTime.After(now)
}
