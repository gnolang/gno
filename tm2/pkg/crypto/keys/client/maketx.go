package client

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// MakeTxConfig contains the configuration to create (and potentially broadcast)
// a transaction.
type MakeTxCfg struct {
	RootCfg *BaseCfg

	// Tx options.
	GasWanted GasWantedValue
	GasFee    string
	Memo      string

	// Broadcast-related options.
	Broadcast bool
	Simulate  SimulateValue
	ChainID   string
}

// GasWantedValue is the maximum amount of gas requested for the execution of
// the transaction.
type GasWantedValue int64

// GasWantedAuto is the default value of [GasWantedValue]. If a [MakeTxCfg] has
// Broadcast set to true and Simulate set to [SimulateTest], then the
// transaction simulation will determine the "real" value of GasWanted, as well
// as the GasFee if unset.
const GasWantedAuto GasWantedValue = 0

// String returns the string value of i, implementing [flag.Value].
func (i GasWantedValue) String() string {
	if i == GasWantedAuto {
		return "auto"
	}
	return strconv.FormatInt(int64(i), 10)
}

// Set sets i to the corresponding value in s, implementing [flag.Value].
func (i *GasWantedValue) Set(s string) error {
	if s == "auto" {
		*i = GasWantedAuto
		return nil
	}
	pi, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	if pi <= 0 {
		return fmt.Errorf("invalid value for GasWantedValue: %d", pi)
	}
	*i = GasWantedValue(pi)
	return nil
}

// SimulateValue is the custom flag value for the -simulate flag.
type SimulateValue byte

// These are the valid options for MakeTxConfig.Simulate.
const (
	SimulateTest SimulateValue = iota
	SimulateSkip
	SimulateOnly
)

// String returns the string value of sv, implementing [flag.Value].
func (sv SimulateValue) String() string {
	switch sv {
	case SimulateTest:
		return "test"
	case SimulateSkip:
		return "skip"
	case SimulateOnly:
		return "only"
	default:
		panic("invalid simulate value")
	}
}

// Set parses s into a value for sv, implementing [flag.Value].
func (sv *SimulateValue) Set(s string) error {
	switch s {
	case "test":
		*sv = SimulateTest
	case "skip":
		*sv = SimulateSkip
	case "only":
		*sv = SimulateOnly
	default:
		return fmt.Errorf(`invalid simulate flag value: %q `+
			`(must be "test", "skip" or "only")`, s)
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
	fs.Var(
		&c.GasWanted,
		"gas-wanted",
		`gas requested for tx; with auto, exclusively with -broadcast and -simulate test,
	detect required gas from the transaction simulation. transaction is simulated with
	max_gas = `+strconv.FormatInt(AutoGasDefaultWanted, 10)+`, then run with min(gas_used * 1.10, max_gas)`,
	)

	fs.StringVar(
		&c.GasFee,
		"gas-fee",
		"",
		"gas payment fee; automatically set with -gas-wanted auto",
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
		"sign, simulate and broadcast the transaction",
	)

	fs.Var(
		&c.Simulate,
		"simulate",
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

// MakeTransaction performs the transaction using the given message and
// MakeTxCfg.
func MakeTransaction(msg std.Msg, cfg *MakeTxCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	switch {
	case cfg.GasWanted == GasWantedAuto && !cfg.Broadcast:
		return errors.New("without -broadcast, -gas-wanted must be set to an integer value")
	case cfg.GasWanted == GasWantedAuto && cfg.Simulate != SimulateTest:
		return errors.New("-gas-wanted must be set to an integer value if -simulate is not test")
	case cfg.GasWanted != GasWantedAuto && cfg.GasFee != "":
		return errors.New("-gas-fee not specified")
	}

	// construct msg & tx and marshal.
	tx := std.Tx{
		Msgs: []std.Msg{msg},
		Memo: cfg.Memo,
	}

	if cfg.Broadcast {
		err := ExecSignAndBroadcast(cfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		// parse gas wanted & fee.
		gasWanted := cfg.GasWanted
		gasFee, err := std.ParseCoin(cfg.GasFee)
		if err != nil {
			return errors.Wrap(err, "parsing gas fee coin")
		}
		tx.Fee = std.NewFee(int64(gasWanted), gasFee)

		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

func ExecSignAndBroadcast(
	cfg *MakeTxCfg,
	args []string,
	tx std.Tx,
	io commands.IO,
) error {
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
		return errors.Wrapf(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	io.Println(string(bres.DeliverTx.Data))
	io.Println("OK!")
	io.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
	io.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
	io.Println("HEIGHT:    ", bres.Height)
	io.Println("EVENTS:    ", string(bres.DeliverTx.EncodeEvents()))
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(bres.Hash))

	return nil
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

	// broadcast signed tx
	bopts := &BroadcastCfg{
		RootCfg: baseopts,
		tx:      &tx,

		DryRun: cfg.Simulate == SimulateOnly,
	}

	if cfg.Simulate == SimulateTest {
		bopts.DryRun = true
		if cfg.GasWanted == GasWantedAuto {
			tx.Fee.GasWanted = AutoGasDefaultWanted
			if tx.Fee.GasFee.IsZero() {
				tx.Fee.GasFee = AutoGasDefaultFee
			}
		}
		if err := signTx(&tx, kb, sOpts, kOpts); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}
		resp, err := BroadcastHandler(bopts)
		if err != nil {
			return resp, err
		}
		if cfg.GasWanted == GasWantedAuto {
			used := resp.DeliverTx.GasUsed
			tx.Fee.GasWanted = min(used+used/10, AutoGasDefaultWanted)
			if err := signTx(&tx, kb, sOpts, kOpts); err != nil {
				return nil, fmt.Errorf("unable to sign transaction, %w", err)
			}
		}
		bopts.DryRun = false
	} else {
		if err := signTx(&tx, kb, sOpts, kOpts); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}
	}

	return BroadcastHandler(bopts)
}

var (
	AutoGasDefaultWanted int64 = 10_000_000
	AutoGasDefaultFee          = std.Coin{Denom: "ugnot", Amount: 1_000_000}
)
