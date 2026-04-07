package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// execGet is the default verb — smart dispatch:
//   - PathCall → EVAL (evaluate function)
//   - PathSymbol → READ (read variable or source)
//   - PathNetwork/PathNamespace/PathPackage → INSPECT
func execGet(ctx context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	cfg.debugf(io, "path parsed: kind=%d domain=%s pkgpath=%s symbol=%s args=%v", p.Kind, p.Domain, p.PkgPath, p.Symbol, p.Args)

	switch p.Kind {
	case PathCall:
		cfg.debugf(io, "GET dispatch → EVAL (function call)")
		return execEval(ctx, cfg, expr, io)
	case PathSymbol, PathFile:
		cfg.debugf(io, "GET dispatch → READ (symbol/file)")
		return execRead(ctx, cfg, expr, io)
	default:
		cfg.debugf(io, "GET dispatch → INSPECT (kind=%d)", p.Kind)
		return execInspect(ctx, cfg, expr, io)
	}
}

// execEval evaluates a read-only function call via qeval.
func execEval(_ context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if p.Kind != PathCall && p.Kind != PathSymbol {
		return fmt.Errorf("EVAL expects a function call like gno.land/r/foo/bar.Func(...)")
	}

	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}

	var qevalExpr string
	if p.Kind == PathCall {
		args := joinArgs(p.Args)
		// Auto-inject `cross` for crossing functions
		crossing := isCrossingFunc(c, p.PkgPath, p.Symbol)
		cfg.debugf(io, "crossing check for %s.%s: %v", p.PkgPath, p.Symbol, crossing)
		if crossing {
			if args == "" {
				args = "cross"
			} else {
				args = "cross," + args
			}
		}
		qevalExpr = p.Symbol + "(" + args + ")"
	} else {
		qevalExpr = p.Symbol
	}

	cfg.debugf(io, "qeval: %s.%s", p.PkgPath, qevalExpr)
	result, _, err := c.QEval(p.PkgPath, qevalExpr)
	if err != nil {
		return fmt.Errorf("eval: %w", err)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"pkg_path":   p.PkgPath,
			"expression": qevalExpr,
			"result":     result,
		})
	}
	io.Println(result)
	return nil
}

// execRead reads a variable value or source code.
func execRead(_ context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}

	switch p.Kind {
	case PathFile:
		// Fetch specific file
		cfg.debugf(io, "reading file: %s/%s", p.PkgPath, p.File)
		source, err := queryFile(c, p.PkgPath+"/"+p.File)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		if cfg.jsonOut {
			return outputJSON(io, map[string]any{
				"pkg_path": p.PkgPath, "file": p.File, "source": source,
			})
		}
		io.Println(source)
		return nil

	case PathSymbol:
		if p.IsPublic() {
			// Public symbol: show source code
			cfg.debugf(io, "reading source for public symbol %s.%s", p.PkgPath, p.Symbol)
			return readSource(cfg, p, io)
		}
		// Private symbol: get value via qeval
		cfg.debugf(io, "reading value for private symbol %s.%s via qeval", p.PkgPath, p.Symbol)
		result, _, err := c.QEval(p.PkgPath, p.Symbol)
		if err != nil {
			return fmt.Errorf("reading %s.%s: %w", p.PkgPath, p.Symbol, err)
		}
		if cfg.jsonOut {
			return outputJSON(io, map[string]any{
				"pkg_path": p.PkgPath, "symbol": p.Symbol, "value": result,
			})
		}
		io.Println(result)
		return nil

	case PathPackage:
		// List files
		fileList, err := queryFile(c, p.PkgPath)
		if err != nil {
			return err
		}
		if cfg.jsonOut {
			return outputJSON(io, splitLines(fileList))
		}
		io.Println(fileList)
		return nil

	case PathNamespace:
		// List packages
		result, err := queryPaths(c, p.PkgPath)
		if err != nil {
			return err
		}
		if cfg.jsonOut {
			return outputJSON(io, splitLines(result))
		}
		io.Println(result)
		return nil

	default:
		return fmt.Errorf("READ expects a symbol, package, or namespace path")
	}
}

func readSource(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	cfg.debugf(io, "qfile: listing files in %s", p.PkgPath)
	fileList, err := queryFile(c, p.PkgPath)
	if err != nil {
		return err
	}
	files := splitLines(fileList)
	cfg.debugf(io, "found %d files, searching for symbol %q", len(files), p.Symbol)
	for _, fname := range files {
		if !strings.HasSuffix(fname, ".gno") || strings.HasSuffix(fname, "_test.gno") {
			continue
		}
		cfg.debugf(io, "qfile: reading %s/%s", p.PkgPath, fname)
		source, err := queryFile(c, p.PkgPath+"/"+fname)
		if err != nil {
			continue
		}
		for _, prefix := range []string{"func ", "var ", "type ", "const "} {
			if strings.Contains(source, prefix+p.Symbol) {
				if cfg.jsonOut {
					return outputJSON(io, map[string]any{
						"pkg_path": p.PkgPath, "symbol": p.Symbol,
						"file": fname, "source": source,
					})
				}
				io.Printfln("// %s/%s", p.PkgPath, fname)
				io.Println(source)
				return nil
			}
		}
	}
	return fmt.Errorf("symbol %q not found in %s", p.Symbol, p.PkgPath)
}

