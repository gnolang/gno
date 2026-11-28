package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeEnablePkgCfg struct {
	RootCfg   *client.MakeTxCfg
	PkgPath   string
	PkgDir    string
	PkgHash   string
	PkgHeight int64
}

func NewMakeEnablePkgCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeEnablePkgCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "enablepkg",
			ShortUsage: "enablepkg [flags] <key-name>",
			ShortHelp:  "activates a package awaiting approval",
			LongHelp: `Activates a package that was submitted under the "inert" code
submission policy and is waiting for an approver.

The approval names the source being approved, not just its path: the submitter
may replace parked bytes at any time, so an approval that named only a path
could be made to activate something the approver never read.

Give the source with -pkgdir, which hashes a local copy of what you reviewed.
Do not take the hash from the chain -- that would approve whatever is parked
right now, which is the thing this guards against. -pkg-hash is for the case
where the hash was computed elsewhere.

The hash covers what the author's directory declares, which is everything you
can read locally. It cannot cover the [addpkg] section the chain writes at
submit, so a re-submission of the same sources keeps the same hash while
changing the storage-deposit ceiling that YOUR transaction pays against. Pass
-pkg-height with the block the submission you reviewed landed in to pin that
too; any re-submission then invalidates the approval.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeEnablePkg(cfg, args, io)
		},
	)
}

func (c *MakeEnablePkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path to activate (required)",
	)

	fs.StringVar(
		&c.PkgDir,
		"pkgdir",
		"",
		"path to a local copy of the reviewed source, hashed to name what is being approved",
	)

	fs.StringVar(
		&c.PkgHash,
		"pkg-hash",
		"",
		"the content hash to approve, if computed elsewhere; alternative to -pkgdir",
	)

	fs.Int64Var(
		&c.PkgHeight,
		"pkg-height",
		0,
		"block height of the submission being approved; pins it so a re-submission invalidates this approval",
	)
}

func execMakeEnablePkg(cfg *MakeEnablePkgCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	switch {
	case cfg.PkgDir == "" && cfg.PkgHash == "":
		return errors.New("specify -pkgdir or -pkg-hash: an approval has to name the source it approves")
	case cfg.PkgDir != "" && cfg.PkgHash != "":
		return errors.New("specify only one of -pkgdir or -pkg-hash")
	}
	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	if len(args) != 1 {
		return flag.ErrHelp
	}

	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	approver := info.GetAddress()

	pkgHash := cfg.PkgHash
	if cfg.PkgDir != "" {
		// MPUserAll, matching addpkg: a parked package is stored exactly as it
		// was submitted, test files included, so anything less would hash a
		// different file set than the chain holds.
		memPkg, err := gno.ReadMemPackage(cfg.PkgDir, cfg.PkgPath, gno.MPUserAll)
		if err != nil {
			return errors.Wrap(err, "reading package")
		}
		if memPkg.IsEmpty() {
			return errors.New("found an empty package at " + cfg.PkgDir)
		}
		// Checked, not assigned: PackageContentHash fails on a directory with
		// no gnomod.toml, a malformed one, or one over gnomod's size limit, and
		// swallowing that would produce a signed approval naming no source at
		// all -- exactly what the -pkgdir/-pkg-hash guard above exists to
		// prevent, except discovered on chain after the fee is paid.
		pkgHash, err = vm.PackageContentHash(memPkg)
		if err != nil {
			return errors.Wrap(err, "hashing the reviewed source")
		}
	}

	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee")
	}

	msg := vm.MsgEnablePackage{
		Approver:  approver,
		PkgPath:   cfg.PkgPath,
		PkgHash:   pkgHash,
		PkgHeight: cfg.PkgHeight,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(cfg.RootCfg.GasWanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		cfg.RootCfg.RootCfg.OnTxSuccess = PrintTxSuccess
		return client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
	}
	io.Println(string(amino.MustMarshalJSON(tx)))
	return nil
}
