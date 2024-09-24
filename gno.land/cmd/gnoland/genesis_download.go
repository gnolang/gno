package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

var (
	errNoChainID         = errors.New("no chain ID specified")
	errChainNotSupported = errors.New("chain ID not supported")
)

const test4ID = "test4"

const deploymentPathFormat = "https://raw.githubusercontent.com/gnolang/gno/refs/heads/master/misc/deployments/%s.gno.land/genesis.json"

var genesisSHAMap = map[string]string{
	test4ID: "beb781dffc09b96e3114fb7439fa85c4fe8ea796f64ec0cd3801a6b518ab023c",
}

type downloadCfg struct {
	commonCfg
}

// newDownloadCmd creates the genesis download subcommand
func newDownloadCmd(io commands.IO) *commands.Command {
	cfg := &downloadCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "download",
			ShortUsage: "download <chain-id>",
			ShortHelp:  "downloads the specific gno chain's genesis.json",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execGenesisDownload(ctx, cfg, io, args)
		},
	)
}

func (c *downloadCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}

func execGenesisDownload(
	ctx context.Context,
	cfg *downloadCfg,
	io commands.IO,
	args []string,
) error {
	// Make sure the chain ID is specified
	if len(args) != 1 {
		return errNoChainID
	}

	// Make sure the chain ID is supported
	chainID := args[0]

	genesisSHA, exists := genesisSHAMap[chainID]
	if !exists {
		return errChainNotSupported
	}

	// Fetch the genesis file
	downloadURL := fmt.Sprintf(deploymentPathFormat, chainID)

	if err := downloadFile(ctx, downloadURL, cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to download genesis.json, %w", err)
	}

	// Verify the SHA
	computedSHA, err := computeSHA256(cfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to compute genesis.json SHA, %w", err)
	}

	if genesisSHA != computedSHA {
		return fmt.Errorf("expected genesis SHA %s, got %s", genesisSHA, computedSHA)
	}

	io.Printfln("Successfully downloaded %s genesis.json", chainID)
	io.Printfln("SHA256: %s", computedSHA)

	return nil
}

// downloadFile downloads the file from the specified URL
func downloadFile(
	ctx context.Context,
	url string,
	outPath string,
) error {
	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("unable to create GET request, %w", err)
	}

	// Execute the fetch
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make GET request, %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	tempPath := "genesis.json.tmp"
	f, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("unable to create file, %w", err)
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("error while downloading, %w", err)
	}

	// After the file is downloaded, rename the temporary file
	if err = os.Rename(tempPath, outPath); err != nil {
		return fmt.Errorf("unable to rename file, %w", err)
	}

	return nil
}

// computeSHA256 computes the SHA-256 hash of a file
func computeSHA256(path string) (string, error) {
	// Open the file
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("unable to open file, %w", err)
	}
	defer f.Close()

	// Create the hasher
	h := sha256.New()

	// Hash the file
	if _, err = io.Copy(h, bufio.NewReader(f)); err != nil {
		return "", fmt.Errorf("unable to hash file, %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
