package status

import (
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

type SyncInfo struct {
	LatestBlockHash   []byte    `json:"latest_block_hash"`
	LatestAppHash     []byte    `json:"latest_app_hash"`
	LatestBlockHeight int64     `json:"latest_block_height"`
	LatestBlockTime   time.Time `json:"latest_block_time"`
	CatchingUp        bool      `json:"catching_up"`
}

type ValidatorInfo struct {
	Address     crypto.Address `json:"address"`
	PubKey      crypto.PubKey  `json:"pub_key"`
	VotingPower int64          `json:"voting_power"`
}

type ResultStatus struct {
	NodeInfo      p2pTypes.NodeInfo `json:"node_info"`
	SyncInfo      SyncInfo          `json:"sync_info"`
	ValidatorInfo ValidatorInfo     `json:"validator_info"`
}

func (s *ResultStatus) TxIndexEnabled() bool {
	if s == nil {
		return false
	}

	return s.NodeInfo.Other.TxIndex == "on"
}
