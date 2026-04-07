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

	// Handle gnoweb modifiers
	if p.RenderPath == "$source" {
		cfg.debugf(io, "GET dispatch → READ ($source modifier)")
		if p.File != "" {
			// $source&file=admin.gno → read specific file
			return readFile(cfg, p, io)
		}
		return execRead(ctx, cfg, expr, io)
	}
	if p.RenderPath == "$help" || p.RenderPath == "$funcs" {
		cfg.debugf(io, "GET dispatch → INSPECT ($help/$funcs modifier)")
		if p.Symbol != "" {
			// $help#func-Name → inspect specific function
			return readFuncSignature(cfg, p, io)
		}
		return execInspect(ctx, cfg, expr, io)
	}

	switch p.Kind {
	case PathCall:
		cfg.debugf(io, "GET dispatch → EVAL (function call)")
		return execEval(ctx, cfg, expr, io)
	case PathSymbol:
		cfg.debugf(io, "GET dispatch → READ (symbol)")
		return execRead(ctx, cfg, expr, io)
	case PathFile:
		cfg.debugf(io, "GET dispatch → READ (file)")
		return readFile(cfg, p, io)
	case PathAddress:
		cfg.debugf(io, "GET dispatch → INSPECT (address)")
		return inspectAddress(cfg, p, io)
	case PathUser:
		cfg.debugf(io, "GET dispatch → user profile")
		return inspectUser(cfg, p, io)
	case PathPackage:
		// Default for packages: call Render("") like gnoweb
		cfg.debugf(io, "GET dispatch → Render (package)")
		return getRender(cfg, p, io)
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

	switch p.Kind {
	case PathFile:
		return readFile(cfg, p, io)

	case PathSymbol:
		if p.IsPublic() {
			cfg.debugf(io, "reading source for public symbol %s.%s", p.PkgPath, p.Symbol)
			return readSource(cfg, p, io)
		}
		cfg.debugf(io, "reading value for private symbol %s.%s via qeval", p.PkgPath, p.Symbol)
		c, _, err := cfg.queryClient(p.Domain)
		if err != nil {
			return err
		}
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
		qc, _, err := cfg.queryClient(p.Domain)
		if err != nil {
			return err
		}
		fileList, err := queryFile(qc, p.PkgPath)
		if err != nil {
			return err
		}
		if cfg.jsonOut {
			return outputJSON(io, splitLines(fileList))
		}
		io.Println(fileList)
		return nil

	case PathNamespace:
		qc, _, err := cfg.queryClient(p.Domain)
		if err != nil {
			return err
		}
		result, err := queryPaths(qc, p.PkgPath)
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
		decl := extractDecl(source, p.Symbol)
		if decl != "" {
			cfg.debugf(io, "found %s in %s", p.Symbol, fname)
			if cfg.jsonOut {
				return outputJSON(io, map[string]any{
					"pkg_path": p.PkgPath, "symbol": p.Symbol,
					"file": fname, "source": decl,
				})
			}
			io.Printfln("// %s/%s", p.PkgPath, fname)
			io.Println(decl)
			return nil
		}
	}
	return fmt.Errorf("symbol %q not found in %s", p.Symbol, p.PkgPath)
}

// extractDecl extracts a single declaration (func, type, var, const) from source code.
// It finds the line starting with "func Symbol", "type Symbol", etc. and returns
// the complete block including its body (tracking brace depth for funcs/types).
func extractDecl(source, symbol string) string {
	lines := strings.Split(source, "\n")
	prefixes := []string{"func ", "var ", "type ", "const "}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, prefix := range prefixes {
			// Match "func Symbol(" or "func (r Type) Symbol(" or "var Symbol" etc.
			if !strings.Contains(trimmed, prefix+symbol) {
				continue
			}
			// For "func ", also check it's the function name not a substring
			// e.g., "func ModAddPost(" should match "ModAddPost" but not "AddPost"
			idx := strings.Index(trimmed, prefix+symbol)
			afterSymbol := idx + len(prefix) + len(symbol)
			if afterSymbol < len(trimmed) {
				next := trimmed[afterSymbol]
				if next != '(' && next != ' ' && next != '\t' && next != '{' && next != '\n' {
					continue // substring match, skip
				}
			}

			// Found the declaration start. Now extract the full block.
			// For single-line declarations (var, const without block), return just the line.
			if prefix == "var " || prefix == "const " {
				if !strings.Contains(line, "{") {
					return strings.TrimRight(line, "\n")
				}
			}

			// Track braces to find the end of the block
			var result strings.Builder
			depth := 0
			started := false
			for j := i; j < len(lines); j++ {
				result.WriteString(lines[j])
				result.WriteByte('\n')

				for _, ch := range lines[j] {
					if ch == '{' {
						depth++
						started = true
					} else if ch == '}' {
						depth--
					}
				}

				if started && depth == 0 {
					return strings.TrimRight(result.String(), "\n")
				}
			}
			// If we never found matching braces, return what we have
			return strings.TrimRight(result.String(), "\n")
		}
	}
	return ""
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
	case PathAddress:
		return inspectAddress(cfg, p, io)
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

