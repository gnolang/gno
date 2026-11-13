package client

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type BroadcastCfg struct {
	RootCfg *BaseCfg

	DryRun bool

	// internal
	tx *std.Tx
	// Set by SignAndBroadcastHandler, similar to DryRun.
	// If true, simulation is attempted but not printed;
	// the result is only returned in case of an error.
	testSimulate bool
	// max gas limit to use for simulation (optional).
	simulateMaxGas int64

	// simTxBytes, when set, is a separately signed tx to simulate with. It is
	// built in maketx.go only for code-bearing txs, whose signature the chain
	// verifies when simulating: rewriting GasWanted invalidates the original
	// one. Every other tx simulates with blanked signatures, see
	// buildSimulationTxBytes.
	simTxBytes   []byte
	GasFeeMargin uint64
}

const simulationMaxGasFallback = int64(math.MaxInt64)

func NewBroadcastCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &BroadcastCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "broadcast",
			ShortUsage: "broadcast [flags] <file-name>",
			ShortHelp:  "broadcasts a signed document",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBroadcast(cfg, args, io)
		},
	)
}

func (c *BroadcastCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.DryRun,
		"dry-run",
		false,
		"perform a dry-run broadcast",
	)
}

func execBroadcast(cfg *BroadcastCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}
	filename := args[0]

	jsonbz, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "reading tx document file "+filename)
	}
	var tx std.Tx
	err = amino.UnmarshalJSON(jsonbz, &tx)
	if err != nil {
		return errors.Wrap(err, "unmarshaling tx json bytes")
	}
	cfg.tx = &tx

	res, err := BroadcastHandler(cfg)
	if err != nil {
		return err
	}

	if res.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", res, res.CheckTx.Log)
	} else if res.DeliverTx.IsErr() {
		return handleDeliverResult(cfg.RootCfg, tx, res, io)
	} else {
		if cfg.RootCfg.OnTxSuccess != nil {
			cfg.RootCfg.OnTxSuccess(io, tx, res)
		} else {
			DefaultOnTxSuccess(io, tx, res)
		}
	}
	return nil
}

func BroadcastHandler(cfg *BroadcastCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	if cfg.tx == nil {
		return nil, errors.New("invalid tx")
	}

	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, errors.New("missing remote url")
	}

	bz, err := amino.Marshal(cfg.tx)
	if err != nil {
		return nil, errors.Wrap(err, "remarshaling tx binary bytes")
	}

	// NewHTTPClient applies a 60s request timeout to every call, so
	// context.Background() (no deadline) is safe to use here.
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	// Both for DryRun and testSimulate, we perform simulation.
	// However, DryRun always returns here, while in case of success
	// testSimulate continues onto broadcasting the transaction.
	if cfg.DryRun || cfg.testSimulate {
		simBz, rewritten, err := buildSimulationTxBytes(cfg.tx, bz, cfg.simulateMaxGas)
		if cfg.simTxBytes != nil {
			simBz, rewritten, err = cfg.simTxBytes, true, nil
		}
		if err != nil {
			return nil, err
		}

		originalGasWanted := cfg.tx.Fee.GasWanted
		res, err := SimulateTx(cli, simBz)
		if err != nil {
			return nil, err
		}

		// Revert to the old GasWanted. If the GasUsed exceeds GasWanted, and
		// there's no other error to report), create synthetic OOG error.
		if rewritten {
			res.DeliverTx.GasWanted = originalGasWanted
			if res.DeliverTx.Error == nil && res.DeliverTx.GasUsed > originalGasWanted {
				log := store.OutOfGasLog(res.DeliverTx.GasUsed, originalGasWanted, cfg.simulateMaxGas, "simulation", false)
				res.DeliverTx.Error = abci.ABCIErrorOrStringError(std.ErrOutOfGas(log))
				res.DeliverTx.Log = log
			}
		}
		hasError := res.CheckTx.IsErr() || res.DeliverTx.IsErr()
		if cfg.DryRun && !hasError {
			err = estimateGasFee(cli, res, cfg.GasFeeMargin)
			return res, err
		}
		if hasError {
			appendSuggestedGasWanted(res)
			return res, err
		}
	}

	bres, err := cli.BroadcastTxCommit(context.Background(), bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	return bres, nil
}

// buildSimulationTxBytes returns tx bytes to use for simulation, overriding
// GasWanted to consensus maxGas. If maxGas is -1 (chain has no gas limit) it
// falls back to MaxInt64. If maxGas is 0 (unknown, e.g. fetch failed) the
// original GasWanted is kept. It also returns whether GasWanted was rewritten.
//
// Unless the chain verifies signatures when simulating tx (see
// txNeedsSimulationSignature), the signatures are blanked: the node simulating
// is whichever RPC the user pointed at, and a real signature would hand it a tx
// it can broadcast, even on a dry run that never does.
func buildSimulationTxBytes(tx *std.Tx, txBytes []byte, maxGas int64) ([]byte, bool, error) {
	simTx, rewritten := txWithGasWanted(tx, maxGas)
	blank := len(tx.Signatures) > 0 && !txNeedsSimulationSignature(*tx)
	if !rewritten && !blank {
		return txBytes, false, nil
	}
	if simTx == nil {
		cp := *tx
		simTx = &cp
	}
	if blank {
		simTx.Signatures = blankSignatures(tx.Signatures)
	}
	simBz, err := amino.Marshal(simTx)
	if err != nil {
		return nil, false, errors.Wrap(err, "remarshaling tx binary bytes for simulation")
	}

	return simBz, rewritten, nil
}

