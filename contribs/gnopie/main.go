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
	network string
	keyName string
	jsonOut bool
	quiet   bool

	// tx flags (used by CALL/RUN)
	send           string
	gasWanted      int64
	gasFee         string
	dryRun         bool
	generateGnokey bool
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.home, "home", defaultHome(), "gno config home directory")
	fs.StringVar(&c.network, "network", "", "network name from remotes config")
	fs.StringVar(&c.keyName, "key", "", "key name or address from keybase")
	fs.BoolVar(&c.jsonOut, "json", false, "output as JSON")
	fs.BoolVar(&c.quiet, "quiet", false, "suppress non-essential output")
	fs.StringVar(&c.send, "send", "", "coins to send with CALL/RUN (e.g., 1000000ugnot)")
	fs.Int64Var(&c.gasWanted, "gas-wanted", 0, "gas limit (0 = auto-estimate)")
	fs.StringVar(&c.gasFee, "gas-fee", "1000000ugnot", "gas fee")
	fs.BoolVar(&c.dryRun, "dry-run", false, "simulate without broadcasting")
	fs.BoolVar(&c.generateGnokey, "generate-gnokey", false, "print equivalent gnokey command")
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
	cfg, err := LoadRemotes(c.home)
	if err != nil {
		return nil, err
	}
	return cfg.Resolve(c.network, domain)
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

func (c *baseCfg) signingClient(domain string, io commands.IO) (*gnoclient.Client, *Remote, error) {
	if c.keyName == "" {
		return nil, nil, fmt.Errorf("--key is required for signing transactions")
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
	pass, err := io.GetPassword("Enter password:", true)
	if err != nil {
		return nil, nil, fmt.Errorf("reading password: %w", err)
	}
	return &gnoclient.Client{
		Signer: &gnoclient.SignerFromKeybase{
			Keybase: kb, Account: c.keyName, Password: pass, ChainID: remote.ChainID,
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
  RUN      Generate and execute Gno code via maketx run (requires --key)

Management commands:
  gnopie remotes     Manage network configurations
  gnopie completion  Generate shell completions
  gnopie version     Print version`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return dispatch(ctx, cfg, args, io)
		},
	)

	cmd.AddSubCommands(
		newRemotesCmd(cfg, io),
		newCompletionCmd(io),
		newVersionCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}

func dispatch(ctx context.Context, cfg *baseCfg, args []string, io commands.IO) error {
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
			if p.Name != "" {
				params = append(params, p.Name+" "+p.Type)
			} else {
				params = append(params, p.Type)
			}
		}
		for _, r := range sig.Results {
			if r.Name != "" {
				results = append(results, r.Name+" "+r.Type)
			} else {
				results = append(results, r.Type)
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
