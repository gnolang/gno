package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/precompile"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// CallCfg contains configuration options for executing a contract call.
type CallCfg struct {
	PkgPath        string   // Package path
	FuncName       string   // Function name
	Args           []string // Function arguments
	GasFee         string   // Gas fee
	GasWanted      int64    // Gas wanted
	Send           string   // Send amount
	AccountNumber  uint64   // Account number
	SequenceNumber uint64   // Sequence number
	Memo           string   // Memo
}

// RunCfg contains configuration options for running a temporary package on the blockchain.
type RunCfg struct {
	Package        *std.MemPackage
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// Call executes a contract call on the blockchain.
func (c *Client) Call(cfg CallCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, errors.Wrap(err, "validate signer")
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, errors.Wrap(err, "validate RPC client")
	}

	pkgPath := cfg.PkgPath
	funcName := cfg.FuncName
	args := cfg.Args
	gasWanted := cfg.GasWanted
	gasFee := cfg.GasFee
	send := cfg.Send
	sequenceNumber := cfg.SequenceNumber
	accountNumber := cfg.AccountNumber
	memo := cfg.Memo

	// Validate config.
	if pkgPath == "" {
		return nil, errors.New("missing PkgPath")
	}
	if funcName == "" {
		return nil, errors.New("missing FuncName")
	}

	// Parse send amount.
	sendCoins, err := std.ParseCoins(send)
	if err != nil {
		return nil, errors.Wrap(err, "parsing send coins")
	}

	// Parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	caller := c.Signer.Info().GetAddress()

	// Construct message & transaction and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    sendCoins,
		PkgPath: pkgPath,
		Func:    funcName,
		Args:    args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       memo,
	}

	return c.signAndBroadcastTxCommit(tx, accountNumber, sequenceNumber)
}

// Temporarily load cfg.Package on the blockchain and run main() which can
// call realm functions and use println() to output to the "console".
// This returns bres where string(bres.DeliverTx.Data) is the "console" output.
func (c *Client) Run(cfg RunCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, errors.Wrap(err, "validate signer")
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, errors.Wrap(err, "validate RPC client")
	}

	memPkg := cfg.Package
	gasWanted := cfg.GasWanted
	gasFee := cfg.GasFee
	sequenceNumber := cfg.SequenceNumber
	accountNumber := cfg.AccountNumber
	memo := cfg.Memo

	// Validate config.
	if memPkg.IsEmpty() {
		return nil, errors.New("found an empty package " + memPkg.Path)
	}

	// Parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	caller := c.Signer.Info().GetAddress()

	// precompile and validate syntax
	err = gno.PrecompileAndCheckPkg(true, memPkg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "precompile and check")
	}
	memPkg.Name = "main"
	memPkg.Path = "gno.land/r/" + caller.String() + "/run"

	// Construct message & transaction and marshal.
	msg := vm.MsgRun{
		Caller:  caller,
		Package: memPkg,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       memo,
	}

	return c.signAndBroadcastTxCommit(tx, accountNumber, sequenceNumber)
}

// signAndBroadcastTxCommit signs a transaction and broadcasts it, returning the result.
func (c Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	caller := c.Signer.Info().GetAddress()

	if sequenceNumber == 0 || accountNumber == 0 {
		account, _, err := c.QueryAccount(caller)
		if err != nil {
			return nil, errors.Wrap(err, "query account")
		}
		accountNumber = account.AccountNumber
		sequenceNumber = account.Sequence
	}

	signCfg := SignCfg{
		UnsignedTX:     tx,
		SequenceNumber: sequenceNumber,
		AccountNumber:  accountNumber,
	}
	signedTx, err := c.Signer.Sign(signCfg)
	if err != nil {
		return nil, errors.Wrap(err, "sign")
	}

	bz, err := amino.Marshal(signedTx)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling tx binary bytes")
	}

	bres, err := c.RPCClient.BroadcastTxCommit(bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	if bres.CheckTx.IsErr() {
		return bres, errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return bres, errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	return bres, nil
}

// TODO: Add more functionality, examples, and unit tests.
