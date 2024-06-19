package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/repl"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type replCfg struct {
	rootDir        string
	initialCommand string
	skipUsage      bool

	remoteAddr string // FIXME: embed keyscli struct directly
}

func newReplCmd() *commands.Command {
	cfg := &replCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "repl",
			ShortUsage: "repl [flags]",
			ShortHelp:  "starts a GnoVM REPL",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRepl(cfg, args)
		},
	)
}

func (c *replCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
	)

	fs.StringVar(
		&c.initialCommand,
		"command",
		"",
		"initial command to run",
	)

	fs.BoolVar(
		&c.skipUsage,
		"skip-usage",
		false,
		"do not print usage",
	)

	fs.StringVar(
		&c.remoteAddr,
		"remote",
		"",
		"Remote RPC Endpoint (gnokey). If empty, use a local gno.Machine.",
	)
}

func execRepl(cfg *replCfg, args []string) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	if !cfg.skipUsage {
		fmt.Fprint(os.Stderr, `// Usage:
//   gno> import "gno.land/p/demo/avl"     // import the p/demo/avl package
//   gno> func a() string { return "a" }   // declare a new function named a
//   gno> /src                             // print current generated source
//   gno> /editor                          // enter in multi-line mode, end with ';'
//   gno> /reset                           // remove all previously inserted code
//   gno> println(a())                     // print the result of calling a()
//   gno> /exit                            // alternative to <Ctrl-D>
`)
	}

	return runRepl(cfg)
}

func runRepl(cfg *replCfg) error {
	r := repl.NewRepl()

	if cfg.initialCommand != "" {
		handleInput(cfg, r, cfg.initialCommand)
	}

	fmt.Fprint(os.Stdout, "gno> ")

	inEdit := false
	prev := ""
	liner := bufio.NewScanner(os.Stdin)

	for liner.Scan() {
		line := liner.Text()

		if l := strings.TrimSpace(line); l == ";" {
			line, inEdit = "", false
		} else if l == "/editor" {
			line, inEdit = "", true
			fmt.Fprintln(os.Stdout, "// enter a single ';' to quit and commit")
		}
		if prev != "" {
			line = prev + "\n" + line
			prev = ""
		}
		if inEdit {
			fmt.Fprint(os.Stdout, "...  ")
			prev = line
			continue
		}

		if err := handleInput(cfg, r, line); err != nil {
			var goScanError scanner.ErrorList
			if errors.As(err, &goScanError) {
				// We assune that a Go scanner error indicates an incomplete Go statement.
				// Append next line and retry.
				prev = line
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		if prev == "" {
			fmt.Fprint(os.Stdout, "gno> ")
		} else {
			fmt.Fprint(os.Stdout, "...  ")
		}
	}
	return nil
}

// handleInput executes specific "/" commands, or evaluates input as Gno source code.
func handleInput(cfg *replCfg, r *repl.Repl, input string) error {
	switch strings.TrimSpace(input) {
	case "/reset":
		r.Reset()
	case "/src":
		fmt.Fprintln(os.Stdout, r.Src())
	case "/exit":
		os.Exit(0)
	case "":
		// Avoid to increase the repl execution counter if no input.
	default:
		_, err := r.Process(input)
		_ = err // XXX: we don't care about errors, we just want repl to parse
		// XXX: refactor (tmp code)
		if cfg.remoteAddr != "" { // remote mode (gnokey)
			rpcClient, err := rpcclient.NewHTTPClient(cfg.remoteAddr)
			if err != nil {
				panic(err)
			}
			kb, err := keys.NewKeyBaseFromDir("/Users/moul/Library/Application Support/gno")
			if err != nil {
				panic(err)
			}
			signer := gnoclient.SignerFromKeybase{
				Keybase:  kb,
				Account:  "moul",
				Password: os.Getenv("GNOKEY_PWD"),
				ChainID:  "portal-loop",
			}
			client := gnoclient.Client{
				Signer:    signer,
				RPCClient: rpcClient,
			}
			runCfg := gnoclient.BaseTxCfg{
				GasWanted: 10000000,
				GasFee:    "1ugnot",
			}
			/*
							body := fmt.Sprintf(`package main
				func main() {
				%s
							   }`, r.Src())*/
			body := fmt.Sprintf(`package main
import "gno.land/r/manfred/home"
import "gno.land/r/demo/users"
func main() {
	_ = home.Render
	_ = users.Render
	%s
}
`, input)
			// println("src", body)
			msg := gnoclient.MsgRun{
				Package: &std.MemPackage{
					Name: "repl",
					Path: "gno.land/r/myself/repl",
					Files: []*std.MemFile{
						{
							Name: "stdin.gno",
							Body: body,
						},
					},
				},
			}
			res, err := client.Run(runCfg, msg)
			if err != nil {
				fmt.Println("error", err)
			}
			fmt.Println(string(res.DeliverTx.Data))
		} else { // local mode (gno.Machine)
			out, err := r.Process(input)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, out)
		}
	}
	return nil
}
