package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/tx-archive/backup"
	"github.com/gnolang/tx-archive/backup/client/http"
	"github.com/gnolang/tx-archive/backup/writer"
	"github.com/gnolang/tx-archive/backup/writer/legacy"
	"github.com/gnolang/tx-archive/backup/writer/standard"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
)

const (
	defaultOutputPath = "./backup.jsonl"
	defaultFromBlock  = 1
	defaultToBlock    = -1 // no limit

	defaultRemoteAddress = "http://127.0.0.1:26657"
)

var (
	errInvalidOutputLocation = errors.New("invalid output location")
	errOutputFileExists      = errors.New("output file exists")
	errInvalidRemote         = errors.New("invalid remote address")
)

// backupCfg is the backup command configuration
type backupCfg struct {
	outputPath string
	remote     string

	toBlock   int64 // < 0 means there is no right bound
	fromBlock uint64

	overwrite bool
	legacy    bool
	watch     bool
}

// newBackupCmd creates the backup command
func newBackupCmd() *ffcli.Command {
	cfg := &backupCfg{}

	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "backup",
		ShortUsage: "backup [flags]",
		LongHelp:   "Runs the chain backup service",
		FlagSet:    fs,
		Exec:       cfg.exec,
	}
}

// registerFlags registers the backup command flags
func (c *backupCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.outputPath,
		"output-path",
		defaultOutputPath,
		"the output path for the JSONL chain data",
	)

	fs.StringVar(
		&c.remote,
		"remote",
		defaultRemoteAddress,
		"the JSON-RPC URL of the chain to be backed up",
	)

	fs.Int64Var(
		&c.toBlock,
		"to-block",
		defaultToBlock,
		"the end block number for the backup (inclusive). If <0, latest chain height is used",
	)

	fs.Uint64Var(
		&c.fromBlock,
		"from-block",
		defaultFromBlock,
		"the starting block number for the backup (inclusive)",
	)

	fs.BoolVar(
		&c.overwrite,
		"overwrite",
		false,
		"flag indicating if the output file should be overwritten during backup",
	)

	fs.BoolVar(
		&c.legacy,
		"legacy",
		false,
		"flag indicating if the legacy output format should be used (tx-per-line)",
	)

	fs.BoolVar(
		&c.watch,
		"watch",
		false,
		"flag indicating if the backup should append incoming tx data",
	)
}

// exec executes the backup command
func (c *backupCfg) exec(ctx context.Context, _ []string) error {
	// Make sure the remote address is set
	if c.remote == "" {
		return errInvalidRemote
	}

	// Make sure the output file path is valid
	if c.outputPath == "" {
		return errInvalidOutputLocation
	}

	// Make sure the output file can be overwritten, if it exists
	if _, err := os.Stat(c.outputPath); err == nil && !c.overwrite {
		// File already exists, and the overwrite flag is not set
		return errOutputFileExists
	}

	// Set up the config
	cfg := backup.DefaultConfig()
	cfg.FromBlock = c.fromBlock
	cfg.Watch = c.watch

	if c.toBlock >= 0 {
		to64 := uint64(c.toBlock)
		cfg.ToBlock = &to64
	}

	// Set up the client
	client, err := http.NewClient(c.remote)
	if err != nil {
		return fmt.Errorf("could not create a gno client, %w", err)
	}

	// Set up the logger
	zapLogger, loggerErr := zap.NewDevelopment()
	if loggerErr != nil {
		return fmt.Errorf("unable to create logger, %w", loggerErr)
	}

	logger := newCommandLogger(zapLogger)

	// Set up the writer (file)
	// Open the file for writing
	outputFile, openErr := os.OpenFile(
		c.outputPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0o755,
	)
	if openErr != nil {
		return fmt.Errorf("unable to open file %s, %w", c.outputPath, openErr)
	}

	closeFile := func() error {
		if err := outputFile.Close(); err != nil {
			logger.Error("unable to close output file", "err", err.Error())

			return err
		}

		return nil
	}

	teardown := func() {
		if err := closeFile(); err != nil {
			if removeErr := os.Remove(outputFile.Name()); removeErr != nil {
				logger.Error("unable to remove file", "err", err.Error())
			}
		}
	}

	// Set up the teardown
	defer teardown()

	var w writer.Writer

	if c.legacy {
		w = legacy.NewWriter(outputFile)
	} else {
		w = standard.NewWriter(outputFile)
	}

	// Create the backup service
	service := backup.NewService(
		client,
		w,
		backup.WithLogger(logger),
	)

	// Run the backup service
	if backupErr := service.ExecuteBackup(ctx, cfg); backupErr != nil {
		return fmt.Errorf("unable to execute backup, %w", backupErr)
	}

	return nil
}
