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

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
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
			ShortHelp:  "run gno packages",
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
	_, testStore := test.Store(
		cfg.rootDir, false,
		stdin, stdout, stderr)
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

	var send std.Coins
	pkgPath := string(files[0].PkgName)
	ctx := test.Context(pkgPath, send)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: pkgPath,
		Output:  stdout,
		Input:   stdin,
		Store:   testStore,
		Context: ctx,
		Debug:   cfg.debug || cfg.debugAddr != "",
	})

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

// runGNOinGo는 Gno 익명 함수를 문자열로 받아 실행하는 함수입니다.
func RunGNOinGo(expr string) error {
	// 1. Gno 환경의 루트 디렉터리를 결정합니다.
	rootDir := gnoenv.RootDir()

	// 2. 테스트용 스토어를 초기화합니다.
	_, testStore := test.Store(rootDir, false, os.Stdin, os.Stdout, os.Stderr)

	// 3. 기본 패키지 경로와 컨텍스트를 설정합니다.
	pkgPath := "main"
	ctx := test.Context(pkgPath, std.Coins{})

	// 4. 가상 머신을 생성합니다.
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: pkgPath,
		Output:  os.Stdout,
		Input:   os.Stdin,
		Store:   testStore,
		Context: ctx,
		Debug:   false, // 필요에 따라 true로 설정
	})
	defer m.Release()

	// 5. runExpr를 사용하여 전달된 문자열 코드를 실행합니다.
	runExpr(m, expr)

	return nil
}
