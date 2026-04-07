package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

// Verbs
const (
	VerbGET     = "GET"     // smart dispatch: EVAL|READ|INSPECT
	VerbEVAL    = "EVAL"    // evaluate function call (read-only)
	VerbREAD    = "READ"    // read variable value or source
	VerbINSPECT = "INSPECT" // inspect domain/realm/namespace
	VerbCALL    = "CALL"    // sign + broadcast transaction
	VerbRUN     = "RUN"     // maketx run
)

type baseCfg struct {
	home    string
	keyName string
	jsonOut bool
	quiet   bool

	// tx flags (used by CALL/RUN)
	send           string
	gasWanted      int64
	gasFee         string
	dryRun         bool
	printGnokeyCmd bool
	debug          bool
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.home, "home", defaultHome(), "gno config home directory")
	fs.StringVar(&c.keyName, "key", "", "key name or address from keybase")
	fs.BoolVar(&c.jsonOut, "json", false, "output as JSON")
	fs.BoolVar(&c.quiet, "quiet", false, "suppress non-essential output")
	fs.BoolVar(&c.debug, "debug", false, "show debug info (cache, discovery, queries)")
	fs.StringVar(&c.send, "send", "", "coins to send with CALL/RUN (e.g., 1000000ugnot)")
	fs.Int64Var(&c.gasWanted, "gas-wanted", 0, "gas limit (0 = auto-estimate)")
	fs.StringVar(&c.gasFee, "gas-fee", "1000000ugnot", "gas fee")
	fs.BoolVar(&c.dryRun, "dry-run", false, "simulate without broadcasting")
	fs.BoolVar(&c.printGnokeyCmd, "print-gnokey-command", false, "print equivalent gnokey command instead of executing")
}

// debugf prints debug info to stderr if --debug is enabled.
func (c *baseCfg) debugf(io commands.IO, format string, args ...any) {
	if c.debug {
		io.ErrPrintfln("[debug] "+format, args...)
	}
}

func defaultHome() string {
	if h := os.Getenv("GNOHOME"); h != "" {
		return h
	}
	if h := os.Getenv("GNO_HOME"); h != "" {
		return h
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "gno")
	}
	return filepath.Join(dir, "gno")
}

func (c *baseCfg) resolveRemote(domain string) (*Remote, error) {
	if domain == "" {
		return nil, fmt.Errorf("no domain specified")
	}
	return DiscoverRemote(c.home, domain, c.dbgFunc())
}

// dbgFunc returns a DebugFunc that prints to stderr if debug is enabled.
func (c *baseCfg) dbgFunc() DebugFunc {
	if !c.debug {
		return nil
	}
	return func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}

func rpcClientFromRemote(remote *Remote) (rpcclient.Client, error) {
	return rpcclient.NewHTTPClient(remote.RPC)
}

func (c *baseCfg) keybase() (keys.Keybase, error) {
	return keys.NewKeyBaseFromDir(c.home)
}

func (c *baseCfg) queryClient(domain string) (*gnoclient.Client, *Remote, error) {
	remote, err := c.resolveRemote(domain)
	if err != nil {
		return nil, nil, err
	}
	rpc, err := rpcClientFromRemote(remote)
	if err != nil {
		return nil, nil, err
	}
	return &gnoclient.Client{RPCClient: rpc}, remote, nil
}

// resolveKeyName returns the effective key name from --key flag or config.
func (c *baseCfg) resolveKeyName() (string, error) {
	if c.keyName != "" {
		return c.keyName, nil
	}
	cfg, err := LoadConfig(c.home)
	if err != nil {
		return "", err
	}
	if cfg.Key != "" {
		return cfg.Key, nil
	}
	return "", fmt.Errorf("no key specified (use --key or 'gnopie config set key=<name>')")
}

func (c *baseCfg) signingClient(domain string, io commands.IO) (*gnoclient.Client, *Remote, error) {
	keyName, err := c.resolveKeyName()
	if err != nil {
		return nil, nil, err
	}
	remote, err := c.resolveRemote(domain)
	if err != nil {
		return nil, nil, err
	}
	rpc, err := rpcClientFromRemote(remote)
	if err != nil {
		return nil, nil, err
	}
	kb, err := c.keybase()
	if err != nil {
		return nil, nil, fmt.Errorf("opening keybase: %w", err)
	}
	pass, err := io.GetPassword(fmt.Sprintf("Enter password (%s):", keyName), false)
	if err != nil {
		return nil, nil, fmt.Errorf("reading password: %w", err)
	}
	return &gnoclient.Client{
		Signer: &gnoclient.SignerFromKeybase{
			Keybase: kb, Account: keyName, Password: pass, ChainID: remote.ChainID,
		},
		RPCClient: rpc,
	}, remote, nil
}

