package client

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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
	// Only used with SimulateOnly
	GasFeeMargin uint64
	ChainID      string
	// Master, when set, signs the tx as a session account on behalf of this master key
	// (name or bech32). The chain enforces which msg types a session may sign.
	Master string
	// GasProfile, when set, signs the tx and writes a pprof gas profile of it
	// to this file path instead of broadcasting. Requires a node with the gas
	// profiler enabled (e.g. gnodev); it is off on real nodes.
	GasProfile string
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

// ShouldSign reports whether the tx must be signed — either to broadcast it or
// to profile it. When false, the subcommand only prints the unsigned tx.
func (c *MakeTxCfg) ShouldSign() bool {
	return c.Broadcast || c.GasProfile != ""
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
		NewSessionCmd(cfg, io),
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
		true,
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

	fs.Uint64Var(
		&c.GasFeeMargin,
		"gas-fee-margin",
		5,
		"percent to increase the simulated gas fee (only useful with -simulate only)",
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"chainid to sign for (only useful with --broadcast)",
	)

	fs.StringVar(
		&c.Master,
		"master",
		"",
		"session account master's key name or bech32 address (optional)",
	)

	fs.StringVar(
		&c.GasProfile,
		"gasprofile",
		"",
		"sign the tx and write a pprof gas profile to this file instead of "+
			"broadcasting (requires a node with the gas profiler enabled, e.g. gnodev); "+
			"same flag name as 'gno test -gasprofile'",
	)
}

// GetCaller returns the address that should appear as msg.Caller. When c.Master
// is set (session-signed tx), the caller is master; otherwise it's the signer
// resolved from nameOrBech32.
func (c *MakeTxCfg) GetCaller(nameOrBech32 string) (crypto.Address, error) {
	if c.Master != "" {
		return c.GetMaster()
	}

	kb, err := keys.NewKeyBaseFromDir(c.RootCfg.Home)
	if err != nil {
		return crypto.Address{}, err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return crypto.Address{}, err
	}
	return info.GetAddress(), nil
}

// GetMaster resolves c.Master (bech32 or keybase name) to an Address. Returns
// an error if c.Master is empty.
func (c *MakeTxCfg) GetMaster() (crypto.Address, error) {
	if c.Master == "" {
		return crypto.Address{}, errors.New("master not set")
	}
	if addr, err := crypto.AddressFromBech32(c.Master); err == nil {
		// Master is already bech32; skip the keybase.
		return addr, nil
	}
	kb, err := keys.NewKeyBaseFromDir(c.RootCfg.Home)
	if err != nil {
		return crypto.Address{}, err
	}
	info, err := kb.GetByNameOrAddress(c.Master)
	if err != nil {
		return crypto.Address{}, err
	}
	return info.GetAddress(), nil
}

// signTx resolves the signer's account number and sequence (querying the node),
// signs tx with the named key, and returns the signed tx. It is shared by the
// broadcast and profile paths.
func signTx(cfg *MakeTxCfg, nameOrBech32 string, tx std.Tx, pass string) (std.Tx, error) {
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.Home)
	if err != nil {
		return std.Tx{}, err
	}

	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return std.Tx{}, err
	}
	accountAddr := info.GetAddress()

	// query for the account number and sequence
	var accountNumber uint64
	var sequence uint64
	qopts := &QueryCfg{
		RootCfg: cfg.RootCfg,
	}
	if cfg.Master == "" {
		qopts.Path = fmt.Sprintf("auth/accounts/%s", accountAddr)
		qres, err := QueryHandler(qopts)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "query account")
		}
		var qret struct {
			BaseAccount std.BaseAccount
			Attributes  uint64 `json:"attributes"` // GnoAccount extension
		}
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return std.Tx{}, err
		}

		accountNumber = qret.BaseAccount.AccountNumber
		sequence = qret.BaseAccount.Sequence
	} else {
		masterAddr, err := cfg.GetMaster()
		if err != nil {
			return std.Tx{}, err
		}
		sessionAddr := accountAddr
		qopts.Path = fmt.Sprintf("auth/accounts/%s/session/%s", crypto.AddressToBech32(masterAddr), sessionAddr)
		qres, err := QueryHandler(qopts)
		if err != nil {
			return std.Tx{}, errors.Wrap(err, "query session account")
		}
		var qret struct {
			BaseSessionAccount std.BaseSessionAccount
			AllowPaths         []string `json:"allow_paths,omitempty"` // GnoSessionAccount extension
		}
		if err = amino.UnmarshalJSON(qres.Response.Data, &qret); err != nil {
			return std.Tx{}, err
		}

		accountNumber = qret.BaseSessionAccount.BaseAccount.AccountNumber
		sequence = qret.BaseSessionAccount.BaseAccount.Sequence
	}

	// sign tx
	sOpts := signOpts{
		chainID:         cfg.ChainID,
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
		return std.Tx{}, fmt.Errorf("unable to sign transaction: %w", err)
	}

	if cfg.Master != "" {
		signature.SessionAddr = info.GetAddress()
	}

	// Add the signature to the tx
	if err = addSignature(&tx, signature); err != nil {
		return std.Tx{}, fmt.Errorf("unable to add signature: %w", err)
	}

	return tx, nil
}