// blankSignatures returns a copy of sigs with every signature byte zeroed. Keys,
// session addresses and lengths are kept, since the ante handler charges gas by
// key type and by tx size: the estimate does not move.
func blankSignatures(sigs []std.Signature) []std.Signature {
	out := make([]std.Signature, len(sigs))
	for i, sig := range sigs {
		out[i] = sig
		out[i].Signature = blankSignature(sig.PubKey, sig.Signature)
	}
	return out
}

// blankSignature returns a zeroed copy of sig. A multisig signature is an
// amino-encoded Multisignature that the gas consumer decodes and walks, so it
// keeps its shape and only its constituents are zeroed, recursively. When pub
// is nil, the chain resolves it from the account, so the signature's own shape
// is used instead.
func blankSignature(pub crypto.PubKey, sig []byte) []byte {
	remainingKeys := multisig.MaxTotalKeys
	return blankSignatureWithBudget(pub, sig, multisig.MaxNestingDepth, &remainingKeys)
}

func blankSignatureWithBudget(pub crypto.PubKey, sig []byte, remainingLevels int, remainingKeys *int) []byte {
	if remainingLevels == 0 || *remainingKeys == 0 {
		return make([]byte, len(sig))
	}
	(*remainingKeys)--

	mpk, isMultisig := pub.(multisig.PubKeyMultisigThreshold)
	if pub != nil && !isMultisig {
		return make([]byte, len(sig))
	}
	var ms multisig.Multisignature
	if amino.Unmarshal(sig, &ms) != nil {
		return make([]byte, len(sig))
	}
	nkeys := ms.BitArray.Size()
	if isMultisig {
		if mpk.ValidateStructure() != nil {
			return make([]byte, len(sig))
		}
		nkeys = len(mpk.PubKeys)
	}
	if (!isMultisig && (nkeys == 0 || nkeys > multisig.MaxTotalKeys)) || ms.ValidateBasic(nkeys) != nil {
		return make([]byte, len(sig))
	}

	next := 0
	for i := range ms.BitArray.Size() {
		if ms.BitArray.GetIndex(i) {
			var subkey crypto.PubKey
			if isMultisig {
				subkey = mpk.PubKeys[i]
			}
			ms.Sigs[next] = blankSignatureWithBudget(subkey, ms.Sigs[next], remainingLevels-1, remainingKeys)
			next++
		}
	}
	return ms.Marshal()
}

// txWithGasWanted returns a copy of tx with GasWanted raised to the
// consensus maxGas, and whether it differs from tx. maxGas of -1 (chain has no
// limit) falls back to MaxInt64; 0 (unknown, e.g. the fetch failed) means no
// change. The copy still carries tx's signature, which no longer matches it:
// changing GasWanted changes the sign bytes, so a caller that needs a valid
// signature must clear and re-sign.
func txWithGasWanted(tx *std.Tx, maxGas int64) (*std.Tx, bool) {
	if maxGas < 0 {
		maxGas = simulationMaxGasFallback
	}
	if maxGas == 0 || tx.Fee.GasWanted >= maxGas {
		return nil, false
	}
	simTx := *tx
	simTx.Fee.GasWanted = maxGas
	return &simTx, true
}

func suggestedGasWanted(gasUsed int64) int64 {
	margin := gasUsed / 20
	if gasUsed%20 != 0 {
		margin++
	}
	return overflow.Addp(gasUsed, margin)
}

func appendSuggestedGasWanted(bres *ctypes.ResultBroadcastTxCommit) {
	suggested := suggestedGasWanted(bres.DeliverTx.GasUsed)
	msg := fmt.Sprintf("suggested gas-wanted (gas used + 5%%): %d", suggested)
	if bres.DeliverTx.Info == "" {
		bres.DeliverTx.Info = msg
	} else {
		bres.DeliverTx.Info = bres.DeliverTx.Info + ", " + msg
	}
}

func estimateGasFee(cli client.ABCIClient, bres *ctypes.ResultBroadcastTxCommit, gasFeeMargin uint64) error {
	gasUsed := bres.DeliverTx.GasUsed
	suggested := suggestedGasWanted(gasUsed)

	gp := std.GasPrice{}
	qres, err := cli.ABCIQuery(context.Background(), "auth/gasprice", []byte{})
	if err != nil {
		return errors.Wrap(err, "query gas price")
	}
	err = amino.UnmarshalJSON(qres.Response.Data, &gp)
	if err != nil {
		return errors.Wrap(err, "unmarshaling query gas price result")
	}

	var s string
	if gp.Gas == 0 {
		s = fmt.Sprintf("estimated gas usage: %d (suggested, with 5%% margin: %d)\n", gasUsed, suggested)
	} else {
		fee := gasUsed/gp.Gas + 1
		fee = overflow.Mulp(fee, gp.Price.Amount)
		// fee buffer to cover the sudden change of gas price
		feeBuffer := overflow.Mulp(fee, int64(gasFeeMargin)) / 100
		fee = overflow.Addp(fee, feeBuffer)
		s = fmt.Sprintf("estimated gas usage: %d (suggested, with 5%% margin: %d), gas fee: %d%s, current gas price: %s\n", gasUsed, suggested, fee, gp.Price.Denom, gp.String())
	}
	if bres.DeliverTx.Info == "" {
		bres.DeliverTx.Info = s
	} else {
		bres.DeliverTx.Info = bres.DeliverTx.Info + ", " + s
	}
	return nil
}

func SimulateTx(cli client.ABCIClient, tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	bres, err := cli.ABCIQuery(context.Background(), ".app/simulate", tx)
	if err != nil {
		return nil, errors.Wrap(err, "simulate tx")
	}

	var result abci.ResponseDeliverTx
	err = amino.Unmarshal(bres.Response.Value, &result)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling simulate result")
	}

	return &ctypes.ResultBroadcastTxCommit{
		DeliverTx: result,
	}, nil
}