func main() {
	io := commands.NewDefaultIO()
	cfg := &baseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnopie",
			ShortUsage: "gnopie [flags] [VERB] <expression>",
			ShortHelp:  "gnopie — like httpie, but for gno.land",
			LongHelp: `gnopie is an opinionated CLI for gno.land chains, inspired by httpie.

Network configuration is auto-discovered from the domain via gnoconnect
meta tags (e.g., gnopie fetches https://gno.land/ to find RPC and chain ID).
Results are cached locally for 24h.

Usage:
  gnopie gno.land/r/foo/bar.Baz("hello")           GET (auto: eval function)
  gnopie gno.land/r/foo/bar.counter                 GET (auto: read variable)
  gnopie gno.land/r/foo/bar                         GET (auto: inspect realm)
  gnopie gno.land                                   GET (auto: inspect network)
  gnopie EVAL gno.land/r/foo/bar.Baz("hello")       EVAL explicitly
  gnopie READ gno.land/r/foo/bar.counter             READ explicitly
  gnopie INSPECT gno.land/r/foo/bar                  INSPECT explicitly
  gnopie CALL gno.land/r/foo/bar.Baz("hello")       CALL (transaction)
  gnopie RUN gno.land/r/foo/bar.Baz("hello")        RUN (maketx run)

Verbs:
  GET      (default) Smart dispatch: EVAL for calls, READ for symbols, INSPECT for the rest
  EVAL     Evaluate a read-only function call via qeval
  READ     Read variable value (qeval) or source code (qfile)
  INSPECT  Inspect network, namespace, realm, or symbol
  CALL     Sign and broadcast a transaction (requires --key)
  RUN      Generate and execute Gno code via maketx run (requires --key)`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return dispatch(ctx, cfg, args, io)
		},
	)

	cmd.AddSubCommands(
		newConfigCmd(cfg, io),
		newCompletionCmd(io),
		newVersionCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}

func dispatch(ctx context.Context, cfg *baseCfg, args []string, io commands.IO) error {
	cfg.debugf(io, "args: %v", args)
	if len(args) == 0 {
		return fmt.Errorf("usage: gnopie [VERB] <expression>\nRun 'gnopie --help' for details")
	}

	verb := VerbGET
	exprArgs := args

	switch first := strings.ToUpper(args[0]); first {
	case VerbGET, VerbEVAL, VerbREAD, VerbINSPECT, VerbCALL, VerbRUN:
		verb = first
		exprArgs = args[1:]
	}

	if len(exprArgs) == 0 {
		return fmt.Errorf("missing expression")
	}

	expr := exprArgs[0]
	cfg.debugf(io, "verb=%s expr=%s", verb, expr)

	switch verb {
	case VerbGET:
		return execGet(ctx, cfg, expr, io)
	case VerbEVAL:
		return execEval(ctx, cfg, expr, io)
	case VerbREAD:
		return execRead(ctx, cfg, expr, io)
	case VerbINSPECT:
		return execInspect(ctx, cfg, expr, io)
	case VerbCALL:
		return execCall(ctx, cfg, expr, io)
	case VerbRUN:
		return execRun(ctx, cfg, expr, io)
	default:
		return fmt.Errorf("unknown verb %q", verb)
	}
}

// --- Query helpers ---

func queryFile(client *gnoclient.Client, pkgPath string) (string, error) {
	res, err := client.Query(gnoclient.QueryCfg{Path: "vm/qfile", Data: []byte(pkgPath)})
	if err != nil {
		return "", err
	}
	return string(res.Response.Data), nil
}

func queryFuncs(client *gnoclient.Client, pkgPath string) (string, error) {
	res, err := client.Query(gnoclient.QueryCfg{Path: "vm/qfuncs", Data: []byte(pkgPath)})
	if err != nil {
		return "", err
	}
	return string(res.Response.Data), nil
}

func queryPaths(client *gnoclient.Client, prefix string) (string, error) {
	res, err := client.Query(gnoclient.QueryCfg{Path: "vm/qpaths", Data: []byte(prefix)})
	if err != nil {
		return "", err
	}
	return string(res.Response.Data), nil
}

func queryStorage(client *gnoclient.Client, pkgPath string) (string, error) {
	res, err := client.Query(gnoclient.QueryCfg{Path: "vm/qstorage", Data: []byte(pkgPath)})
	if err != nil {
		return "", err
	}
	return string(res.Response.Data), nil
}

func splitLines(s string) []string {
	var result []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			result = append(result, l)
		}
	}
	return result
}

func outputJSON(io commands.IO, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	io.Println(string(data))
	return nil
}

