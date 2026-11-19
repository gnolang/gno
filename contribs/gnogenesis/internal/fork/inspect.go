package fork

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type inspectCfg struct{}

func newInspectCmd(io commands.IO) *commands.Command {
	cfg := &inspectCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "inspect",
			ShortUsage: "inspect <genesis.json>",
			ShortHelp:  "print a provenance report for a hardfork genesis.json",
			LongHelp: `Read a genesis.json and group its appState.Txs by GnoTxMetadata.Source.

Categories:
  base         Inherited from the source chain's genesis (appState.Txs in the
               base genesis used by 'gnogenesis fork generate').
  historical   Source-chain tx history applied unmodified during replay.
  patched      Historical tx whose body was rewritten by '--patch-txs'; the
               report includes the patch reason and original-tx pointer.
  migration    Tx contributed by '--migration-tx' (addpkg, valoper-seed,
               rotation scripts, etc.); the report shows the per-tx reason.
  unannotated  Tx without a Source field — typically a genesis produced by an
               older toolchain. Listed for completeness so the operator can
               spot pre-provenance artifacts.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execInspect(cfg, io, args)
		},
	)
}

func (c *inspectCfg) RegisterFlags(_ *flag.FlagSet) {}

func execInspect(_ *inspectCfg, io commands.IO, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: inspect <genesis.json>")
	}
	path := args[0]

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var doc bftypes.GenesisDoc
	if err := amino.UnmarshalJSON(data, &doc); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	state, ok := doc.AppState.(gnoland.GnoGenesisState)
	if !ok {
		raw, err := amino.MarshalJSON(doc.AppState)
		if err != nil {
			return fmt.Errorf("re-encode appState: %w", err)
		}
		if err := amino.UnmarshalJSON(raw, &state); err != nil {
			return fmt.Errorf("decode appState as GnoGenesisState: %w", err)
		}
	}

	io.Printf("%s\n", inspectReport(&state))
	return nil
}

// inspectReport returns a multi-line provenance report. Pulled out for unit
// testing without going through file I/O.
func inspectReport(state *gnoland.GnoGenesisState) string {
	groups := map[string][]int{
		gnoland.SourceBase:       nil,
		gnoland.SourceHistorical: nil,
		gnoland.SourcePatched:    nil,
		gnoland.SourceMigration:  nil,
	}
	var unannotated []int

	for i, tx := range state.Txs {
		if tx.Metadata == nil || tx.Metadata.Source == "" {
			unannotated = append(unannotated, i)
			continue
		}
		if _, known := groups[tx.Metadata.Source]; known {
			groups[tx.Metadata.Source] = append(groups[tx.Metadata.Source], i)
		} else {
			unannotated = append(unannotated, i)
		}
	}

	var b strings.Builder
	fmt.Fprintln(&b, "=== Provenance ===")
	fmt.Fprintf(&b, "Base genesis: %d\n", len(groups[gnoland.SourceBase]))
	fmt.Fprintf(&b, "Historical:   %d\n", len(groups[gnoland.SourceHistorical]))
	fmt.Fprintf(&b, "Patched:      %d\n", len(groups[gnoland.SourcePatched]))
	fmt.Fprintf(&b, "Migration:    %d\n", len(groups[gnoland.SourceMigration]))
	if len(unannotated) > 0 {
		fmt.Fprintf(&b, "Unannotated:  %d\n", len(unannotated))
	}
	fmt.Fprintf(&b, "Total:        %d\n", len(state.Txs))

	if len(groups[gnoland.SourcePatched]) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Patched txs:")
		for _, i := range groups[gnoland.SourcePatched] {
			meta := state.Txs[i].Metadata
			senderInfo := "—"
			if len(meta.SignerInfo) > 0 {
				senderInfo = fmt.Sprintf("sender=%s seq=%d", meta.SignerInfo[0].Address, meta.SignerInfo[0].Sequence)
			}
			fmt.Fprintf(&b, "  - h=%d %s: %q\n", meta.BlockHeight, senderInfo, meta.Note)
		}
	}

	if len(groups[gnoland.SourceMigration]) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Migration txs:")
		for _, i := range groups[gnoland.SourceMigration] {
			note := state.Txs[i].Metadata.Note
			if note == "" {
				note = "(no reason)"
			}
			fmt.Fprintf(&b, "  - %s\n", note)
		}
	}

	return b.String()
}
