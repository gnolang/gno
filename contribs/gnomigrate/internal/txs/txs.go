package txs

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	errInvalidInputDir  = errors.New("invalid input directory")
	errInvalidOutputDir = errors.New("invalid output directory")
)

type txsCfg struct {
	inputDir  string
	outputDir string
}

// NewTxsCmd creates the migrate txs subcommand
func NewTxsCmd(io commands.IO) *commands.Command {
	cfg := &txsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "txs",
			ShortUsage: "<subcommand> [flags]",
			ShortHelp:  "manages the legacy transaction sheet migrations",
			LongHelp:   "Manages legacy transaction migrations through sheet input files",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return cfg.execMigrate(ctx, io)
		},
	)

	return cmd
}

func (c *txsCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.inputDir,
		"input-dir",
		"",
		"the input directory for the legacy transaction sheets",
	)

	fs.StringVar(
		&c.outputDir,
		"output-dir",
		"",
		"the output directory for the standard transaction sheets",
	)
}

func (c *txsCfg) execMigrate(ctx context.Context, io commands.IO) error {
	// Make sure the dirs are set
	if c.inputDir == "" {
		return errInvalidInputDir
	}

	if c.outputDir == "" {
		return errInvalidOutputDir
	}

	// Make sure the output dir is present
	if err := os.MkdirAll(c.outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create output dir, %w", err)
	}

	return migrateDir(ctx, io, c.inputDir, c.outputDir)
}

// migrateDir migrates the transaction sheet directory
func migrateDir(
	ctx context.Context,
	io commands.IO,
	sourceDir string,
	outputDir string,
) error {
	// Read the sheet directory
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("error reading directory %s, %w", sourceDir, err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil
		default:
			var (
				srcPath  = filepath.Join(sourceDir, entry.Name())
				destPath = filepath.Join(outputDir, entry.Name())
			)

			// Check if a dir is encountered
			if !entry.IsDir() {
				// Make sure the file type is valid
				if !strings.HasSuffix(entry.Name(), ".jsonl") {
					continue
				}

				// Process the tx sheet
				io.Printfln("Migrating %s -> %s", srcPath, destPath)

				if err := processFile(ctx, io, srcPath, destPath); err != nil {
					io.ErrPrintfln("unable to process file %s, %w", srcPath, err)
				}

				continue
			}

			// Ensure destination directory exists
			if err = os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating directory %s, %w", destPath, err)
			}

			// Recursively process the directory
			if err = migrateDir(ctx, io, srcPath, destPath); err != nil {
				io.ErrPrintfln("unable migrate directory %s, %w", srcPath, err)
			}
		}
	}

	return nil
}

// processFile processes the old legacy std.Tx sheet into the new standard gnoland.TxWithMetadata
func processFile(ctx context.Context, io commands.IO, source, destination string) error {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open file, %w", err)
	}
	defer file.Close()

	// Create the destination file
	outputFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("unable to create file, %w", err)
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer(make([]byte, 1_000_000), 2_000_000)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			var (
				tx             std.Tx
				txWithMetadata gnoland.TxWithMetadata
			)

			if err = amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				io.ErrPrintfln("unable to read line, %s", err)

				continue
			}

			// Convert the std.Tx -> gnoland.TxWithMetadata
			txWithMetadata = gnoland.TxWithMetadata{
				Tx:       tx,
				Metadata: nil, // not set
			}

			// Save the new transaction with metadata
			marshaledData, err := amino.MarshalJSON(txWithMetadata)
			if err != nil {
				io.ErrPrintfln("unable to marshal tx, %s", err)

				continue
			}

			if _, err = fmt.Fprintf(outputFile, "%s\n", marshaledData); err != nil {
				io.ErrPrintfln("unable to save to output file, %s", err)
			}
		}
	}

	// Check if there were any scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error encountered during scan, %w", err)
	}

	return nil
}
