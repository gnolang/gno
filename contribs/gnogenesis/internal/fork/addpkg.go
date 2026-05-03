package fork

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Default deployer for hardfork addpkg txs. The hardfork ceremony runs
// with --skip-genesis-sig-verification=true, so the actual signature
// is irrelevant; what matters is the deployer address that becomes
// the package's owner. Mirrors gnogenesis txs add packages' default
// account, which gnoland-1's genesis used.
const defaultDeployerAddr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

// genesisDeployFee mirrors gno.land/cmd/start.go's genesis fee. Each
// addpkg tx in the .jsonl carries the same fee for parity with the
// fee used when these realms were originally deployed; under
// --skip-genesis-sig-verification the fee is also un-checked.
var genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1)))

type addpkgCfg struct {
	output      string
	deployerStr string
}

// newAddpkgCmd builds a deterministic .jsonl of MsgAddPackage txs
// from one or more local package directories. Used during a hardfork
// ceremony to deploy realms that don't exist on the source chain
// (e.g., r/sys/validators/v3 when forking from gnoland-1, where v3
// was added post-source-launch).
//
// Output format matches what `gnogenesis fork generate --migration-tx`
// expects: gnoland.TxWithMetadata, one amino-JSON line per tx,
// BlockHeight forced to 0 by readMigrationTxs at consume time.
func newAddpkgCmd(io commands.IO) *commands.Command {
	cfg := &addpkgCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "addpkg",
			ShortUsage: "addpkg [flags] <pkgdir> [<pkgdir>...]",
			ShortHelp:  "build a .jsonl of MsgAddPackage txs from local package dirs",
			LongHelp: `Build a deterministic .jsonl of MsgAddPackage migration txs from one
or more local package directories. Output is intended for
'gnogenesis fork generate --migration-tx' as a prerequisite step
when the source chain doesn't have a needed realm deployed.

Example: forking from gnoland-1 (which doesn't have v3) to a chain
that requires r/sys/validators/v3:

  gnogenesis fork addpkg \
      --output addpkg-v3.jsonl \
      examples/gno.land/r/sys/validators/v3
  gnogenesis fork generate \
      --source ... \
      --migration-tx addpkg-v3.jsonl \
      --migration-tx valoper-seed.jsonl \
      ...

Each emitted tx is a MsgAddPackage with:
  - Caller = --deployer (default: gnoland-1's test1 account)
  - Package = LoadPackagesFromDir(<pkgdir>) — recursive, includes sub-realms
  - Metadata.BlockHeight = 0 (genesis-mode)
  - Signatures = [] (consumer runs with --skip-genesis-sig-verification)

Output is written in the order packages are loaded; LoadPackagesFromDir
sorts by pkgpath internally, so the same input produces a byte-equal
output across runs.`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execAddpkg(ctx, cfg, io, args)
		},
	)
}

func (c *addpkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.output, "output", "", "output .jsonl path (required)")
	fs.StringVar(&c.deployerStr, "deployer", defaultDeployerAddr,
		"bech32 address that becomes the package owner; defaults to gnoland-1's test1 account")
}

func execAddpkg(_ context.Context, cfg *addpkgCfg, io commands.IO, args []string) error {
	if cfg.output == "" {
		return errors.New("--output is required")
	}
	if len(args) == 0 {
		return errors.New("at least one pkgdir argument is required")
	}

	deployer, err := crypto.AddressFromBech32(cfg.deployerStr)
	if err != nil {
		return fmt.Errorf("invalid --deployer %q: %w", cfg.deployerStr, err)
	}

	var allTxs []gnoland.TxWithMetadata
	for _, dir := range args {
		txs, err := gnoland.LoadPackagesFromDir(dir, deployer, genesisDeployFee)
		if err != nil {
			return fmt.Errorf("LoadPackagesFromDir %q: %w", dir, err)
		}
		// Ensure each tx has Metadata.BlockHeight=0 explicitly,
		// even though readMigrationTxs forces it at consume time —
		// keeps the .jsonl self-describing.
		for i := range txs {
			if txs[i].Metadata == nil {
				txs[i].Metadata = &gnoland.GnoTxMetadata{}
			}
			txs[i].Metadata.BlockHeight = 0
			// Strip signatures: consumer runs with
			// --skip-genesis-sig-verification.
			txs[i].Tx.Signatures = []std.Signature{}
		}
		allTxs = append(allTxs, txs...)
	}

	var buf strings.Builder
	for _, tx := range allTxs {
		line, err := amino.MarshalJSON(tx)
		if err != nil {
			return fmt.Errorf("marshal tx: %w", err)
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	if err := os.WriteFile(cfg.output, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", cfg.output, err)
	}

	io.Printfln("wrote %d MsgAddPackage txs to %s", len(allTxs), cfg.output)
	return nil
}
