package backup

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"google.golang.org/grpc"
)

type Config struct {
	// Address for the backup server to listen on. Empty means disabled.
	ListenAddress string `json:"laddr" toml:"laddr" comment:"Address for the backup server to listen on. Empty means disabled."`
}

func DefaultConfig() *Config {
	return &Config{}
}

type blockStore interface {
	Height() int64
	LoadBlock(height int64) *types.Block
}

func NewBackupServiceHandler(store blockStore) backuppb.BackupServiceServer {
	backupServ := &backupServer{store: store}
	return backupServ
}

type backupServer struct {
	backuppb.UnimplementedBackupServiceServer
	store blockStore
}

// StreamBlocks implements backuppbconnect.BackupServiceHandler.
func (b *backupServer) StreamBlocks(req *backuppb.StreamBlocksRequest, stream grpc.ServerStreamingServer[backuppb.StreamBlocksResponse]) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	startHeight := req.StartHeight
	if startHeight == 0 {
		startHeight = 1
	}
	if startHeight < 1 {
		return fmt.Errorf("start height must be >= 1, got %d", startHeight)
	}

	blockStoreHeight := b.store.Height()
	if blockStoreHeight < 1 {
		return fmt.Errorf("block store returned invalid max height (%d)", blockStoreHeight)
	}

	endHeight := req.EndHeight
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
		if block == nil {
			return fmt.Errorf("block store returned nil block for height %d", height)
		}

		data, err := amino.Marshal(block)
		if err != nil {
			return err
		}

		if err := stream.Send(&backuppb.StreamBlocksResponse{Data: data}); err != nil {
			return err
		}
	}

	return nil
}

var _ backuppb.BackupServiceServer = (*backupServer)(nil)
