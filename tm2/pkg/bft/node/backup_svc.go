package node

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/gnolang/gno/tm2/pkg/amino"
	backup "github.com/gnolang/gno/tm2/pkg/bft/node/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/node/backuppb/backupconnect"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
)

type backupServer struct {
	store *store.BlockStore
}

// StreamBlocks implements backupconnect.BackupServiceHandler.
func (b *backupServer) StreamBlocks(_ context.Context, req *connect.Request[backup.StreamBlocksRequest], stream *connect.ServerStream[backup.StreamBlocksResponse]) error {
	startHeight := req.Msg.StartHeight
	if startHeight == 0 {
		startHeight = 1
	}
	if startHeight < 1 {
		return fmt.Errorf("start height must be >= 1, got %d", startHeight)
	}

	endHeight := req.Msg.EndHeight
	blockStoreHeight := b.store.Height()
	if endHeight == 0 {
		endHeight = blockStoreHeight
	} else if endHeight > blockStoreHeight {
		return fmt.Errorf("end height must be <= %d", blockStoreHeight)
	}

	if startHeight > endHeight {
		return fmt.Errorf("end height must be >= than start height")
	}

	for height := startHeight; height <= endHeight; height++ {
		block := b.store.LoadBlock(height)
		data, err := amino.Marshal(block)
		if err != nil {
			return err
		}

		if err := stream.Send(&backup.StreamBlocksResponse{
			Height: height,
			Data:   data,
		}); err != nil {
			return err
		}
	}

	return nil
}

var _ backupconnect.BackupServiceHandler = (*backupServer)(nil)
