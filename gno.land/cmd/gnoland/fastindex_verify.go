package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/commands"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// mainStorePrefix is the rootmulti key prefix of gno.land's main store within
// the shared "gnolang" DB (see rootmulti.constructStore). The bptree fast index
// lives under this prefix.
const mainStorePrefix = "s/_/"

type fastindexVerifyCfg struct {
	dataDir   string
	dbBackend string
}

// newFastindexVerifyCmd creates the fastindex verify command.
func newFastindexVerifyCmd(io commands.IO) *commands.Command {
	cfg := &fastindexVerifyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "fastindex verify [flags]",
			ShortHelp:  "audits the persisted fast index against the authoritative tree",
			LongHelp: "Walks the bptree main store and checks that every persisted fast-index " +
				"entry matches the authoritative tree value for its key (the gno#6011 " +
				"consistency invariant). READ-ONLY. The node must be stopped (the DB is " +
				"opened exclusively) or the check run against a copied data directory. " +
				"Exit status is non-zero when a stamp-current index disagrees with the " +
				"tree (corruption) or the stamp is ahead of the latest version (rewound DB).",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execFastindexVerify(cfg, io)
		},
	)
}

func (c *fastindexVerifyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.dataDir,
		"data-dir",
		defaultNodeDir,
		"the node's data directory (the DB is read from <data-dir>/"+config.DefaultDBDir+")",
	)
	fs.StringVar(
		&c.dbBackend,
		"db-backend",
		string(dbm.PebbleDBBackend),
		"the DB backend the node was run with",
	)
}

func execFastindexVerify(cfg *fastindexVerifyCfg, io commands.IO) error {
	dbDir := filepath.Join(cfg.dataDir, config.DefaultDBDir)
	if !isValidDirectory(dbDir) {
		return fmt.Errorf("DB directory %q not found (is -data-dir correct?)", dbDir)
	}

	raw, err := dbm.NewDB("gnolang", dbm.BackendType(cfg.dbBackend), dbDir)
	if err != nil {
		return fmt.Errorf("open %q DB at %q: %w (is the node stopped?)", cfg.dbBackend, dbDir, err)
	}
	defer raw.Close()

	// The main store's bptree (fast index included) lives under mainStorePrefix.
	tree := bptree.NewMutableTreeWithDB(
		dbm.NewPrefixDB(raw, []byte(mainStorePrefix)),
		10000, bptree.NewNopLogger(), bptree.FastIndexOption(true),
	)
	// LoadReadonly: never runs fast-index maintenance, so it inspects even a
	// rewound DB that the node itself would refuse to boot.
	if _, err := tree.LoadReadonly(); err != nil {
		return fmt.Errorf("load main store: %w", err)
	}

	rep, err := tree.VerifyFastIndex()
	if err != nil {
		return fmt.Errorf("verify fast index: %w", err)
	}

	io.Printfln("fast-index audit: version=%d stamp=%d (present=%t) entries=%d mismatches=%d",
		rep.Version, rep.Stamp, rep.StampPresent, rep.Entries, rep.MismatchCount)
	for _, m := range rep.Mismatches {
		io.Printfln("  %s", m.String())
	}
	if extra := rep.MismatchCount - len(rep.Mismatches); extra > 0 {
		io.Printfln("  ... and %d more (sample capped at %d)", extra, len(rep.Mismatches))
	}

	switch {
	case !rep.StampPresent:
		io.Println("no fast index present (feature disabled or never built) — nothing to verify")
		return nil
	case rep.Stamp > rep.Version:
		return fmt.Errorf("fast-index stamp (%d) is ahead of the latest version (%d): "+
			"the DB was rewound; the node will refuse to boot until the index is rebuilt "+
			"(resync, or drop the fast-index stamp to force a rebuild)", rep.Stamp, rep.Version)
	case rep.Stamp < rep.Version:
		io.Printfln("WARN: fast index is behind the latest version (%d < %d); "+
			"it will be rebuilt automatically on the next node start", rep.Stamp, rep.Version)
		return nil
	case rep.MismatchCount > 0:
		return fmt.Errorf("fast-index CORRUPTION: %d stamp-current entries disagree with the "+
			"authoritative tree (gno#6011 class) — do not serve; rebuild the index", rep.MismatchCount)
	default:
		io.Println("OK: fast index is current and consistent with the authoritative tree")
		return nil
	}
}
