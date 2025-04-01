package backup

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb/backuppbconnect"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

func NewServer(conf *Config, store blockStore) *http.Server {
	backupServ := &backupServer{store: store}
	mux := http.NewServeMux()
	path, handler := backuppbconnect.NewBackupServiceHandler(backupServ)
	mux.Handle(path, handler)
	return &http.Server{Addr: conf.ListenAddress, Handler: h2c.NewHandler(mux, &http2.Server{}), ReadHeaderTimeout: time.Second * 5}
}

type backupServer struct {
	store blockStore
}

// StreamBlocks implements backuppbconnect.BackupServiceHandler.
func (b *backupServer) StreamBlocks(_ context.Context, req *connect.Request[backuppb.StreamBlocksRequest], stream *connect.ServerStream[backuppb.StreamBlocksResponse]) error {
	startHeight := req.Msg.StartHeight
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

	endHeight := req.Msg.EndHeight
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

		if err := stream.Send(&backuppb.StreamBlocksResponse{
			Height: height,
			Data:   data,
		}); err != nil {
			return err
		}
	}

	return nil
}

var _ backuppbconnect.BackupServiceHandler = (*backupServer)(nil)
