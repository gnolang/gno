package tendermint

import (
	"errors"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

type TMLightClient struct {
	ClientState            *ClientState
	ConsensusStateByHeight avl.Tree // key: height, value: *ConsensusState
}

var _ lightclient.Interface = (*TMLightClient)(nil)

// Implements lightclient.Interface
func (tm *TMLightClient) Initialize(clientState lightclient.ClientState, consensusState lightclient.ConsensusState) error {
	cs := clientState.(ClientState)
	tm.ClientState = &cs
	consState := consensusState.(ConsensusState)
	tm.ConsensusStateByHeight.Set(cs.LatestHeight.String(), &consState)
	return nil
}

// Implements lightclient.Interface
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

// Implements lightclient.Interface
func (tm *TMLightClient) CheckForMisbehaviour(clientMsg lightclient.ClientMessage) bool {
	// TODO
	panic("not implemented")
}

// Implements lightclient.Interface
func (tm *TMLightClient) UpdateStateOnMisbehaviour(clientMsg lightclient.ClientMessage) {
	// TODO
	panic("not implemented")
}

// Implements lightclient.Interface
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
	return []types.Height{msgHeight}
}

// Implements lightclient.Interface
func (tm *TMLightClient) VerifyMembership(height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte, path string,
	value []byte) error {
	// TODO
	panic("not implemented")
}

// Implements lightclient.Interface
func (tm *TMLightClient) VerifyNonMembership(height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte,
	path string) error {
	// TODO
	panic("not implemented")
}

// Implements lightclient.Interface
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

// Implements lightclient.Interface
func (tm *TMLightClient) LatestHeight() types.Height {
	return tm.ClientState.LatestHeight
}

// Implements lightclient.Interface
func (tm *TMLightClient) TimestampAtHeight(height types.Height) (uint64, error) {
	cs, found := tm.GetConsensusState(height)
	if !found {
		return 0, ufmt.Errorf("no consensus state found for height %s", height.String())
	}
	return uint64(cs.Timestamp.UnixNano()), nil
}

// Implements lightclient.Interface
func (tm *TMLightClient) RecoverClient(substituteClientID string) error {
	// TODO
	panic("not implemented")
}

// Implements lightclient.Interface
func (tm *TMLightClient) VerifyUpgradeAndUpdateState(newClient, newConsState,
	upgradeClientProof, upgradeConsensusStateProof []byte) error {
	// TODO
	panic("not implemented")
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
