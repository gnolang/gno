package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/stretchr/testify/require"
)

const gracefulShutdown = time.Second * 5

type ProcessNodeConfig struct {
	ValidatorKey ed25519.PrivKeyEd25519 `json:"priv"`
	Verbose      bool                   `json:"verbose"`
	DBDir        string                 `json:"dbdir"`
	RootDir      string                 `json:"rootdir"`
	Genesis      *MarshalableGenesisDoc `json:"genesis"`
	TMConfig     *tmcfg.Config          `json:"tm"`
}

type ProcessConfig struct {
	Node *ProcessNodeConfig

	// These parameters are not meant to be passed to the process
	CoverDir       string
	Stderr, Stdout io.Writer
}

func (i ProcessConfig) validate() error {
	if i.Node.TMConfig == nil {
		return errors.New("no tm config set")
	}

	if i.Node.Genesis == nil {
		return errors.New("no genesis is set")
	}

	return nil
}

// RunNode initializes and runs a gnoland node with the provided configuration.
func RunNode(ctx context.Context, pcfg *ProcessNodeConfig, stdout, stderr io.Writer) error {
	// Setup logger based on verbosity
	var handler slog.Handler
	if pcfg.Verbose {
		handler = slog.NewTextHandler(stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewTextHandler(stdout, &slog.HandlerOptions{Level: slog.LevelError})
	}
	logger := slog.New(handler)

	// Initialize database
	db, err := initDatabase(pcfg.DBDir)
	if err != nil {
		return err
	}
	defer db.Close() // ensure db is close

	nodecfg := TestingMinimalNodeConfig(pcfg.RootDir)

	// Configure validator if provided
	if len(pcfg.ValidatorKey) > 0 && !isAllZero(pcfg.ValidatorKey) {
		nodecfg.PrivValidator = bft.NewMockPVWithPrivKey(pcfg.ValidatorKey)
	}
	pk := nodecfg.PrivValidator.PubKey()

	// Setup node configuration
	nodecfg.DB = db
	nodecfg.TMConfig.DBPath = pcfg.DBDir
	nodecfg.TMConfig = pcfg.TMConfig
	// Ensure WAL is disabled for tests. WALDisabled has `json:"-"` tag,
	// so it's lost when config is serialized to JSON for subprocess communication.
	nodecfg.TMConfig.Consensus.WALDisabled = true
	nodecfg.Genesis = pcfg.Genesis.ToGenesisDoc()
	nodecfg.Genesis.Validators = []bft.GenesisValidator{
		{
			Address: pk.Address(),
			PubKey:  pk,
			Power:   10,
			Name:    "self",
		},
	}

	// Create and start the node
	node, err := gnoland.NewInMemoryNode(logger, nodecfg)
	if err != nil {
		return fmt.Errorf("failed to create new in-memory node: %w", err)
	}

	if err := node.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}
	defer node.Stop()

	// Get the validator public key
	ourAddress := nodecfg.PrivValidator.PubKey().Address()

	// Determine if the node is a validator
	isValidator := slices.ContainsFunc(nodecfg.Genesis.Validators, func(val bft.GenesisValidator) bool {
		return val.Address == ourAddress
	})

	lisnAddress := node.Config().RPC.ListenAddress
	if isValidator {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for the node to start: %w", ctx.Err())
		case <-node.Ready():
		}
	}

	// Write READY signal to stdout
	signalWriteReady(stdout, lisnAddress)

	<-ctx.Done()
	return node.Stop()
}

type NodeProcess interface {
	Stop() error
	Address() string
}

type nodeProcess struct {
	cmd     *exec.Cmd
	address string

	stopOnce sync.Once
	stopErr  error
}

func (n *nodeProcess) Address() string {
	return n.address
}

func (n *nodeProcess) Stop() error {
	n.stopOnce.Do(func() {
		// Send SIGTERM to the process
		if err := n.cmd.Process.Signal(os.Interrupt); err != nil {
			n.stopErr = fmt.Errorf("error sending `SIGINT` to the node: %w", err)
			return
		}

		// Optionally wait for the process to exit
		if _, err := n.cmd.Process.Wait(); err != nil {
			n.stopErr = fmt.Errorf("process exited with error: %w", err)
			return
		}
	})

	return n.stopErr
}

// RunNodeProcess runs the binary at the given path with the provided configuration.
func RunNodeProcess(ctx context.Context, cfg ProcessConfig, name string, args ...string) (NodeProcess, error) {
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}

	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Marshal the configuration to JSON
	nodeConfigData, err := json.Marshal(cfg.Node)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Create and configure the command to execute the binary
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = bytes.NewReader(nodeConfigData)

	if cfg.CoverDir != "" {
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+cfg.CoverDir)
	}

	// Redirect all errors into a buffer
	cmd.Stderr = os.Stderr
	if cfg.Stderr != nil {
		cmd.Stderr = cfg.Stderr
	}

	// Create pipes for stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	address, err := waitForProcessReady(ctx, stdoutPipe, cfg.Stdout)
	if err != nil {
		cmd.Process.Kill() // kill process if it isn't ready yet
		return nil, fmt.Errorf("waiting for readiness: %w", err)
	}

	return &nodeProcess{
		cmd:     cmd,
		address: address,
	}, nil
}

