package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
)

type ReplOption func(*Repl)

// WithStore allows to modify the default Store implementation used by the VM.
// If nil is provided, the VM will use a default implementation.
func WithStore(s gno.Store) ReplOption {
	return func(r *Repl) {
		r.store = s
	}
}

func WithIO(input io.Reader, output, errput io.Writer) ReplOption {
	return func(r *Repl) {
		r.input = input
		r.output = output
		r.errput = errput
	}
}

type Repl struct {
	m *gno.Machine

	// package+file is the top-most scope.
	pv *gno.PackageValue
	pn *gno.PackageNode
	fn *gno.FileNode // contains .bs and func/type/import decls
	fb *gno.Block    // file block for .fn

	rec any // last exception recovered

	// rw joins stdout and stderr to give an unified output and group with stdin.
	rw *bufio.ReadWriter

	// Repl options:
	pkgPath string
	output  io.Writer // machine output
	errput  io.Writer // repl printing of errors
	input   io.Reader
	store   gno.Store
	debug   bool
}

// NewRepl creates a Repl struct. It is able to process input source code and eventually run it.
func NewRepl(opts ...ReplOption) *Repl {
	r := &Repl{}

	// init with defaults and config.
	r.pkgPath = "repl"
	r.input = os.Stdin
	r.output = os.Stdout
	r.errput = os.Stderr
	_, r.store = test.TestStore(gnoenv.RootDir(), test.OutputWithError(r.output, r.errput), nil)

	var nilAllocator = (*gno.Allocator)(nil)
	r.pn = gno.NewPackageNode("repl", r.pkgPath, &gno.FileSet{})
	r.pv = r.pn.NewPackage(nilAllocator)
	r.fn = &gno.FileNode{
		FileName: "<repl>",
		PkgName:  "repl",
		Decls:    nil,
	}
	r.fb = gno.NewBlock(nilAllocator, r.fn, r.pv.GetBlock(r.store))
	for _, opt := range opts {
		opt(r)
	}

	// register package node and value.
	r.store.SetBlockNode(r.pn)
	r.store.SetCachePackage(r.pv)

	// construct machine.
	input := bufio.NewReader(r.input)
	output := bufio.NewWriter(r.output)
	r.rw = bufio.NewReadWriter(input, output)
	r.m = gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: r.pkgPath,
		Debug:   r.debug,
		Input:   input,
		Output:  output,
		Store:   r.store,
	})
	r.m.SetActivePackage(r.pv)

	// preprocess nodes.
	r.fn = gno.Preprocess(r.store, r.pn, r.fn).(*gno.FileNode)

	// set blocks.
	// r.m.PushBlock(r.fb)

	return r
}

func (r *Repl) Print(args ...any) {
	fmt.Fprint(r.output, args...)
}

func (r *Repl) Printf(fstr string, args ...any) {
	fmt.Fprintf(r.output, fstr, args...)
}

func (r *Repl) Printfln(fstr string, args ...any) {
	fmt.Fprintf(r.output, fstr+"\n", args...)
}

func (r *Repl) Println(args ...any) {
	fmt.Fprintln(r.output, args...)
}

func (r *Repl) Errorf(fstr string, args ...any) {
	fmt.Fprintf(r.errput, fstr, args...)
}

func (r *Repl) Errorfln(fstr string, args ...any) {
	fmt.Fprintf(r.errput, fstr+"\n", args...)
}

func (r *Repl) Errorln(args ...any) {
	fmt.Fprintln(r.errput, args...)
}

func (r *Repl) RunStatements(code string) {
	if os.Getenv("DEBUG_PANIC") != "1" {
		defer func() {
			if rec := recover(); rec != nil {
				r.rec = rec
				switch rec := rec.(type) {
				case *gno.PreprocessError:
					err := rec.Unwrap()
					match := gno.ReErrorLine.Match(err.Error())
					if match == nil {
						r.Errorln(err.Error())
					} else {
						r.Errorln(match.Get("MSG"))
					}
				case error:
					err := rec
					match := gno.ReErrorLine.Match(err.Error())
					if match == nil {
						r.Errorln(err.Error())
					} else {
						r.Errorln(match.Get("MSG"))
					}
				}
			}
		}()
	}

	if r.debug {
		// Activate debugger for this statement only.
		r.m.Debugger.Enable(os.Stdin, os.Stdout, func(ppath, file string) string { return code })
		r.debug = false
		defer r.m.Debugger.Disable()
	}

	decls, err := r.m.ParseDecls(code)
	if err != nil {
		stmts, err2 := r.m.ParseStmts(code)
		if err2 != nil {
			r.Errorln(err2.Error())
			return
		}
		// e.g. var a = 1; or b := 2
		for _, stmt := range stmts {
			r.m.RunStatement(gno.StageRun, stmt)
			r.rw.Flush()
		}
	} else {
		// e.g. import "bytes"; or func (Foo)Bar(){}
		for _, decl := range decls {
			r.m.RunDeclaration(decl)
			r.rw.Flush()
		}
	}
}

// Reset will reset the actual repl state, restarting the internal VM.
func (r *Repl) Reset() {
	panic("not yet implemented")
}

// Debug activates the GnoVM debugger for the next evaluation.
func (r *Repl) Debug() {
	r.debug = true
}
