package client

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"context"
	"flag"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type addPkgCfg struct {
	rootCfg *makeTxCfg

	pkgPath string
	pkgDir  string
	deposit string
}

func newAddPkgCmd(rootCfg *makeTxCfg) *commands.Command {
	cfg := &addPkgCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "addpkg",
			ShortUsage: "addpkg [flags] <key-name>",
			ShortHelp:  "Uploads a new package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddPkg(cfg, args, commands.NewDefaultIO())
		},
	)
}

func (c *addPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.pkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.StringVar(
		&c.pkgDir,
		"pkgdir",
		"",
		"path to package files (required)",
	)

	fs.StringVar(
		&c.deposit,
		"deposit",
		"",
		"deposit coins",
	)
}

func execAddPkg(cfg *addPkgCfg, args []string, io *commands.IO) error {
	if cfg.pkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.pkgDir == "" {
		return errors.New("pkgdir not specified")
	}

	if len(args) != 1 {
		return flag.ErrHelp
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.rootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	creator := info.GetAddress()
	// info.GetPubKey()

	// parse deposit.
	deposit, err := std.ParseCoins(cfg.deposit)
	if err != nil {
		panic(err)
	}

	// open files in directory as MemPackage.
	memPkg := gno.ReadMemPackage(cfg.pkgDir, cfg.pkgPath)
	if memPkg.IsEmpty() {
		panic(fmt.Sprintf("found an empty package %q", cfg.pkgPath))
	}

	// precompile and validate syntax
	err = gno.PrecompileAndCheckMempkg(memPkg)
	if err != nil {
		panic(err)
	}

	// parse gas wanted & fee.
	gaswanted := cfg.rootCfg.gasWanted
	gasfee, err := std.ParseCoin(cfg.rootCfg.gasFee)
	if err != nil {
		panic(err)
	}
	// construct msg & tx and marshal.
	msg := vm.MsgAddPackage{
		Creator: creator,
		Package: memPkg,
		Deposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.rootCfg.memo,
	}

	if cfg.rootCfg.broadcast {
		err := signAndBroadcast(cfg.rootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

func signAndBroadcast(
	cfg *makeTxCfg,
	args []string,
	tx std.Tx,
	io *commands.IO,
) error {
	baseopts := cfg.rootCfg
	txopts := cfg

	// query account
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(baseopts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	accountAddr := info.GetAddress()

	qopts := &queryCfg{
		rootCfg: baseopts,
		path:    fmt.Sprintf("auth/accounts/%s", accountAddr),
	}
	qres, err := queryHandler(qopts)
	if err != nil {
		return errors.Wrap(err, "query account")
	}
	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return err
	}

	// sign tx
	accountNumber := qret.BaseAccount.AccountNumber
	sequence := qret.BaseAccount.Sequence
	sopts := &signCfg{
		rootCfg:       baseopts,
		sequence:      sequence,
		accountNumber: accountNumber,
		chainID:       txopts.chainID,
		nameOrBech32:  nameOrBech32,
		txJSON:        amino.MustMarshalJSON(tx),
	}
	if baseopts.Quiet {
		sopts.pass, err = io.GetPassword("", baseopts.InsecurePasswordStdin)
	} else {
		sopts.pass, err = io.GetPassword("Enter password.", baseopts.InsecurePasswordStdin)
	}
	if err != nil {
		return err
	}

	signedTx, err := SignHandler(sopts)
	if err != nil {
		return errors.Wrap(err, "sign tx")
	}

	// broadcast signed tx
	bopts := &broadcastCfg{
		rootCfg: baseopts,
		tx:      signedTx,
	}
	bres, err := broadcastHandler(bopts)
	if err != nil {
		return errors.Wrap(err, "broadcast tx")
	}
	if bres.CheckTx.IsErr() {
		return errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}
	io.Println(string(bres.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
	io.Println("GAS USED:  ", bres.DeliverTx.GasUsed)

	return nil
}
