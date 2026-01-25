package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
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

func execBackup(ctx context.Context, c *backupCfg, cmdIO commands.IO) (resErr error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), zapcore.AddSync(cmdIO.Out()), config.Level)
	logger := zap.New(core)
	conn, err := grpc.NewClient(c.remote, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
		return
	}

	return backup.WithWriter(c.outDir, c.startHeight, c.endHeight, logger, func(startHeight int64, write func(bytes []byte) error) error {
		client := backuppb.NewBackupServiceClient(conn)
		res, err := client.StreamBlocks(
			ctx,
			&backuppb.StreamBlocksRequest{
				StartHeight: startHeight,
				EndHeight:   c.endHeight,
			},
		)

		if err != nil {
			return fmt.Errorf("open blocks stream: %w", err)
		}

		for {
			res, err := res.Recv()
			// Stream closed, no error
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
			if err := write(res.Data); err != nil {
				return err
			}
		}
	})
}