// execInspect provides detailed inspection of any gno resource.
func execInspect(_ context.Context, cfg *baseCfg, expr string, io commands.IO) error {
	p, err := ParsePath(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	switch p.Kind {
	case PathNetwork:
		return inspectNetwork(cfg, p, io)
	case PathNamespace:
		return inspectNamespace(cfg, p, io)
	case PathPackage:
		return inspectPackage(cfg, p, io)
	case PathSymbol:
		return inspectSymbol(cfg, p, io)
	case PathCall:
		return execEval(context.Background(), cfg, expr, io)
	default:
		return fmt.Errorf("don't know how to inspect %q", expr)
	}
}

func inspectNetwork(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	remote, err := cfg.resolveRemote(p.Domain)
	if err != nil {
		return err
	}
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	height, _ := c.LatestBlockHeight()
	ver, _, _ := c.QueryAppVersion()

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"domain": p.Domain, "rpc": remote.RPC, "chain_id": remote.ChainID,
			"indexer": remote.Indexer, "block_height": height, "app_version": ver,
		})
	}
	io.Printfln("Network: %s", p.Domain)
	io.Printfln("  RPC:          %s", remote.RPC)
	io.Printfln("  Chain ID:     %s", remote.ChainID)
	if remote.Indexer != "" {
		io.Printfln("  Indexer:      %s", remote.Indexer)
	}
	io.Printfln("  Block height: %d", height)
	if ver != "" {
		io.Printfln("  App version:  %s", ver)
	}
	return nil
}

func inspectNamespace(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	result, err := queryPaths(c, p.PkgPath)
	if err != nil {
		return err
	}
	paths := splitLines(result)
	if cfg.jsonOut {
		return outputJSON(io, map[string]any{"path": p.PkgPath, "packages": paths, "count": len(paths)})
	}
	io.Printfln("Namespace: %s (%d packages)", p.PkgPath, len(paths))
	for _, path := range paths {
		io.Printfln("  %s", path)
	}
	return nil
}

func inspectPackage(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	fileList, _ := queryFile(c, p.PkgPath)
	files := splitLines(fileList)
	funcsJSON, _ := queryFuncs(c, p.PkgPath)
	storage, _ := queryStorage(c, p.PkgPath)

	if cfg.jsonOut {
		m := map[string]any{"pkg_path": p.PkgPath, "files": files}
		if funcsJSON != "" {
			m["functions_raw"] = funcsJSON
		}
		if storage != "" {
			m["storage"] = storage
		}
		return outputJSON(io, m)
	}

	io.Printfln("Realm: %s", p.PkgPath)
	if storage != "" {
		io.Printfln("Storage: %s", storage)
	}
	io.Println()
	if len(files) > 0 {
		io.Println("Files:")
		for _, f := range files {
			io.Printfln("  %s", f)
		}
	}
	if funcsJSON != "" && funcsJSON != "null" && funcsJSON != "[]" {
		io.Println()
		io.Println("Functions:")
		formatFuncs(io, funcsJSON)
	}
	return nil
}

func inspectSymbol(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	result, _, err := c.QEval(p.PkgPath, p.Symbol)
	if err != nil {
		return fmt.Errorf("inspecting %s.%s: %w", p.PkgPath, p.Symbol, err)
	}
	if cfg.jsonOut {
		return outputJSON(io, map[string]any{"pkg_path": p.PkgPath, "symbol": p.Symbol, "value": result})
	}
	io.Printfln("%s.%s = %s", p.PkgPath, p.Symbol, result)
	return nil
}

// isCrossingFunc checks if a function's first parameter is a realm type,
// meaning it's a crossing function that needs `cross` as first arg in qeval.
func isCrossingFunc(client *gnoclient.Client, pkgPath, funcName string) bool {
	funcsJSON, err := queryFuncs(client, pkgPath)
	if err != nil || funcsJSON == "" {
		return false
	}

	type nt struct {
		Name string `json:"Name"`
		Type string `json:"Type"`
	}
	type fs struct {
		FuncName string `json:"FuncName"`
		Params   []nt   `json:"Params"`
	}

	var sigs []fs
	if err := json.Unmarshal([]byte(funcsJSON), &sigs); err != nil {
		return false
	}

	for _, sig := range sigs {
		if sig.FuncName == funcName && len(sig.Params) > 0 {
			// Crossing functions have realm as first param
			return strings.Contains(sig.Params[0].Type, "realm")
		}
	}
	return false
}
