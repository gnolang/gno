// Command gnodrand relays drand evmnet beacons to gno.land/r/drand
// for every round that has pending requests.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gnolang/gno/contribs/gnodrand/internal/relayer"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type config struct {
	remote       string
	chainID      string
	pkgPath      string
	home         string
	keyName      string
	feePerBeacon string
	gasPerBeacon int64
	maxBatch     int
	interval     time.Duration
	mirrors      string
}

func (c *config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.remote, "remote", "https://rpc.gno.land:443", "gno.land RPC endpoint")
	fs.StringVar(&c.chainID, "chain-id", "gnoland-1", "gno.land chain id")
	fs.StringVar(&c.pkgPath, "pkgpath", "gno.land/r/drand/v0", "randomness realm to serve")
	fs.StringVar(&c.home, "home", gnoenv.HomeDir(), "gnokey home, used with -key")
	fs.StringVar(&c.keyName, "key", "", "gnokey key name or address (password in GNODRAND_PASSWORD)")
	fs.StringVar(&c.feePerBeacon, "fee-per-beacon", "22000ugnot", "fee per submitted beacon; the whole tx fee is charged, so keep it at gas-per-beacon x gas price")
	fs.Int64Var(&c.gasPerBeacon, "gas-per-beacon", 22_000_000, "gas wanted per submitted beacon (one Submit uses about 18M)")
	fs.IntVar(&c.maxBatch, "max-batch", 10, "max beacons per transaction")
	fs.DurationVar(&c.interval, "interval", 3*time.Second, "poll interval (the evmnet period is 3s)")
	fs.StringVar(&c.mirrors, "drand", strings.Join(relayer.DefaultMirrors, ","), "comma-separated drand HTTP mirrors, tried in order")
}

func main() {
	cfg := &config{}
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "gnodrand [flags]",
			ShortHelp:  "relay drand evmnet beacons to gno.land",
			LongHelp: `Polls gno.land/r/drand for requested rounds, fetches them from drand,
verifies them locally and submits them. The signing key comes from
GNODRAND_MNEMONIC, or from -key with its password in GNODRAND_PASSWORD.`,
		},
		cfg,
		func(ctx context.Context, _ []string) error { return run(ctx, cfg) },
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cmd.Execute(ctx, os.Args[1:])
}

func run(ctx context.Context, cfg *config) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	fee, err := std.ParseCoin(cfg.feePerBeacon)
	if err != nil {
		return fmt.Errorf("invalid -fee-per-beacon: %w", err)
	}
	signer, err := newSigner(cfg.home, cfg.keyName, cfg.chainID)
	if err != nil {
		return err
	}
	info, err := signer.Info()
	if err != nil {
		return err
	}
	rpc, err := rpcclient.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("rpc client: %w", err)
	}

	r := &relayer.Relayer{
		Source: relayer.HTTPSource{
			Mirrors: strings.Split(cfg.mirrors, ","),
			Client:  &http.Client{Timeout: 5 * time.Second},
		},
		Chain: relayer.GnoChain{
			Client:       &gnoclient.Client{Signer: signer, RPCClient: rpc},
			PkgPath:      cfg.pkgPath,
			FeePerBeacon: fee,
			GasPerBeacon: cfg.gasPerBeacon,
		},
		Log:      log,
		MaxBatch: cfg.maxBatch,
	}

	log.Info("relayer started", "remote", cfg.remote, "chain_id", cfg.chainID, "pkgpath", cfg.pkgPath, "address", info.GetAddress().String())
	if err := r.Run(ctx, cfg.interval); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func newSigner(home, keyName, chainID string) (gnoclient.Signer, error) {
	if m := os.Getenv("GNODRAND_MNEMONIC"); m != "" {
		return gnoclient.SignerFromBip39(m, chainID, "", 0, 0)
	}
	if keyName == "" {
		return nil, errors.New("set GNODRAND_MNEMONIC or -key")
	}
	kb, err := keys.NewKeyBaseFromDir(home)
	if err != nil {
		return nil, fmt.Errorf("open keybase: %w", err)
	}
	s := gnoclient.SignerFromKeybase{
		Keybase:  kb,
		Account:  keyName,
		Password: os.Getenv("GNODRAND_PASSWORD"),
		ChainID:  chainID,
	}
	return s, s.Validate()
}
