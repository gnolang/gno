package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb/backuppbconnect"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/v1"
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
		"http://localhost:4242",
		"Backup service remote.",
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

func execBackup(ctx context.Context, c *backupCfg, io commands.IO) (resErr error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), zapcore.AddSync(io.Out()), config.Level)
	logger := zap.New(core)

	return backup.WithWriter(c.outDir, c.startHeight, c.endHeight, logger, func(startHeight int64, write func(bytes []byte) error) error {
		client := backuppbconnect.NewBackupServiceClient(
			http.DefaultClient,
			c.remote,
			connect.WithGRPC(),
		)
		res, err := client.StreamBlocks(
			ctx,
			connect.NewRequest(&backuppb.StreamBlocksRequest{
				StartHeight: startHeight,
				EndHeight:   c.endHeight,
			}),
		)
		if err != nil {
			return fmt.Errorf("open blocks stream: %w", err)
		}

		for {
			ok := res.Receive()
			if !ok {
				return res.Err()
			}

			if err := write(res.Msg().Data); err != nil {
				return err
			}
		}
	})
}
