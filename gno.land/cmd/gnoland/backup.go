package main

import (
	"archive/tar"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"connectrpc.com/connect"
	backup "github.com/gnolang/gno/tm2/pkg/bft/node/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/node/backuppb/backupconnect"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/klauspost/compress/zstd"
)

type backupCfg struct {
	remote string
}

func newBackupCmd(io commands.IO) *commands.Command {
	cfg := &backupCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "backup",
			ShortUsage: "backup [flags]",
			ShortHelp:  "backups the Gnoland blockchain node",
			LongHelp:   "Backups the Gnoland blockchain node, with accompanying setup",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBackup(ctx, cfg, io)
		},
	)
}

func (c *backupCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"http://localhost:4242",
		"backup service remote",
	)

}

func execBackup(ctx context.Context, c *backupCfg, io commands.IO) error {
	client := backupconnect.NewBackupServiceClient(
		http.DefaultClient,
		c.remote,
		connect.WithGRPC(),
	)
	res, err := client.StreamBlocks(
		ctx,
		connect.NewRequest(&backup.StreamBlocksRequest{}),
	)
	if err != nil {
		return err
	}

	const chunkSize = 100

	outdir := "blocks-backup"
	if err := os.MkdirAll(outdir, 0o775); err != nil {
		return err
	}

	height := 1
	chunkStart := height

	latestFP := filepath.Join(outdir, "latest.tar.zst")
	outFile, err := os.OpenFile(latestFP, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o664)
	if err != nil {
		return err
	}
	zstw, err := zstd.NewWriter(outFile)
	if err != nil {
		return err
	}
	w := tar.NewWriter(zstw)

	finalizeChunk := func(last bool) error {
		if err := w.Close(); err != nil {
			return err
		}
		if err := zstw.Close(); err != nil {
			return err
		}
		if err := outFile.Close(); err != nil {
			return err
		}

		chunkFP := filepath.Join(outdir, fmt.Sprintf("%019d-%019d.tm2blocks.tar.zst", chunkStart, height-1))
		if err := os.Rename(latestFP, chunkFP); err != nil {
			return err
		}

		io.ErrPrintln(height - 1)

		if last {
			return nil
		}

		outFile, err = os.OpenFile(latestFP, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o664)
		if err != nil {
			return err
		}
		zstw, err = zstd.NewWriter(outFile)
		if err != nil {
			return err
		}
		w = tar.NewWriter(zstw)

		chunkStart = height

		return nil
	}

	for {
		ok := res.Receive()
		if !ok {
			if res.Err() != nil {
				return err
			}

			return finalizeChunk(true)
		}
		msg := res.Msg()

		header := &tar.Header{
			Name: fmt.Sprintf("%d", msg.Height),
			Size: int64(len(msg.Data)),
			Mode: 0o664,
		}
		if err := w.WriteHeader(header); err != nil {
			return err
		}

		_, err := w.Write(msg.Data)
		if err != nil {
			return err
		}

		height += 1
		if height%chunkSize != 1 {
			continue
		}

		if err := finalizeChunk(false); err != nil {
			return err
		}
	}
}
