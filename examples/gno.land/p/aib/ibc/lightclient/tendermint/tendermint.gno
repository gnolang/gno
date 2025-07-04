package tendermint

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/demo/avl"
)

type TMLightClient struct {
	ClientState            *ClientState
	ConsensusStateByHeight avl.Tree // key: height, value: *ConsensusState
}

var _ lightclient.LightClientModule = (*TMLightClient)(nil)

// TODO remove clientID ?
func (tm *TMLightClient) Initialize(clientID string, clientStateBz, consensusStateBz []byte) error {
	/*
		// FIXME find encoding alternative
		if err := json.Unmarshal(clientStateBz, tm.ClientState); err != nil {
			// TODO embed err
			return err
		}
		// TODO validate clientState

		var consensusState ConsensusState
		if err := json.Unmarshal(consensusStateBz, &consensusState); err != nil {
			// TODO enmbed err
			return err
		}
		// TODO validate consensusState

		tm.ConsensusStateByHeight.Set(tm.ClientState.LatestHeight.String(), &consensusState)
	*/
	return nil
}

func (tm *TMLightClient) VerifyClientMessage(clientID string, clientMsg any) error {
	// TODO
	return nil
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