func SignAndBroadcastHandler(
	cfg *MakeTxCfg,
	nameOrBech32 string,
	tx std.Tx,
	pass string,
	io commands.IO,
) (*types.ResultBroadcastTxCommit, error) {
	baseopts := cfg.RootCfg

	// Fetch consensus max gas concurrently with signing — both hit the network.
	var maxGasCh chan consensusMaxGasResult
	if cfg.Simulate != SimulateSkip {
		maxGasCh = make(chan consensusMaxGasResult, 1)
		go func() {
			maxGas, err := fetchConsensusMaxGas(baseopts.Remote)
			maxGasCh <- consensusMaxGasResult{maxGas: maxGas, err: err}
		}()
	}

	signedTx, err := signTx(cfg, nameOrBech32, tx, pass)
	if err != nil {
		return nil, err
	}

	maxGas := resolveMaxGas(maxGasCh, io)

	// broadcast signed tx
	bopts := &BroadcastCfg{
		RootCfg: baseopts,
		tx:      &signedTx,

		DryRun:         cfg.Simulate == SimulateOnly,
		testSimulate:   cfg.Simulate == SimulateTest,
		simulateMaxGas: maxGas,
		GasFeeMargin:   cfg.GasFeeMargin,
	}

	return BroadcastHandler(bopts)
}

// SignAndProfileHandler signs tx and queries the node's .app/profiletx endpoint,
// returning a pprof gas profile of the tx and a status log. The endpoint is off
// by default on real nodes and enabled on dev nodes (e.g. gnodev); against a
// node without it, ProfileTx returns a clear "not enabled" error.
func SignAndProfileHandler(
	cfg *MakeTxCfg,
	nameOrBech32 string,
	tx std.Tx,
	pass string,
) (profile []byte, log string, err error) {
	signedTx, err := signTx(cfg, nameOrBech32, tx, pass)
	if err != nil {
		return nil, "", err
	}

	bz, err := amino.Marshal(&signedTx)
	if err != nil {
		return nil, "", errors.Wrap(err, "remarshaling tx binary bytes")
	}

	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, "", errors.New("missing remote url")
	}
	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return nil, "", err
	}

	return ProfileTx(cli, bz)
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

	// -gasprofile: sign and write a pprof gas profile instead of broadcasting.
	// Say so explicitly: -broadcast is on by default, so staying quiet about
	// not sending would leave a dropped tx looking like a successful send.
	if cfg.GasProfile != "" {
		profile, log, err := SignAndProfileHandler(cfg, nameOrBech32, tx, pass)
		if err != nil {
			return errors.Wrap(err, "profile tx")
		}
		if err := os.WriteFile(cfg.GasProfile, profile, 0o644); err != nil {
			return errors.Wrap(err, "writing gas profile")
		}
		// stderr, not stdout: stdout carries the machine-readable tx result
		// (TX HASH, OK!), and `gno test -gasprofile` reports the same thing on
		// stderr.
		io.ErrPrintfln("gas profile written to %s (%s)\n"+
			"transaction was NOT broadcast (-gasprofile profiles instead of sending)\n"+
			"view with: go tool pprof %s",
			cfg.GasProfile, log, cfg.GasProfile)
		return nil
	}

	bres, err := SignAndBroadcastHandler(cfg, nameOrBech32, tx, pass, io)
	if err != nil {
		return errors.Wrap(err, "broadcast tx")
	}
	if bres.CheckTx.IsErr() {
		return errors.Wrapf(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return handleDeliverResult(cfg.RootCfg, tx, bres, io)
	}

	if cfg.RootCfg.OnTxSuccess != nil {
		cfg.RootCfg.OnTxSuccess(io, tx, bres)
	} else {
		DefaultOnTxSuccess(io, tx, bres)
	}

	return nil
}

// handleDeliverResult handles a failed DeliverTx by invoking OnTxFailure or printing defaults.
func handleDeliverResult(cfg *BaseCfg, tx std.Tx, bres *types.ResultBroadcastTxCommit, io commands.IO) error {
	if cfg.OnTxFailure != nil {
		cfg.OnTxFailure(io, tx, bres)
	} else {
		DefaultOnTxFailure(io, tx, bres)
	}
	return errors.Wrapf(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
}

type consensusMaxGasResult struct {
	maxGas int64
	err    error
}

func resolveMaxGas(ch chan consensusMaxGasResult, io commands.IO) int64 {
	if ch == nil {
		return 0
	}
	res := <-ch
	if res.err != nil {
		io.ErrPrintfln("warning: could not fetch consensus max gas, simulation will use the provided gas-wanted: %v", res.err)
		return 0
	}
	return res.maxGas
}

func fetchConsensusMaxGas(remote string) (int64, error) {
	if remote == "" {
		return 0, errors.New("missing remote url")
	}

	cli, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return 0, err
	}

	res, err := cli.ConsensusParams(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	if res == nil || res.ConsensusParams.Block == nil {
		return 0, nil
	}

	return res.ConsensusParams.Block.MaxGas, nil
}
