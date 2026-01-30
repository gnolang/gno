package core

import (
	"fmt"
	"time"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Get Tendermint status including node info, pubkey, latest block
// hash, app hash, block height and time.
//
// ```shell
// curl 'localhost:26657/status'
// ```
//
// Additionally, it has an optional `heightGte` parameter than will return a `409` if the latest chain height is less than it.
// This parameter is useful for readyness probes.
//
// ```shell
// curl 'localhost:26657/status?heightGte=1'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:26657", "/websocket")
// err := client.Start()
//
//	if err != nil {
//	  // handle error
//	}
//
// defer client.Stop()
// result, err := client.Status()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
// {
// "jsonrpc": "2.0",
// "id": "",
//
//	"result": {
//	  "node_info": {
//	  		"protocol_version": {
//	  			"p2p": "4",
//	  			"block": "7",
//	  			"app": "0"
//	  		},
//	  		"id": "53729852020041b956e86685e24394e0bee4373f",
//	  		"listen_addr": "10.0.2.15:26656",
//	  		"network": "test-chain-Y1OHx6",
//	  		"version": "0.24.0-2ce1abc2",
//	  		"channels": "4020212223303800",
//	  		"moniker": "ubuntu-xenial",
//	  		"other": {
//	  			"tx_index": "on",
//	  			"rpc_addr": "tcp://0.0.0.0:26657"
//	  		}
//	  	},
//	  	"sync_info": {
//	  		"latest_block_hash": "F51538DA498299F4C57AC8162AAFA0254CE08286",
//	  		"latest_app_hash": "0000000000000000",
//	  		"latest_block_height": "18",
//	  		"latest_block_time": "2018-09-17T11:42:19.149920551Z",
//	  		"catching_up": false
//	  	},
//	  	"validator_info": {
//	  		"address": "D9F56456D7C5793815D0E9AF07C3A355D0FC64FD",
//	  		"pub_key": {
//	  			"type": "tendermint/PubKeyEd25519",
//	  			"value": "wVxKNtEsJmR4vvh651LrVoRguPs+6yJJ9Bz174gw9DM="
//	  		},
//	  		"voting_power": "10"
//	  	}
//	  }
//	}
//
// ```
func Status(ctx *rpctypes.Context, heightGtePtr *int64) (*ctypes.ResultStatus, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Status")
	defer span.End()
	var latestHeight int64
	if getFastSync() {
		latestHeight = blockStore.Height()
	} else {
		latestHeight = consensusState.GetLastHeight()
	}

	if heightGtePtr != nil && latestHeight < *heightGtePtr {
		// Using `409 Conflict` since it's spec states:
		// > 409 responses may be used for implementation-specific purposes
		return nil, rpctypes.NewHTTPStatusError(409, fmt.Sprintf("latest height is %d, which is less than %d", latestHeight, *heightGtePtr))
	}

	var (
		latestBlockMeta     *types.BlockMeta
		latestBlockHash     []byte
		latestAppHash       []byte
		latestBlockTimeNano int64
	)
	if latestHeight != 0 {
		latestBlockMeta = blockStore.LoadBlockMeta(latestHeight)
		latestBlockHash = latestBlockMeta.BlockID.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTimeNano = latestBlockMeta.Header.Time.UnixNano()
	}

	latestBlockTime := time.Unix(0, latestBlockTimeNano)

	var votingPower int64
	if val := validatorAtHeight(latestHeight); val != nil {
		votingPower = val.VotingPower
	}

	result := &ctypes.ResultStatus{
		NodeInfo: p2pTransport.NodeInfo(),
		SyncInfo: ctypes.SyncInfo{
			LatestBlockHash:   latestBlockHash,
			LatestAppHash:     latestAppHash,
			LatestBlockHeight: latestHeight,
			LatestBlockTime:   latestBlockTime,
			CatchingUp:        getFastSync(),
		},
		ValidatorInfo: ctypes.ValidatorInfo{
			Address:     pubKey.Address(),
			PubKey:      pubKey,
			VotingPower: votingPower,
		},
	}

	return result, nil
}

func validatorAtHeight(h int64) *types.Validator {
	privValAddress := pubKey.Address()

	// If we're still at height h, search in the current validator set.
	lastBlockHeight, vals := consensusState.GetValidators()
	if lastBlockHeight == h {
		for _, val := range vals {
			if val.Address == privValAddress {
				return val
			}
		}
	}

	// If we've moved to the next height, retrieve the validator set from DB.
	if lastBlockHeight > h {
		vals, err := sm.LoadValidators(stateDB, h)
		if err != nil {
			return nil // should not happen
		}
		_, val := vals.GetByAddress(privValAddress)
		return val
	}

	return nil
}
