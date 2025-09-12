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

// TODO remove clientID ?
func (tm *TMLightClient) Initialize(clientID string, clientState, consensusState any) error {
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

func (tm *TMLightClient) VerifyClientMessage(clientID string, clientMsg any) error {
	switch msg := clientMsg.(type) {
	case *MsgHeader:
		return tm.verifyHeader(msg)
	case *Misbehaviour:
		return tm.verifyMisbehavior(msg)
	default:
		return errors.New("unknown client message type")
	}
}

func (tm *TMLightClient) CheckForMisbehaviour(clientID string, clientMsg any) bool {
	// TODO
	return false
}

func (tm *TMLightClient) UpdateStateOnMisbehaviour(clientID string, clientMsg any) {
	// TODO
}

func (tm *TMLightClient) UpdateState(clientID string, clientMsg any) []int64 {
	// TODO
	return nil
}

func (tm *TMLightClient) VerifyMembership(clientID string, height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte, path string,
	value []byte) error {
	// TODO
	return nil
}

func (tm *TMLightClient) VerifyNonMembership(clientID string, height types.Height,
	delayTimePeriod uint64, delayBlockPeriod uint64, proof []byte, path string) error {
	// TODO
	return nil
}

func (tm *TMLightClient) Status(clientID string) lightclient.Status {
	// TODO
	return lightclient.Unknown
}

func (tm *TMLightClient) LatestHeight(clientID string) int64 {
	// TODO
	return 0
}

func (tm *TMLightClient) TimestampAtHeight(clientID string, height types.Height) (uint64, error) {
	// TODO
	return 0, nil
}

func (tm *TMLightClient) RecoverClient(clientID, substituteClientID string) error {
	// TODO
	return nil
}

func (tm *TMLightClient) VerifyUpgradeAndUpdateState(clientID string,
	newClient []byte, newConsState []byte, upgradeClientProof,
	upgradeConsensusStateProof []byte) error {
	// TODO
	return nil
}
