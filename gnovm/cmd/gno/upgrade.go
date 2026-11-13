package main

import (
	"context"
	"flag"
	"fmt"
	goio "io"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newUpgradeCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "upgrade",
			ShortUsage: "upgrade <command>",
			ShortHelp:  "versioned redeploys of a realm",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)
	cmd.AddSubCommands(newUpgradeCheckCmd(io))
	return cmd
}

type upgradeCheckCmd struct {
	rootDir string
}

func newUpgradeCheckCmd(io commands.IO) *commands.Command {
	cmd := &upgradeCheckCmd{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "check",
			ShortUsage: "upgrade check [flags] <prior-dir> <next-dir>",
			ShortHelp:  "checks that next could be redeployed over prior, as the chain would",
			LongHelp: `Runs the versioned-redeploy rules the chain applies when next replaces
prior at the same path, without a node: the gnomod.toml version and [upgrade]
fields, the type check of both versions, and the carry rules over their
declarations, globals, types and methods. Prints the chain's refusal and exits
non-zero on the first one. See docs/resources/realm-upgrades.md.`,
		},
		cmd,
		func(_ context.Context, args []string) error {
			return execUpgradeCheck(cmd, args, io)
		},
	)
}

func (c *upgradeCheckCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.rootDir, "root-dir", gnoenv.RootDir(), "clone location of github.com/gnolang/gno (gno tries to guess it)")
}

func execUpgradeCheck(cmd *upgradeCheckCmd, args []string, io commands.IO) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	priorDir, nextDir := args[0], args[1]
	priorMod, err := gnomod.ParseDir(priorDir)
	if err != nil {
		return err
	}
	nextMod, err := gnomod.ParseDir(nextDir)
	if err != nil {
		return err
	}
	if priorMod.Module != nextMod.Module {
		return fmt.Errorf("module paths differ: %s is %q, %s is %q", priorDir, priorMod.Module, nextDir, nextMod.Module)
	}
	pkgPath := nextMod.Module
	if err := nextMod.CheckUpgradeFields(priorMod, gno.IsRealmPath(pkgPath)); err != nil {
		io.ErrPrintln(err)
		return errUpgradeRefused
	}

	bs, gs := test.StoreWithOptions(cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true, WithExamples: true})
	// load type-checks and preprocesses one version, printing what fails.
	load := func(dir string) (pn *gno.PackageNode, ok bool) {
		mpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPUserProd)
		if err == nil {
			err = test.LoadImports(gs, mpkg, true)
		}
		if err == nil {
			_, err = gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
				Getter: gs, Mode: gno.TCLatestStrict, ProdOnly: true,
			})
		}
		if err != nil {
			printError(io.Err(), dir, pkgPath, err)
			return nil, false
		}
		// Each version preprocesses in its own transaction: both declare
		// the same path.
		cw := bs.CacheWrap()
		tm := test.Machine(gs.BeginTransaction(cw, cw, nil, nil), goio.Discard, pkgPath, false, nil)
		defer tm.Release()
		didPanic := catchPanic(dir, pkgPath, io.Err(), func() {
			pn, _ = tm.PreprocessFiles(mpkg.Name, mpkg.Path, tm.ParseMemPackage(mpkg), false, false)
		})
		return pn, !didPanic
	}
	prior, okPrior := load(priorDir)
	next, okNext := load(nextDir)
	if !okPrior || !okNext {
		return errUpgradeRefused
	}
	if err := gno.CheckUpgrade(prior, next, !nextMod.Private); err != nil {
		io.ErrPrintln(filepath.ToSlash(nextDir) + ": " + err.Error())
		return errUpgradeRefused
	}
	return nil
}

var errUpgradeRefused = fmt.Errorf("the chain would refuse this redeploy")
