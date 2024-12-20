package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// Function to cast back to the original bft.GenesisDoc
func (m *MarshalableGenesisDoc) ToGenesisDoc() *bft.GenesisDoc {
	return (*bft.GenesisDoc)(m)
}

type ForkConfig struct {
	PrivValidator ed25519.PrivKeyEd25519 `json:"priv"`
	DBDir         string                 `json:"dbdir"`
	RootDir       string                 `json:"rootdir"`
	Genesis       *MarshalableGenesisDoc `json:"genesis"`
	TMConfig      *tmcfg.Config          `json:"tm"`
}

// ExecuteForkBinary runs the binary at the given path with the provided configuration.
// It marshals the configuration to JSON and passes it to the binary via stdin.
// The function waits for "READY:<address>" on stdout and returns the address if successful,
// or kills the process and returns an error if "READY" is not received within 10 seconds.
func ExecuteForkBinary(ctx context.Context, binaryPath string, cfg *ForkConfig) (string, *exec.Cmd, error) {
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

	// Goroutine to read stdout and check for "READY"
	go func() {
		var scanned bool
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // Print each line to stdout for logging
			if scanned {
				continue
			}
			if _, err := fmt.Sscanf(line, "READY:%s", &address); err == nil {
				readyChan <- nil
				scanned = true
			}
		}
		if err := scanner.Err(); err != nil {
			readyChan <- fmt.Errorf("error reading stdout: %w", err)
		} else {
			readyChan <- fmt.Errorf("process exited without 'READY'")
		}
	}()

	// Wait for either the "READY" signal or a timeout
	select {
	case err := <-readyChan:
		if err != nil {
			fmt.Println("ERR", err)
			cmd.Process.Kill()
			return "", cmd, err
		}
	case <-ctx.Done():
		cmd.Process.Kill()
		return "", cmd, ctx.Err()
	}

	return address, cmd, nil
}
