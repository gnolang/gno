// gpao (gno package-approver oracle) is a small off-chain approval daemon for
// gno.land chains running with the "inert" code-submission policy (PR #5888).
//
// It watches new blocks, extracts MsgAddPackage transactions, runs the Gno
// typechecker on each submitted package off-chain, and — if the package is
// well-typed — broadcasts a MsgEnablePackage transaction signed by an approver
// key to activate the package on-chain.
//
// The oracle is untrusted for correctness: the chain re-runs the typechecker
// at MsgEnablePackage time and rejects ill-typed code. gpao only decides
// *which* pending packages get proposed for activation, and *when*.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const (
	defaultRemote       = "http://127.0.0.1:26657"
	defaultGasFee       = "1000000ugnot"
	defaultGasWanted    = int64(20_000_000)
	defaultPollInterval = time.Second
)

func main() {
	cfg := &config{}
	io := commands.NewDefaultIO()

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gpao",
			ShortUsage: "gpao [flags]",
			ShortHelp:  "watch a chain, typecheck submitted packages, and approve the good ones",
			LongHelp: "gpao (gno package-approver oracle) watches new blocks for MsgAddPackage " +
				"transactions on a gno.land chain running the \"inert\" code-submission policy, " +
				"typechecks each submitted package off-chain, and broadcasts a MsgEnablePackage " +
				"transaction (signed by an approver key) to activate packages that pass. The chain " +
				"re-verifies on enable.",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execOracle(ctx, cfg, io)
		},
	)

	// Cancel on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cmd.Execute(ctx, os.Args[1:])
}

// config holds the oracle's runtime configuration.
type config struct {
	remote  string
	chainID string

	// Signing: the approver key is loaded from the local gnokey keystore
	// (home + key). The key's address must be in the chain's PkgApprovers param.
	home string
	key  string
	// mnemonic is a dev-only fallback read from $GPAO_MNEMONIC; prefer the
	// keystore. When set it takes precedence over home/key.
	mnemonic string

	gnoRoot      string
	gasFee       string
	gasWanted    int64
	pollInterval time.Duration
	startHeight  int64
}

func (c *config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.remote, "remote", defaultRemote,
		"RPC address of the gno.land node to watch")
	fs.StringVar(&c.chainID, "chain-id", "",
		"chain ID used to sign approval transactions (required)")
	fs.StringVar(&c.home, "home", gnoenv.HomeDir(),
		"gnokey keystore directory holding the approver key")
	fs.StringVar(&c.key, "key", "",
		"name or bech32 address of the approver key in the keystore; "+
			"its address must be listed in the chain's vm PkgApprovers param")
	fs.StringVar(&c.gnoRoot, "gno-root", gnoenv.RootDir(),
		"path to the gno repository root, used to resolve stdlibs and examples for typechecking")
	fs.StringVar(&c.gasFee, "gas-fee", defaultGasFee,
		"gas fee for approval transactions")
	fs.Int64Var(&c.gasWanted, "gas-wanted", defaultGasWanted,
		"gas wanted for approval transactions")
	fs.DurationVar(&c.pollInterval, "poll-interval", defaultPollInterval,
		"how often to poll the node for new blocks")
	fs.Int64Var(&c.startHeight, "start-height", 0,
		"block height to start watching from (0 = start from the current tip)")
}

func (c *config) validate() error {
	if c.chainID == "" {
		return fmt.Errorf("--chain-id is required")
	}
	if c.mnemonic == "" && c.key == "" {
		return fmt.Errorf("--key is required (name or address of the approver key in --home)")
	}
	if c.gnoRoot == "" {
		return fmt.Errorf("--gno-root is required (could not auto-detect the gno root)")
	}
	if c.gasWanted <= 0 {
		return fmt.Errorf("--gas-wanted must be > 0")
	}
	return nil
}

func execOracle(ctx context.Context, cfg *config, io commands.IO) error {
	// Dev-only signing fallback: a mnemonic via env avoids needing a keystore.
	cfg.mnemonic = os.Getenv("GPAO_MNEMONIC")

	if err := cfg.validate(); err != nil {
		return err
	}

	oracle, err := newOracle(*cfg, io)
	if err != nil {
		return err
	}

	io.Println("gpao: approver", oracle.approver.String(),
		"watching", cfg.remote, "chain", cfg.chainID)

	return oracle.run(ctx)
}
