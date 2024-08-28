package agent

import (
	"context"
	"fmt"
	"runtime"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"

	"github.com/gnolang/gnostats/proto"
)

type collector struct {
	// rpcClient is used by the collector to send RPC request to the Gno node
	caller rpcClient
}

// RPCClient and RPCBatch are local interfaces that include only the
// required methods to make them easily mockable ands allow testing
// of the collector
type rpcClient interface {
	NewBatch() rpcBatch
}
type rpcBatch interface {
	Status() error
	Validators() error
	NetInfo() error
	NumUnconfirmedTxs() error
	Block(*uint64) error
	BlockResults(*uint64) error
	Send(context.Context) ([]any, error)
}

// CollectDynamic collects dynamic info from the Gno node using RPC
func (c *collector) CollectDynamic(ctx context.Context) (*proto.DynamicInfo, error) {
	// Create a new batch of RPC requests
	batch := c.caller.NewBatch()

	for _, request := range [](func() error){
		// Request Status to get address, moniker and validator info
		batch.Status,
		// Request Validators to get the list of validators
		batch.Validators,
		// Request NetInfo to get peers info
		batch.NetInfo,
		// Request NumUnconfirmedTxs to get pending txs
		batch.NumUnconfirmedTxs,
		// Request Block to get the last block timestamp and proposer
		func() error { return batch.Block(nil) },
		// Request BlockResults to get the last block number, gas used and gas wanted
		func() error { return batch.BlockResults(nil) },
	} {
		if err := request(); err != nil {
			return nil, fmt.Errorf("unable to batch request: %w", err)
		}
	}

	// Send the batch of requests
	results, err := batch.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to send batch: %w", err)
	}

	// Cast responses to the appropriate types
	var (
		status     = results[0].(*ctypes.ResultStatus)
		validators = results[1].(*ctypes.ResultValidators)
		netInfo    = results[2].(*ctypes.ResultNetInfo)
		uncTxs     = results[3].(*ctypes.ResultUnconfirmedTxs)
		blk        = results[4].(*ctypes.ResultBlock)
		blkRes     = results[5].(*ctypes.ResultBlockResults)
	)

	// Convert the list of peers from NetInfo to proto type
	peers := make([]*proto.PeerInfo, len(netInfo.Peers))
	for i, peer := range netInfo.Peers {
		peers[i] = &proto.PeerInfo{
			Moniker: peer.NodeInfo.Moniker,
		}
		if peer.NodeInfo.NetAddress != nil {
			peers[i].P2PAddress = peer.NodeInfo.NetAddress.String()
		}
	}

	// Determine if the node is a validator for the last block by searching for
	// own validatorInfo address in validators list
	isValidator := false
	for _, validator := range validators.Validators {
		if validator.Address.Compare(status.ValidatorInfo.Address) == 0 {
			isValidator = true
		}
	}

	// Get gas used / wanted in DeliverTxs (if any)
	var gasUsed, gasWanted uint64
	if blkRes.Results != nil && len(blkRes.Results.DeliverTxs) > 0 {
		gasUsed = uint64(blkRes.Results.DeliverTxs[0].GasUsed)
		gasWanted = uint64(blkRes.Results.DeliverTxs[0].GasWanted)
	}

	// Fill the DynamicInfo fields with the corresponding values
	return &proto.DynamicInfo{
		Address:     status.NodeInfo.ID().String(),
		Moniker:     status.NodeInfo.Moniker,
		IsValidator: isValidator,
		NetInfo: &proto.NetInfo{
			P2PAddress: status.NodeInfo.NetAddress.String(),
			Peers:      peers,
		},
		PendingTxs: uint64(uncTxs.Total),
		BlockInfo: &proto.BlockInfo{
			Number:    uint64(blk.Block.Height),
			Timestamp: uint64(blk.Block.Time.Unix()),
			GasUsed:   gasUsed,
			GasWanted: gasWanted,
			Proposer:  blk.Block.ProposerAddress.ID().String(),
		},
	}, nil
}

// CollectStatic collects static info on the Gno node using RPC
func (c *collector) CollectStatic(ctx context.Context) (*proto.StaticInfo, error) {
	// Use a batch instead of a single request to allow passing a context
	batch := c.caller.NewBatch()

	// Request Status to get all node info
	if err := batch.Status(); err != nil {
		return nil, fmt.Errorf("unable to batch request: %w", err)
	}

	// Send the batch of requests
	results, err := batch.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to send batch: %w", err)
	}

	// Cast response to the appropriate type
	status := results[0].(*ctypes.ResultStatus)

	// Fill the StaticInfo fields with the corresponding values
	return &proto.StaticInfo{
		Address:    status.NodeInfo.ID().String(),
		GnoVersion: status.NodeInfo.Version,
		OsVersion:  fmt.Sprintf("%s - %s", runtime.GOOS, runtime.GOARCH),
	}, nil
}

// NewCollector creates a new collector using the provided RPC client
func NewCollector(caller rpcClient) *collector {
	return &collector{
		caller: caller,
	}
}
