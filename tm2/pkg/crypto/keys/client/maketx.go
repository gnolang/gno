package client

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
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
	// GasAuto enables automatic gas estimation when set to true
	GasAuto bool
	Memo    string

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

// parseGasWanted parses a gas value from string to int64
func parseGasWanted(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// registerGasFlags sets up gas-wanted and gas-fee flags with auto mode support
func (c *MakeTxCfg) registerGasFlags(fs *flag.FlagSet) {
	// Set defaults to auto mode
	c.GasAuto = true
	c.GasWanted = 0
	c.GasFee = ""

	// gas-wanted flag: supports numeric values or "auto"
	fs.Func("gas-wanted", "gas requested for tx (default: auto)", func(s string) error {
		if s == "auto" {
			c.GasAuto = true
			c.GasWanted = 0
		} else if s != "" {
			gasWanted, err := parseGasWanted(s)
			if err != nil {
				return fmt.Errorf("invalid gas value: %w", err)
			}
			c.GasWanted = gasWanted
			c.GasAuto = false
		}
		return nil
	})

	// gas-fee flag: supports coin values or "auto"
	fs.Func("gas-fee", "gas payment fee (default: auto)", func(s string) error {
		if s == "auto" {
			c.GasFee = "" // Empty triggers auto fee calculation
		} else if s != "" {
			c.GasFee = s
			// Only disable auto if we also have explicit gas-wanted
			if c.GasWanted > 0 {
				c.GasAuto = false
			}
		}
		return nil
	})
}

func (c *MakeTxCfg) RegisterFlags(fs *flag.FlagSet) {
	c.registerGasFlags(fs)

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
	io.Println("EVENTS:    ", string(bres.DeliverTx.EncodeEvents()))
	io.Println("INFO:      ", bres.DeliverTx.Info)
	io.Println("TX HASH:   ", base64.StdEncoding.EncodeToString(bres.Hash))

	return nil
}

// EstimateOrSetFee handles both auto gas estimation and manual fee setting for a transaction
func EstimateOrSetFee(cfg *MakeTxCfg, tx *std.Tx) error {
	if !cfg.GasAuto {
		// Manual gas setting
		gasfee, err := std.ParseCoin(cfg.GasFee)
		if err != nil {
			return errors.Wrap(err, "parsing gas fee coin")
		}
		tx.Fee = std.NewFee(cfg.GasWanted, gasfee)
		return nil
	}

	// Auto gas estimation
	// Get the remote client
	remote := cfg.RootCfg.Remote
	if remote == "" {
		return errors.New("missing remote url for gas estimation")
	}

	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return errors.Wrap(err, "creating HTTP client for gas estimation")
	}

	// Set temporary fee for simulation
	// We need to set a non-zero fee for the simulation to work
	tempGasFee := std.NewCoin("ugnot", 1000000) // 1 GNOT temporary fee
	tempGasWanted := int64(10000000)            // 10M gas temporary limit

	// Store original fee and signatures to restore if needed
	originalFee := tx.Fee
	originalSignatures := tx.Signatures

	// Set temporary values for simulation
	tx.Fee = std.NewFee(tempGasWanted, tempGasFee)

	// Add a dummy signature for simulation to pass signature validation
	// We need at least one signature with the correct length for the ante handler
	if len(tx.Signatures) == 0 {
		dummySig := std.Signature{
			PubKey:    nil,              // Can be nil for simulation
			Signature: make([]byte, 64), // Empty signature, correct length
		}
		tx.Signatures = []std.Signature{dummySig}
	}

	// Serialize the transaction for simulation
	bz, err := amino.Marshal(tx)
	if err != nil {
		tx.Fee = originalFee // Restore original
		tx.Signatures = originalSignatures
		return errors.Wrap(err, "marshaling tx for gas estimation")
	}

	// Simulate the transaction to get gas usage
	res, err := SimulateTx(cli, bz)
	if err != nil {
		tx.Fee = originalFee // Restore original
		tx.Signatures = originalSignatures
		return errors.Wrap(err, "simulating transaction for gas estimation")
	}

	if res.CheckTx.IsErr() {
		tx.Fee = originalFee // Restore original
		tx.Signatures = originalSignatures
		return errors.Wrapf(res.CheckTx.Error, "transaction check failed during gas estimation: %s", res.CheckTx.Log)
	}

	if res.DeliverTx.IsErr() {
		tx.Fee = originalFee // Restore original
		tx.Signatures = originalSignatures
		return errors.Wrapf(res.DeliverTx.Error, "transaction delivery failed during gas estimation: %s", res.DeliverTx.Log)
	}

	// Get the estimated gas used and add a buffer
	gasUsed := res.DeliverTx.GasUsed
	// Add 10% buffer to the estimated gas
	gasWanted := gasUsed + (gasUsed / 10)

	// Query gas price if fee is not set
	var gasFee std.Coin
	if cfg.GasFee == "" {
		gp := std.GasPrice{}
		qres, err := cli.ABCIQuery("auth/gasprice", []byte{})
		if err != nil {
			return errors.Wrap(err, "querying gas price for fee estimation")
		}
		err = amino.UnmarshalJSON(qres.Response.Data, &gp)
		if err != nil {
			return errors.Wrap(err, "unmarshaling gas price result")
		}

		if gp.Gas == 0 {
			// No gas price set, use a default fee
			gasFee = std.NewCoin("ugnot", 1000000) // Default fee
		} else {
			// Calculate fee based on gas price
			feeAmount := gasWanted/gp.Gas + 1
			// Add 5% buffer for gas price fluctuation
			feeBuffer := feeAmount * 5 / 100
			totalFee := feeAmount + feeBuffer
			gasFee = std.NewCoin(gp.Price.Denom, totalFee)
		}
	} else {
		// Parse existing gas fee
		gasFee, err = std.ParseCoin(cfg.GasFee)
		if err != nil {
			return errors.Wrap(err, "parsing existing gas fee")
		}
	}

	// Update the transaction with estimated values
	cfg.GasWanted = gasWanted
	if cfg.GasFee == "" {
		cfg.GasFee = gasFee.String()
	}
	tx.Fee = std.NewFee(gasWanted, gasFee)

	// Clear dummy signatures - they'll be added properly during signing
	tx.Signatures = originalSignatures

	return nil
}
