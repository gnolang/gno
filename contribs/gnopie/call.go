package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// execCall executes a realm function as a signed transaction.
func execCall(_ context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if p.Kind != PathCall && p.Kind != PathSymbol {
		return fmt.Errorf("CALL expects gno.land/r/foo/bar.Func(...)")
	}

	// Generate gnokey mode
	if cfg.generateGnokey {
		return printGnokeyCmd(cfg, p, io)
	}

	client, remote, err := cfg.signingClient(p.Domain, io)
	if err != nil {
		return err
	}
	_ = remote

	info, err := client.Signer.Info()
	if err != nil {
		return fmt.Errorf("getting signer info: %w", err)
	}

	var funcArgs []string
	if p.Kind == PathCall {
		funcArgs = p.Args
	}

	msg := vm.MsgCall{
		Caller:  info.GetAddress(),
		PkgPath: p.PkgPath,
		Func:    p.Symbol,
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

	// Auto-estimate gas
	if gasWanted == 0 {
		if !cfg.quiet {
			io.ErrPrintfln("Estimating gas...")
		}
		simCfg := gnoclient.BaseTxCfg{GasFee: gasFee, GasWanted: 100_000_000}
		tx, err := gnoclient.NewCallTx(simCfg, msg)
		if err != nil {
			return fmt.Errorf("building sim tx: %w", err)
		}
		// Sign before simulation — the node requires a valid signature
		signedTx, err := client.SignTx(*tx, 0, 0)
		if err != nil {
			return fmt.Errorf("signing for simulation: %w", err)
		}
		gasUsed, err := client.EstimateGas(signedTx)
		if err != nil {
			return fmt.Errorf("gas estimation: %w", err)
		}
		gasWanted = gasUsed + gasUsed/2
		if gasWanted < 100_000 {
			gasWanted = 100_000
		}
		if !cfg.quiet {
			io.ErrPrintfln("Estimated gas: %d", gasWanted)
		}
	}

	txCfg := gnoclient.BaseTxCfg{GasFee: gasFee, GasWanted: gasWanted}

	if cfg.dryRun {
		tx, err := gnoclient.NewCallTx(txCfg, msg)
		if err != nil {
			return err
		}
		signedTx, err := client.SignTx(*tx, 0, 0)
		if err != nil {
			return fmt.Errorf("signing for simulation: %w", err)
		}
		result, err := client.Simulate(signedTx)
		if err != nil {
			return fmt.Errorf("simulation: %w", err)
		}
		if cfg.jsonOut {
			return outputJSON(io, map[string]any{
				"gas_used": result.GasUsed, "gas_wanted": gasWanted,
				"data": string(result.Data),
			})
		}
		io.Printfln("Simulation OK — gas used: %d, gas wanted: %d", result.GasUsed, gasWanted)
		if len(result.Data) > 0 {
			io.Printfln("Data: %s", string(result.Data))
		}
		return nil
	}

	res, err := client.Call(txCfg, msg)
	if err != nil {
		return fmt.Errorf("call: %w", err)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"height": res.Height, "hash": fmt.Sprintf("%X", res.Hash),
			"gas_used": res.DeliverTx.GasUsed, "gas_wanted": res.DeliverTx.GasWanted,
			"data": string(res.DeliverTx.Data),
		})
	}
	io.Printfln("TX committed — height: %d, hash: %X", res.Height, res.Hash)
	io.Printfln("  Gas: %d/%d", res.DeliverTx.GasUsed, res.DeliverTx.GasWanted)
	if len(res.DeliverTx.Data) > 0 {
		io.Printfln("  Data: %s", string(res.DeliverTx.Data))
	}
	return nil
}

func printGnokeyCmd(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	remote, err := cfg.resolveRemote(p.Domain)
	if err != nil {
		return err
	}

	parts := []string{
		"gnokey", "maketx", "call",
		"-broadcast",
		fmt.Sprintf("-chainid=%s", remote.ChainID),
		fmt.Sprintf("-remote=%s", remote.RPC),
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
	parts = append(parts, fmt.Sprintf("-pkgpath=%s", p.PkgPath))
	parts = append(parts, fmt.Sprintf("-func=%s", p.Symbol))
	if p.Kind == PathCall {
		for _, arg := range p.Args {
			parts = append(parts, fmt.Sprintf("-args=%s", arg))
		}
	}
	keyName, err := cfg.resolveKeyName()
	if err != nil {
		parts = append(parts, "<key-name>")
	} else {
		parts = append(parts, keyName)
	}
	io.Println(strings.Join(parts, " \\\n  "))
	return nil
}
