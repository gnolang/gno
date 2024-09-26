package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/tests"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type runCfg struct {
	verbose   bool
	rootDir   string
	expr      string
	debug     bool
	debugAddr string
}

func newRunCmd(io commands.IO) *commands.Command {
	cfg := &runCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <file> [<file>...]",
			ShortHelp:  "runs the specified gno files",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(cfg, args, io)
		},
	)
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno binary tries to guess it)",
	)

	fs.StringVar(
		&c.expr,
		"expr",
		"main()",
		"value of expression to evaluate. Defaults to executing function main() with no args",
	)

	fs.BoolVar(
		&c.debug,
		"debug",
		false,
		"enable interactive debugger using stdin and stdout",
	)

	fs.StringVar(
		&c.debugAddr,
		"debug-addr",
		"",
		"enable interactive debugger using tcp address in the form [host]:port",
	)
}

func execRun(cfg *runCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	stdin := io.In()
	stdout := io.Out()
	stderr := io.Err()

	// init store and machine
	testStore := tests.TestStore(cfg.rootDir,
		"", stdin, stdout, stderr,
		tests.ImportModeStdlibsPreferred)
	if cfg.verbose {
		testStore.SetLogStoreOps(true)
	}

	if len(args) == 0 {
		args = []string{"."}
	}

	// read files
	files, err := parseFiles(args, stderr)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no files to run")
	}

	m := runMachine(
		testStore,
		stdin,
		stdout,
		string(files[0].PkgName),
		cfg.debug || cfg.debugAddr != "",
	)

	defer m.Release()

	// If the debug address is set, the debugger waits for a remote client to connect to it.
	if cfg.debugAddr != "" {
		if err := m.Debugger.Serve(cfg.debugAddr); err != nil {
			return err
		}
	}

	// run files
	m.RunFiles(files...)
	runExpr(m, cfg.expr)

	return nil
}

func runMachine(store gno.Store, stdin io.Reader, stdout io.Writer, pkgPath string, debug bool) *gno.Machine {
	var (
		send     std.Coins
		maxAlloc int64
	)

	return runMachineCustom(store, pkgPath, stdin, stdout, maxAlloc, send, debug)
}

func runMachineCustom(store gno.Store, pkgPath string, stdin io.Reader, stdout io.Writer, maxAlloc int64, send std.Coins, debug bool) *gno.Machine {
	ctx := runContext(pkgPath, send)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       pkgPath,
		Output:        stdout,
		Store:         store,
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
	})
	return m
}

// runContext() creates the context for gno run. It has been intentially setup to mirror the context that is
// created for gno test, so that behavior remains consistent between running code and testing code.
func runContext(pkgPath string, send std.Coins) *teststd.TestExecContext {
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	caller := gno.DerivePkgAddr("user1.gno")

	pkgCoins := std.MustParseCoins(ugnot.ValueString(200000000)).Add(send) // >= send.
	banker := newTestBanker(pkgAddr.Bech32(), pkgCoins)
	ctx := stdlibs.ExecContext{
		ChainID:       "run",
		Height:        123,
		Timestamp:     1234567890,
		Msg:           nil,
		OrigCaller:    caller.Bech32(),
		OrigPkgAddr:   pkgAddr.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		Banker:        banker,
		EventLogger:   sdk.NewEventLogger(),
	}
	return &teststd.TestExecContext{
		ExecContext: ctx,
		RealmFrames: make(map[*gno.Frame]teststd.RealmOverride),
	}
}

type testBanker struct {
	coinTable map[crypto.Bech32Address]std.Coins
}

func newTestBanker(args ...interface{}) *testBanker {
	coinTable := make(map[crypto.Bech32Address]std.Coins)
	if len(args)%2 != 0 {
		panic("newTestBanker requires even number of arguments; addr followed by coins")
	}
	for i := 0; i < len(args); i += 2 {
		addr := args[i].(crypto.Bech32Address)
		amount := args[i+1].(std.Coins)
		coinTable[addr] = amount
	}
	return &testBanker{
		coinTable: coinTable,
	}
}

func (tb *testBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return tb.coinTable[addr]
}

func (tb *testBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
	fcoins, fexists := tb.coinTable[from]
	if !fexists {
		panic(fmt.Sprintf(
			"source address %s does not exist",
			from.String()))
	}
	if !fcoins.IsAllGTE(amt) {
		panic(fmt.Sprintf(
			"source address %s has %s; cannot send %s",
			from.String(), fcoins, amt))
	}
	// First, subtract from 'from'.
	frest := fcoins.Sub(amt)
	tb.coinTable[from] = frest
	// Second, add to 'to'.
	// NOTE: even works when from==to, due to 2-step isolation.
	tcoins, _ := tb.coinTable[to]
	tsum := tcoins.Add(amt)
	tb.coinTable[to] = tsum
}

func (tb *testBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (tb *testBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	sum := coins.Add(std.Coins{{denom, amt}})
	tb.coinTable[addr] = sum
}

func (tb *testBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	rest := coins.Sub(std.Coins{{denom, amt}})
	tb.coinTable[addr] = rest
}

func parseFiles(fnames []string, stderr io.WriteCloser) ([]*gno.FileNode, error) {
	files := make([]*gno.FileNode, 0, len(fnames))
	var hasError bool
	for _, fname := range fnames {
		if s, err := os.Stat(fname); err == nil && s.IsDir() {
			subFns, err := listNonTestFiles(fname)
			if err != nil {
				return nil, err
			}
			subFiles, err := parseFiles(subFns, stderr)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
			continue
		} else if err != nil {
			// either not found or some other kind of error --
			// in either case not a file we can parse.
			return nil, err
		}

		hasError = catchRuntimeError(fname, stderr, func() {
			files = append(files, gno.MustReadFile(fname))
		})
	}

	if hasError {
		return nil, commands.ExitCodeError(1)
	}
	return files, nil
}

func listNonTestFiles(dir string) ([]string, error) {
	fs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fn := make([]string, 0, len(fs))
	for _, f := range fs {
		n := f.Name()
		if isGnoFile(f) &&
			!strings.HasSuffix(n, "_test.gno") &&
			!strings.HasSuffix(n, "_filetest.gno") {
			fn = append(fn, filepath.Join(dir, n))
		}
	}
	return fn, nil
}

func runExpr(m *gno.Machine, expr string) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case gno.UnhandledPanicError:
				fmt.Printf("panic running expression %s: %v\nStacktrace: %s\n",
					expr, r.Error(), m.ExceptionsStacktrace())
			default:
				fmt.Printf("panic running expression %s: %v\nMachine State:%s\nStacktrace: %s\n",
					expr, r, m.String(), m.Stacktrace().String())
			}
			panic(r)
		}
	}()
	ex, err := gno.ParseExpr(expr)
	if err != nil {
		panic(fmt.Errorf("could not parse: %w", err))
	}
	m.Eval(ex)
}
