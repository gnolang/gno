package core

import (
	"fmt"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

func BackupBlocks(ctx *rpctypes.Context, startHeight int64, endHeight int64) (*ctypes.ResultBackupBlock, error) {
	logger.Info("On BackupBlocks", "start", startHeight, "end", endHeight)
	_, span := traces.Tracer().Start(ctx.Context(), "BackupBlocks")
	defer span.End()

	if ctx == nil || ctx.WSConn == nil || ctx.JSONReq == nil {
		return nil, fmt.Errorf("backup method requires websocket context")
	}

	if startHeight == 0 {
		startHeight = 1
	}
	if startHeight < 1 {
		return nil, fmt.Errorf("start height must be >= 1, got %d", startHeight)
	}

	blockStoreHeight := blockStore.Height()
	if blockStoreHeight < 1 {
		return nil, fmt.Errorf("block store returned invalid max height (%d)", blockStoreHeight)
	}

	if endHeight == 0 {
		endHeight = blockStoreHeight
	} else if endHeight > blockStoreHeight {
		return nil, fmt.Errorf("end height must be <= %d", blockStoreHeight)
	}
	if startHeight > endHeight {
		return nil, fmt.Errorf("end height must be >= than start height")
	}

	for height := startHeight; height <= endHeight; height++ {
		block := blockStore.LoadBlock(height)
		if block == nil {
			return nil, fmt.Errorf("block store returned nil block for height %d", height)
		}

		// Emit one response per block
		ctx.WSConn.WriteRPCResponses(rpctypes.RPCResponses{
			rpctypes.NewRPCSuccessResponse(ctx.JSONReq.ID, &ctypes.ResultBackupBlock{
				Height: height,
				Block:  block,
			}),
		})
	}

	return &ctypes.ResultBackupBlock{
		Done: true,
	}, nil
}
