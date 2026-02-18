package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/v1"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gorilla/websocket"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cfg := &backupCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "[flags]",
			ShortHelp:  "efficiently backup tm2 blocks",
			LongHelp:   "Efficiently backup tendermint2 blocks",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBackup(ctx, cfg, io)
		},
	)

	return cmd
}

type backupCfg struct {
	remote      string
	outDir      string
	startHeight int64
	endHeight   int64
}

func (c *backupCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"ws://localhost:26657/websocket",
		"Node RPC service remote.",
	)
	fs.StringVar(
		&c.outDir,
		"o",
		"blocks-backup",
		"Output directory.",
	)
	fs.Int64Var(
		&c.startHeight,
		"start",
		0,
		fmt.Sprintf("Start height. Will be aligned at a multiple of %d blocks. This option can't be used when resuming from an existing output directory.", backup.ChunkSize),
	)
	fs.Int64Var(
		&c.endHeight,
		"end",
		0,
		"End height, inclusive. Use 0 for latest height.",
	)
}

func execBackup(ctx context.Context, c *backupCfg, cmdIO commands.IO) (resErr error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), zapcore.AddSync(cmdIO.Out()), config.Level)
	logger := zap.New(core)

	url := c.remote
	// we need to add /websocket as the websocket handler is in this path
	if !strings.HasSuffix(url, "/websocket"){
		url += "/websocket" 
	}
	logger.Info("connecting to RPC", zap.String("url", url))

	connection, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("dial websocket RPC: %w", err)
	}
	defer connection.Close()

	return backup.WithWriter(c.outDir, c.startHeight, c.endHeight, logger, func(startHeight int64, write func(block *types.Block) error) error {
		req, err := rpctypes.MapToRequest(
			rpctypes.JSONRPCStringID("tm2backup"),
			"backup",
			map[string]any{
				"start": startHeight,
				"end":   c.endHeight,
			},
		)
		if err != nil {
			return fmt.Errorf("create backup request: %w", err)
		}

		reqBz, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshal backup request: %w", err)
		}

		if err := connection.WriteMessage(websocket.TextMessage, reqBz); err != nil {
			return fmt.Errorf("send backup request: %w", err)
		}

		for {
			_, msg, err := connection.ReadMessage()
			if err != nil {
				return fmt.Errorf("read backup response: %w", err)
			}

			var resp rpctypes.RPCResponse
			if err := json.Unmarshal(msg, &resp); err != nil {
				return fmt.Errorf("parse backup response: %w", err)
			}

			if resp.Error != nil {
				return resp.Error
			}

			var backupBlock ctypes.ResultBackupBlock
			err = amino.UnmarshalJSON(resp.Result, &backupBlock)
			if err != nil {
				return err
			}

			if backupBlock.Done {
				logger.Info("Stream completed", zap.Int64("height", backupBlock.Height))
				return nil
			}

			if backupBlock.Block == nil {
				return fmt.Errorf("block not found on height %d", backupBlock.Height)
			}

			err = write(backupBlock.Block)
			if err != nil {
				return err
			}
		}
	})
}