func joinArgs(args []string) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		if isNumeric(arg) || arg == "true" || arg == "false" {
			parts[i] = arg
		} else {
			parts[i] = `"` + arg + `"`
		}
	}
	return strings.Join(parts, ",")
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if c == '-' && i == 0 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func formatFuncs(io commands.IO, jsonStr string) {
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
	if err := json.Unmarshal([]byte(jsonStr), &sigs); err != nil {
		io.Println(jsonStr)
		return
	}
	for _, sig := range sigs {
		var params, results []string
		for _, p := range sig.Params {
			name := cleanParamName(p.Name)
			typ := cleanType(p.Type)
			if name != "" {
				params = append(params, name+" "+typ)
			} else {
				params = append(params, typ)
			}
		}
		for _, r := range sig.Results {
			typ := cleanType(r.Type)
			// Skip synthetic result names like .res.0
			name := cleanParamName(r.Name)
			if name != "" {
				results = append(results, name+" "+typ)
			} else {
				results = append(results, typ)
			}
		}
		line := fmt.Sprintf("  func %s(%s)", sig.FuncName, strings.Join(params, ", "))
		switch len(results) {
		case 1:
			line += " " + results[0]
		case 0:
		default:
			line += " (" + strings.Join(results, ", ") + ")"
		}
		io.Println(line)
	}
}

// cleanParamName cleans up internal parameter names.
// Removes synthetic names like ".arg_0", ".res.0", etc.
func cleanParamName(name string) string {
	if name == "" {
		return ""
	}
	// Skip synthetic names
	if strings.HasPrefix(name, ".arg_") || strings.HasPrefix(name, ".res.") {
		return ""
	}
	return name
}

// realmInterfacePattern matches the verbose realm interface type from qfuncs.
const realmInterfacePrefix = "interface {Address func() .uverse.address; Coins func() .uverse.gnocoins;"

// errorInterfaceStr matches the error interface pattern.
const errorInterfaceStr = "interface {Error func() string}"

// cleanType simplifies verbose internal type representations.
func cleanType(t string) string {
	// realm interface → realm
	if strings.Contains(t, realmInterfacePrefix) {
		return "realm"
	}

	// error interface → error
	if t == errorInterfaceStr {
		return "error"
	}

	// .uverse.error → error
	t = strings.ReplaceAll(t, ".uverse.error", "error")
	t = strings.ReplaceAll(t, ".uverse.realm", "realm")
	t = strings.ReplaceAll(t, ".uverse.address", "address")
	t = strings.ReplaceAll(t, ".uverse.gnocoins", "gnocoins")

	// Resolve struct literals to short type names where possible.
	// e.g., struct{title string; description string; executor gno.land/r/gov/dao.Executor; ...}
	// Try to find a type name from the fields — if it has a field with a fully qualified type
	// from the same package, use that package's type.
	if strings.HasPrefix(t, "struct{") {
		// Try to extract a meaningful name from qualified field types
		if short := extractStructTypeName(t); short != "" {
			return short
		}
	}

	// Pointer to qualified type: *gno.land/r/gov/dao.Proposal → *dao.Proposal
	t = shortenQualifiedTypes(t)

	return t
}

// extractStructTypeName tries to find a type name for anonymous struct types.
// If the struct has fields with qualified types from a package, we try to match
// it to a known type name pattern.
func extractStructTypeName(t string) string {
	// Look for qualified types in the struct fields
	// e.g., "gno.land/r/gov/dao.Executor" → package is "dao"
	idx := strings.Index(t, "gno.land/")
	if idx < 0 {
		return ""
	}
	// Find the type reference
	rest := t[idx:]
	dotIdx := strings.Index(rest, ".")
	if dotIdx < 0 {
		return ""
	}
	// Get package path up to the dot
	pkgPath := rest[:dotIdx]
	// Get short package name
	lastSlash := strings.LastIndex(pkgPath, "/")
	if lastSlash < 0 {
		return ""
	}
	// We can't determine the exact type name from the struct literal,
	// but we can shorten the qualified types within it
	return ""
}

// shortenQualifiedTypes replaces fully qualified type paths with short names.
// e.g., "gno.land/r/gov/dao.Executor" → "dao.Executor"
// e.g., "*gno.land/r/gov/dao.Proposal" → "*dao.Proposal"
func shortenQualifiedTypes(t string) string {
	// Process all occurrences of gno.land/... qualified types
	for {
		idx := strings.Index(t, "gno.land/")
		if idx < 0 {
			break
		}
		// Check for pointer prefix
		prefix := t[:idx]

		// Find the end of the qualified name (next space, comma, }, ), or end of string)
		rest := t[idx:]
		end := len(rest)
		for i, ch := range rest {
			if ch == ' ' || ch == ',' || ch == '}' || ch == ')' || ch == ';' {
				end = i
				break
			}
		}
		qualifiedName := rest[:end]
		remainder := rest[end:]

		// Extract short name: "gno.land/r/gov/dao.Proposal" → "dao.Proposal"
		lastSlash := strings.LastIndex(qualifiedName, "/")
		shortName := qualifiedName
		if lastSlash >= 0 {
			shortName = qualifiedName[lastSlash+1:]
		}

		t = prefix + shortName + remainder
	}

	// Also clean up chain/runtime.Realm → runtime.Realm
	t = strings.ReplaceAll(t, "chain/runtime.", "runtime.")

	return t
}