// getRender calls Render() on a realm — the default GET behavior for packages.
func getRender(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}

	renderPath := p.RenderPath
	if renderPath == "" || strings.HasPrefix(renderPath, "$") {
		renderPath = ""
	}

	cfg.debugf(io, "qrender: %s:%s", p.PkgPath, renderPath)
	result, _, err := c.Render(p.PkgPath, renderPath)
	if err != nil {
		// If Render fails, fall back to inspect
		cfg.debugf(io, "Render failed, falling back to inspect: %v", err)
		return inspectPackage(cfg, p, io)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"pkg_path":    p.PkgPath,
			"render_path": renderPath,
			"result":      result,
		})
	}
	io.Println(result)
	return nil
}

// readFile fetches a specific file from a package.
func readFile(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	filePath := p.PkgPath + "/" + p.File
	cfg.debugf(io, "qfile: %s", filePath)
	source, err := queryFile(c, filePath)
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
}

// readFuncSignature shows a specific function's signature from qfuncs.
func readFuncSignature(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}
	funcsJSON, err := queryFuncs(c, p.PkgPath)
	if err != nil {
		return fmt.Errorf("querying functions: %w", err)
	}

	type nt struct {
		Name string `json:"Name"`
		Type string `json:"Type"`
	}
	type fs struct {
		FuncName string `json:"FuncName"`
		Params   []nt   `json:"Params"`
		Results  []nt   `json:"Results"`
	}
	var sigs []fs
	if err := json.Unmarshal([]byte(funcsJSON), &sigs); err != nil {
		return fmt.Errorf("parsing functions: %w", err)
	}

	for _, sig := range sigs {
		if sig.FuncName != p.Symbol {
			continue
		}
		var params, results []string
		for _, param := range sig.Params {
			if param.Name != "" {
				params = append(params, param.Name+" "+param.Type)
			} else {
				params = append(params, param.Type)
			}
		}
		for _, r := range sig.Results {
			if r.Name != "" {
				results = append(results, r.Name+" "+r.Type)
			} else {
				results = append(results, r.Type)
			}
		}
		line := fmt.Sprintf("func %s(%s)", sig.FuncName, strings.Join(params, ", "))
		if len(results) == 1 {
			line += " " + results[0]
		} else if len(results) > 1 {
			line += " (" + strings.Join(results, ", ") + ")"
		}

		if cfg.jsonOut {
			return outputJSON(io, map[string]any{
				"pkg_path": p.PkgPath, "function": line,
			})
		}
		io.Println(line)
		return nil
	}
	return fmt.Errorf("function %q not found in %s", p.Symbol, p.PkgPath)
}

// inspectAddress queries account info for a bech32 address.
func inspectAddress(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	// Use default domain for address queries
	c, remote, err := cfg.queryClient("gno.land")
	if err != nil {
		return err
	}
	_ = remote

	cfg.debugf(io, "querying account %s", p.Address)

	// Query account via auth/accounts path
	res, err := c.Query(gnoclient.QueryCfg{
		Path: fmt.Sprintf("auth/accounts/%s", p.Address),
		Data: []byte{},
	})
	if err != nil {
		return fmt.Errorf("querying account: %w", err)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"address":  p.Address,
			"response": string(res.Response.Data),
		})
	}

	io.Printfln("Address: %s", p.Address)
	if len(res.Response.Data) > 0 {
		io.Println(string(res.Response.Data))
	}
	return nil
}

// inspectUser handles /u/username URLs by querying r/sys/users.
func inspectUser(cfg *baseCfg, p *GnoPath, io commands.IO) error {
	c, _, err := cfg.queryClient(p.Domain)
	if err != nil {
		return err
	}

	username := p.Symbol // stored in Symbol by parser
	cfg.debugf(io, "looking up user %q", username)

	// Render the user page via r/sys/users
	result, _, err := c.Render("gno.land/r/sys/users", username)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}

	if cfg.jsonOut {
		return outputJSON(io, map[string]any{
			"username": username,
			"result":   result,
		})
	}
	io.Println(result)
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
