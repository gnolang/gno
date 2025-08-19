package client

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeTxCfg struct {
	RootCfg *BaseCfg

	GasWanted int64
	GasFee    string
	Memo      string

	Broadcast bool
	// Valid options are SimulateTest, SimulateSkip or SimulateOnly.
	Simulate string
	ChainID  string
}

// These are the valid options for MakeTxConfig.Simulate.
const (
	SimulateTest = "test"
	SimulateSkip = "skip"
	SimulateOnly = "only"
)

func (c *MakeTxCfg) Validate() error {
	switch c.Simulate {
	case SimulateTest, SimulateSkip, SimulateOnly:
	default:
		return fmt.Errorf("invalid simulate option: %q", c.Simulate)
	}
	return nil
}

func NewMakeTxCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &MakeTxCfg{
		RootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "maketx",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "composes a tx document to sign",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		NewMakeSendCmd(cfg, io),
	)

	return cmd
}

func (c *MakeTxCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(
		&c.GasWanted,
		"gas-wanted",
		0,
		"gas requested for tx",
	)

	fs.StringVar(
		&c.GasFee,
		"gas-fee",
		"",
		"gas payment fee",
	)

	fs.StringVar(
		&c.Memo,
		"memo",
		"",
		"any descriptive text",
	)

	fs.BoolVar(
		&c.Broadcast,
		"broadcast",
		false,
		"sign, simulate and broadcast",
	)

	fs.StringVar(
		&c.Simulate,
		"simulate",
		"test",
		`select how to simulate the transaction (only useful with --broadcast); valid options are
		- test: attempts simulating the transaction, and if successful performs broadcasting (default)
		- skip: avoids performing transaction simulation
		- only: avoids broadcasting transaction (ie. dry run)`,
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"chainid to sign for (only useful with --broadcast)",
	)
}

func SignAndBroadcastHandler(
	cfg *MakeTxCfg,
	nameOrBech32 string,
	tx std.Tx,
	pass string,
) (*types.ResultBroadcastTxCommit, error) {
	baseopts := cfg.RootCfg
	txopts := cfg

	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return nil, err
	}

	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return nil, err
	}
	accountAddr := info.GetAddress()

	qopts := &QueryCfg{
		RootCfg: baseopts,
		Path:    fmt.Sprintf("auth/accounts/%s", accountAddr),
	}
	qres, err := QueryHandler(qopts)
	if err != nil {
		return nil, errors.Wrap(err, "query account")
	}
	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, err
	}

	// sign tx
	accountNumber := qret.BaseAccount.AccountNumber
	sequence := qret.BaseAccount.Sequence

	sOpts := signOpts{
		chainID:         txopts.ChainID,
		accountSequence: sequence,
		accountNumber:   accountNumber,
	}

	kOpts := keyOpts{
		keyName:     nameOrBech32,
		decryptPass: pass,
	}

	// Generate the transaction signature
	signature, err := generateSignature(&tx, kb, sOpts, kOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to sign transaction: %w", err)
	}

	// Add the signature to the tx
	if err = addSignature(&tx, signature); err != nil {
		return nil, fmt.Errorf("unable to add signature: %w", err)
	}

	// broadcast signed tx
	bopts := &BroadcastCfg{
		RootCfg: baseopts,
		tx:      &tx,

		DryRun:       cfg.Simulate == SimulateOnly,
		testSimulate: cfg.Simulate == SimulateTest,
	}

	return BroadcastHandler(bopts)
}

func ExecSignAndBroadcast(
	cfg *MakeTxCfg,
	args []string,
	tx std.Tx,
	io commands.IO,
) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	baseopts := cfg.RootCfg

	// query account
	nameOrBech32 := args[0]

	var err error
	var pass string
	if baseopts.Quiet {
		pass, err = io.GetPassword("", baseopts.InsecurePasswordStdin)
	} else {
		pass, err = io.GetPassword("Enter password.", baseopts.InsecurePasswordStdin)
	}

	if err != nil {
		return err
	}

	bres, err := SignAndBroadcastHandler(cfg, nameOrBech32, tx, pass)
	if err != nil {
		return errors.Wrap(err, "broadcast tx")
	}
	if bres.CheckTx.IsErr() {
		return errors.Wrapf(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(bres.Hash))
		io.Println("INFO:      ", bres.DeliverTx.Info)
		return errors.Wrapf(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	io.Println(string(bres.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
	io.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
	io.Println("HEIGHT:    ", bres.Height)
	if delta, storageFee, ok := getStorageInfo(bres.DeliverTx.Events); ok {
		io.Printfln("STORAGE DELTA: %d bytes", delta)
		io.Println("STORAGE FEE:  ", storageFee)
		if tx.Fee.GasFee.Denom == storageFee.Denom {
			io.Println("TOTAL TX COST:", tx.Fee.GasFee.Add(storageFee))
		}
	}
	io.Println("EVENTS:    ", string(bres.DeliverTx.EncodeEvents()))
	io.Println("INFO:      ", bres.DeliverTx.Info)
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(bres.Hash))

	return nil
}

// getStorageInfo searches events for StorageDeposit or UnlockDeposit and
// returns the delta and fee. If this is "UnlockDeposit", then delta and fee are negative.
// The third return is true if found, else false.
func getStorageInfo(events []abci.Event) (int64, std.Coin, bool) {
	for _, event := range events {
		eventKV, ok := event.(abci.EventKeyValue)
		if !ok {
			continue
		}

		isUnlock := false
		deltaStr := ""
		feeStr := ""
		switch eventKV.Type {
		case "StorageDeposit":
			deltaStr, ok = eventKV.FindAttribute("Storage")
			if !ok {
				continue
			}
			feeStr, ok = eventKV.FindAttribute("Deposit")
			if !ok {
				continue
			}
		case "UnlockDeposit":
			deltaStr, ok = eventKV.FindAttribute("ReleaseStorage")
			if !ok {
				continue
			}
			isUnlock = true
			feeStr, ok = eventKV.FindAttribute("Deposit")
			if !ok {
				continue
			}
		default:
			continue
		}

		delta, err := strconv.ParseInt(strings.TrimSuffix(deltaStr, " bytes"), 10, 64)
		if err != nil {
			continue
		}
		fee, err := std.ParseCoin(feeStr)
		if err != nil {
			continue
		}
		if isUnlock {
			delta = -delta
			fee.Amount = -fee.Amount
		}

		return delta, fee, true
	}

	return 0, std.Coin{}, false
}