type nodeInMemoryProcess struct {
	address string

	stopOnce    sync.Once
	stopErr     error
	stop        context.CancelFunc
	ccNodeError chan error
}

func (n *nodeInMemoryProcess) Address() string {
	return n.address
}

func (n *nodeInMemoryProcess) Stop() error {
	n.stopOnce.Do(func() {
		n.stop()
		var err error
		select {
		case err = <-n.ccNodeError:
		case <-time.After(time.Second * 5):
			err = fmt.Errorf("timeout while waiting for node to stop")
		}

		if err != nil {
			n.stopErr = fmt.Errorf("unable to node gracefully: %w", err)
		}
	})

	return n.stopErr
}

func RunInMemoryProcess(ctx context.Context, cfg ProcessConfig) (NodeProcess, error) {
	ctx, cancel := context.WithCancel(ctx)

	out, in := io.Pipe()
	ccStopErr := make(chan error, 1)
	go func() {
		defer close(ccStopErr)
		defer cancel()

		err := RunNode(ctx, cfg.Node, in, cfg.Stderr)
		if err != nil {
			fmt.Fprintf(cfg.Stderr, "run node failed: %v", err)
		}

		ccStopErr <- err
	}()

	address, err := waitForProcessReady(ctx, out, cfg.Stdout)
	if err == nil { // ok
		return &nodeInMemoryProcess{
			address:     address,
			stop:        cancel,
			ccNodeError: ccStopErr,
		}, nil
	}

	cancel()

	select {
	case err = <-ccStopErr: // return node error in priority
	default:
	}

	return nil, err
}

func RunMain(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	// Read the configuration from standard input
	configData, err := io.ReadAll(stdin)
	if err != nil {
		// log.Fatalf("error reading stdin: %v", err)
		return fmt.Errorf("error reading stdin: %w", err)
	}

	// Unmarshal the JSON configuration
	var cfg ProcessNodeConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
		// log.Fatalf("error unmarshaling JSON: %v", err)
	}

	// Run the node
	ccErr := make(chan error, 1)
	go func() {
		ccErr <- RunNode(ctx, &cfg, stdout, stderr)
		close(ccErr)
	}()

	// Wait for the node to gracefully terminate
	<-ctx.Done()

	// Attempt graceful shutdown
	select {
	case <-time.After(gracefulShutdown):
		return fmt.Errorf("unable to gracefully stop the node, exiting now")
	case err = <-ccErr: // done
	}

	return err
}

func runTestingNodeProcess(t TestingTS, ctx context.Context, pcfg ProcessConfig) NodeProcess {
	bin, err := os.Executable()
	require.NoError(t, err)
	args := []string{
		"-test.run=^$",
		"-run-node-process",
	}

	if pcfg.CoverDir != "" && testing.CoverMode() != "" {
		args = append(args, "-test.gocoverdir="+pcfg.CoverDir)
	}

	node, err := RunNodeProcess(ctx, pcfg, bin, args...)
	require.NoError(t, err)

	return node
}

// initDatabase initializes the database based on the provided directory configuration.
func initDatabase(dbDir string) (db.DB, error) {
	if dbDir == "" {
		return memdb.NewMemDB(), nil
	}

	data, err := goleveldb.NewGoLevelDB("testdb", dbDir)
	if err != nil {
		return nil, fmt.Errorf("unable to init database in %q: %w", dbDir, err)
	}

	return data, nil
}

func signalWriteReady(w io.Writer, address string) error {
	_, err := fmt.Fprintf(w, "READY:%s\n", address)
	return err
}

func signalReadReady(line string) (string, bool) {
	var address string
	if _, err := fmt.Sscanf(line, "READY:%s", &address); err == nil {
		return address, true
	}
	return "", false
}

// waitForProcessReady waits for the process to signal readiness and returns the address.
func waitForProcessReady(ctx context.Context, stdoutPipe io.Reader, out io.Writer) (string, error) {
	var address string

	cReady := make(chan error, 2)
	go func() {
		defer close(cReady)

		scanner := bufio.NewScanner(stdoutPipe)
		ready := false
		for scanner.Scan() {
			line := scanner.Text()

			if !ready {
				if addr, ok := signalReadReady(line); ok {
					address = addr
					ready = true
					cReady <- nil
				}
			}

			fmt.Fprintln(out, line)
		}

		if err := scanner.Err(); err != nil {
			cReady <- fmt.Errorf("error reading stdout: %w", err)
		} else {
			cReady <- fmt.Errorf("process exited without 'READY'")
		}
	}()

	select {
	case err := <-cReady:
		return address, err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// isAllZero checks if a 64-byte key consists entirely of zeros.
func isAllZero(key [64]byte) bool {
	for _, v := range key {
		if v != 0 {
			return false
		}
	}
	return true
}

// XXX why is this needed?
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
