package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gnolang/gno/tm2/pkg/amino"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

type MarshalableGenesisDoc bft.GenesisDoc

func NewMarshalableGenesisDoc(doc *bft.GenesisDoc) *MarshalableGenesisDoc {
	m := MarshalableGenesisDoc(*doc)
	return &m
}

func (m *MarshalableGenesisDoc) MarshalJSON() ([]byte, error) {
	doc := (*bft.GenesisDoc)(m)
	return amino.MarshalJSON(doc)
}

func (m *MarshalableGenesisDoc) UnmarshalJSON(data []byte) (err error) {
	doc, err := bft.GenesisDocFromJSON(data)
	if err != nil {
		return err
	}

	*m = MarshalableGenesisDoc(*doc)
	return
}

// Cast back to the original bft.GenesisDoc.
func (m *MarshalableGenesisDoc) ToGenesisDoc() *bft.GenesisDoc {
	return (*bft.GenesisDoc)(m)
}

type Config struct {
	ValidatorKey ed25519.PrivKeyEd25519 `json:"priv"`
	Verbose      bool                   `json:"verbose"`
	DBDir        string                 `json:"dbdir"`
	RootDir      string                 `json:"rootdir"`
	Genesis      *MarshalableGenesisDoc `json:"genesis"`
	TMConfig     *tmcfg.Config          `json:"tm"`
}

func (i Config) validate() error {
	if i.TMConfig == nil {
		return errors.New("no tm config set")
	}

	if i.Genesis == nil {
		return errors.New("no genesis is set")
	}

	return nil
}

// ExecuteNode runs the binary at the given path with the provided configuration.
// It marshals the configuration to JSON and passes it to the binary via stdin.
// The function waits for "READY:<address>" on stdout and returns the address if successful,
// or kills the process and returns an error if "READY" is not received within the context deadline.
func ExecuteNode(ctx context.Context, binaryPath string, cfg *Config, out io.Writer) (string, *exec.Cmd, error) {
	if err := cfg.validate(); err != nil {
		return "", nil, err
	}

	// Marshal the configuration to JSON
	configData, err := json.Marshal(cfg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Create the command to execute the binary
	cmd := exec.Command(binaryPath)
	cmd.Env = os.Environ()

	// Set the standard input to the JSON data
	cmd.Stdin = bytes.NewReader(configData)

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start command: %w", err)
	}

	readyChan := make(chan error, 1)
	var address string

	// Wait for ready signal
	go func() {
		var ready bool
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if !ready {
				if _, err := fmt.Sscanf(line, "READY:%s", &address); err == nil {
					readyChan <- nil
					ready = true
					continue
				}
			}

			fmt.Fprintln(out, line)
		}

		if err := scanner.Err(); err != nil {
			readyChan <- fmt.Errorf("error reading stdout: %w", err)
		} else {
			readyChan <- fmt.Errorf("process exited without 'READY'")
		}
	}()

	// Wait for either the "READY" signal or a context timeout.
	select {
	case err := <-readyChan:
		if err != nil {
			fmt.Fprintf(out, "err: %q\n", err)
			cmd.Process.Kill()
			return "", cmd, err
		}
	case <-ctx.Done():
		cmd.Process.Kill()
		return "", cmd, ctx.Err()
	}

	return address, cmd, nil
}
