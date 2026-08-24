package core

import (
	"fmt"
	"time"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
	"github.com/gnolang/gno/tm2/pkg/version"
)

// Status returns Tendermint status including node info, pubkey, latest block
// hash, app hash, block height and time.
//
// `heightGte` optionally returns 409 if the latest chain height is less than
// it, which is useful for readyness probes.
func (env *Environment) Status(ctx *rpctypes.Context, heightGtePtr *int64) (*ctypes.ResultStatus, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Status")
	defer span.End()
	var latestHeight int64
	if env.GetFastSync() {
		latestHeight = env.BlockStore.Height()
	} else {
		latestHeight = env.Consensus.GetLastHeight()
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
		latestBlockMeta = env.BlockStore.LoadBlockMeta(latestHeight)
		if latestBlockMeta == nil {
			return nil, fmt.Errorf("block meta not found for height %d", latestHeight)
		}
		latestBlockHash = latestBlockMeta.BlockID.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTimeNano = latestBlockMeta.Header.Time.UnixNano()
	}

	latestBlockTime := time.Unix(0, latestBlockTimeNano)

	var votingPower int64
	if val := env.validatorAtHeight(latestHeight); val != nil {
		votingPower = val.VotingPower
	}

	result := &ctypes.ResultStatus{
		NodeInfo: env.P2PTransport.NodeInfo(),
		SyncInfo: ctypes.SyncInfo{
			LatestBlockHash:   latestBlockHash,
			LatestAppHash:     latestAppHash,
			LatestBlockHeight: latestHeight,
			LatestBlockTime:   latestBlockTime,
			CatchingUp:        env.GetFastSync(),
		},
		ValidatorInfo: ctypes.ValidatorInfo{
			Address:     env.PubKey.Address(),
			PubKey:      env.PubKey,
			VotingPower: votingPower,
		},
		BuildVersion: version.Version,
	}

	return result, nil
}

func (env *Environment) validatorAtHeight(h int64) *types.Validator {
	privValAddress := env.PubKey.Address()

	// If we're still at height h, search in the current validator set.
	lastBlockHeight, vals := env.Consensus.GetValidators()
	if lastBlockHeight == h {
		for _, val := range vals {
			if val.Address == privValAddress {
				return val
			}
		}
	}

	// If we've moved to the next height, retrieve the validator set from DB.
	if lastBlockHeight > h {
		vals, err := sm.LoadValidators(env.StateDB, h)
		if err != nil {
			return nil // should not happen
		}
		_, val := vals.GetByAddress(privValAddress)
		return val
	}

	return nil
}
