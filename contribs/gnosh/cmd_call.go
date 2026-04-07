package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type callCfg struct {
	send           string
	gasWanted      int64
	gasFee         string
	dryRun         bool
	generateGnokey bool
}

func (c *callCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.send, "send", "", "coins to send with the call (e.g., 1000000ugnot)")
	fs.Int64Var(&c.gasWanted, "gas-wanted", 0, "gas limit (0 = auto-estimate)")
	fs.StringVar(&c.gasFee, "gas-fee", "1000000ugnot", "gas fee")
	fs.BoolVar(&c.dryRun, "dry-run", false, "simulate the transaction without broadcasting")
	fs.BoolVar(&c.generateGnokey, "generate-gnokey", false, "print equivalent gnokey command instead of executing")
}

func newCallCmd(base *baseCfg, io commands.IO) *commands.Command {
	cfg := &callCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "call",
			ShortUsage: "gnosh call [flags] <pkg-path> <func> [args...]",
			ShortHelp:  "Execute a realm function.",
			LongHelp: `Execute a realm function as a transaction.

By default, gas is auto-estimated via simulation. Use --gas-wanted to override.
Use --dry-run to simulate without broadcasting.
Use --generate-gnokey to print the equivalent gnokey command.

Examples:
  gnosh call gno.land/r/demo/wugnot Deposit --send 1000000ugnot
  gnosh call --dry-run gno.land/r/demo/boards CreateThread --key mykey 1 "Hello" "World"
  gnosh call --generate-gnokey gno.land/r/demo/wugnot Deposit`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execCall(ctx, base, cfg, args, io)
		},
	)
}

func execCall(ctx context.Context, base *baseCfg, cfg *callCfg, args []string, io commands.IO) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: gnosh call <pkg-path> <func> [args...]")
	}

	pkgPath := args[0]
	funcName := args[1]
	funcArgs := args[2:]

	// Generate gnokey command mode
	if cfg.generateGnokey {
		return printGnokeyCmd(base, cfg, pkgPath, funcName, funcArgs, io)
	}

	// Build the message
	caller, client, err := prepareSigningClient(base, io)
	if err != nil {
		return err
	}

	msg := vm.MsgCall{
		Caller:  caller,
		PkgPath: pkgPath,
		Func:    funcName,
		Args:    funcArgs,
	}

	if cfg.send != "" {
		coins, err := std.ParseCoins(cfg.send)
		if err != nil {
			return fmt.Errorf("parsing --send: %w", err)
		}
		msg.Send = coins
	}

	gasWanted := cfg.gasWanted
	gasFee := cfg.gasFee

	// Auto-estimate gas if not specified
	if gasWanted == 0 {
		if !base.quiet {
			io.ErrPrintfln("Estimating gas...")
		}
		estimated, err := estimateCallGas(client, base, msg, gasFee)
		if err != nil {
			return fmt.Errorf("gas estimation failed: %w", err)
		}
		gasWanted = estimated
		if !base.quiet {
			io.ErrPrintfln("Estimated gas: %d", gasWanted)
		}
	}

	txCfg := gnoclient.BaseTxCfg{
		GasFee:    gasFee,
		GasWanted: gasWanted,
	}

	// Dry-run mode: simulate and report
	if cfg.dryRun {
		tx, err := gnoclient.NewCallTx(txCfg, msg)
		if err != nil {
			return fmt.Errorf("building tx: %w", err)
		}
		result, err := client.Simulate(tx)
		if err != nil {
			return fmt.Errorf("simulation failed: %w", err)
		}

		if base.json {
			return outputJSON(io, map[string]any{
				"gas_used":   result.GasUsed,
				"gas_wanted": gasWanted,
				"data":       string(result.Data),
			})
		}

		io.Printfln("Simulation successful")
		io.Printfln("  Gas used:   %d", result.GasUsed)
		io.Printfln("  Gas wanted: %d", gasWanted)
		if len(result.Data) > 0 {
			io.Printfln("  Data:       %s", string(result.Data))
		}
		return nil
	}

	// Broadcast
	res, err := client.Call(txCfg, msg)
	if err != nil {
		return fmt.Errorf("call failed: %w", err)
	}

	if base.json {
		return outputJSON(io, map[string]any{
			"height":     res.Height,
			"hash":       fmt.Sprintf("%X", res.Hash),
			"gas_used":   res.DeliverTx.GasUsed,
			"gas_wanted": res.DeliverTx.GasWanted,
			"data":       string(res.DeliverTx.Data),
		})
	}

	io.Printfln("Transaction committed")
	io.Printfln("  Height:     %d", res.Height)
	io.Printfln("  Hash:       %X", res.Hash)
	io.Printfln("  Gas used:   %d", res.DeliverTx.GasUsed)
	io.Printfln("  Gas wanted: %d", res.DeliverTx.GasWanted)
	if len(res.DeliverTx.Data) > 0 {
		io.Printfln("  Data:       %s", string(res.DeliverTx.Data))
	}
	return nil
}

// estimateCallGas simulates a call transaction to estimate gas usage.
// Returns the estimated gas with a 50% buffer.
func estimateCallGas(client *gnoclient.Client, base *baseCfg, msg vm.MsgCall, gasFee string) (int64, error) {
	// Build an unsigned tx with a large gas limit for simulation
	txCfg := gnoclient.BaseTxCfg{
		GasFee:    gasFee,
		GasWanted: 100_000_000, // large limit for simulation
	}
	tx, err := gnoclient.NewCallTx(txCfg, msg)
	if err != nil {
		return 0, err
	}

	gasUsed, err := client.EstimateGas(tx)
	if err != nil {
		return 0, err
	}

	// Add 50% buffer
	estimated := gasUsed + gasUsed/2
	if estimated < 100_000 {
		estimated = 100_000
	}
	return estimated, nil
}

// prepareSigningClient creates a signing client and returns the caller address.
func prepareSigningClient(base *baseCfg, io commands.IO) (crypto.Address, *gnoclient.Client, error) {
	client, err := base.signingClient(io)
	if err != nil {
		return crypto.Address{}, nil, err
	}
	info, err := client.Signer.Info()
	if err != nil {
		return crypto.Address{}, nil, fmt.Errorf("getting signer info: %w", err)
	}
	return info.GetAddress(), client, nil
}

func printGnokeyCmd(base *baseCfg, cfg *callCfg, pkgPath, funcName string, funcArgs []string, io commands.IO) error {
	parts := []string{
		"gnokey", "maketx", "call",
		"-broadcast",
		fmt.Sprintf("-chainid=%s", base.chainID),
		fmt.Sprintf("-remote=%s", base.remote),
	}
	if cfg.gasWanted > 0 {
		parts = append(parts, fmt.Sprintf("-gas-wanted=%d", cfg.gasWanted))
	} else {
		parts = append(parts, "-gas-wanted=10000000")
	}
	parts = append(parts, fmt.Sprintf("-gas-fee=%s", cfg.gasFee))
	if cfg.send != "" {
		parts = append(parts, fmt.Sprintf("-send=%s", cfg.send))
	}
	parts = append(parts, fmt.Sprintf("-pkgpath=%s", pkgPath))
	parts = append(parts, fmt.Sprintf("-func=%s", funcName))
	for _, arg := range funcArgs {
		parts = append(parts, fmt.Sprintf("-args=%s", arg))
	}
	if base.keyName != "" {
		parts = append(parts, base.keyName)
	} else {
		parts = append(parts, "<key-name>")
	}

	io.Println(strings.Join(parts, " \\\n  "))
	return nil
}
