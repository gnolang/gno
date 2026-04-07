package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// execRun generates Gno code that calls the expression and executes via maketx run.
// This allows importing the realm and calling with full Go syntax.
func execRun(_ context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if p.Kind != PathCall {
		return fmt.Errorf("RUN expects a function call like gno.land/r/foo/bar.Func(...)")
	}

	// Generate the code
	pkgAlias := path.Base(p.PkgPath)
	code := generateRunCode(p.PkgPath, pkgAlias, p.Symbol, p.Args)

	if cfg.printGnokeyCmd {
		return printRunGnokeyCmd(cfg, p, code, io)
	}

	client, _, err := cfg.signingClient(p.Domain, io)
	if err != nil {
		return err
	}

	info, err := client.Signer.Info()
	if err != nil {
		return err
	}

	msg := vm.MsgRun{
		Caller: info.GetAddress(),
		Package: &std.MemPackage{
			Name: "main",
			Path: "", // ephemeral
			Files: []*std.MemFile{
				{Name: "run.gno", Body: code},
			},
		},
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

	if gasWanted == 0 {
		if !cfg.quiet {
			io.ErrPrintfln("Estimating gas...")
		}
		simCfg := gnoclient.BaseTxCfg{GasFee: gasFee, GasWanted: 100_000_000}
		tx, err := gnoclient.NewRunTx(simCfg, msg)
		if err != nil {
			return fmt.Errorf("building sim tx: %w", err)
		}
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
		tx, err := gnoclient.NewRunTx(txCfg, msg)
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
				"code": code, "data": string(result.Data),
			})
		}
		io.Printfln("Simulation OK — gas used: %d", result.GasUsed)
		io.Println()
		io.Println("Generated code:")
		io.Println(code)
		return nil
	}

	res, err := client.Run(txCfg, msg)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"height": res.Height, "hash": fmt.Sprintf("%X", res.Hash),
			"gas_used": res.DeliverTx.GasUsed, "code": code,
			"data": string(res.DeliverTx.Data),
		})
	}
	io.Printfln("TX committed — height: %d, hash: %X", res.Height, res.Hash)
	io.Printfln("  Gas: %d/%d", res.DeliverTx.GasUsed, res.DeliverTx.GasWanted)
	return nil
}

// generateRunCode generates a main.gno file that imports the realm and calls the function.
func generateRunCode(pkgPath, pkgAlias, funcName string, args []string) string {
	var sb strings.Builder
	sb.WriteString("package main\n\n")
	sb.WriteString(fmt.Sprintf("import %q\n\n", pkgPath))
	sb.WriteString("func main() {\n")
	sb.WriteString(fmt.Sprintf("\t%s.%s(%s)\n", pkgAlias, funcName, joinArgs(args)))
	sb.WriteString("}\n")
	return sb.String()
}

func printRunGnokeyCmd(cfg *baseCfg, p *GnoPath, code string, io commands.IO) error {
	remote, err := cfg.resolveRemote(p.Domain)
	if err != nil {
		return err
	}

	io.Println("# Generated code (save to run.gno):")
	io.Println(code)
	io.Println()

	parts := []string{
		"gnokey", "maketx", "run",
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
	keyName, err := cfg.resolveKeyName()
	if err != nil {
		parts = append(parts, "<key-name>")
	} else {
		parts = append(parts, keyName)
	}
	parts = append(parts, "run.gno")

	io.Println(strings.Join(parts, " \\\n  "))
	return nil
}
