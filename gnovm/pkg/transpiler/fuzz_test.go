package transpiler_test

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

func FuzzTranspiling(f *testing.F) {
	if testing.Short() {
		f.Skip("Running in -short mode")
	}

	// 1. Derive the seeds from our seedGnoFiles.
	addGnoExamplesAsSeedsToFuzzer(f)

	// 2. Run the fuzzer.
	f.Fuzz(func(t *testing.T, gnoSourceCode []byte) {
		var m *gnolang.Machine
		gnoSrc := string(gnoSourceCode)
		fn, err := m.ParseFile("main.go", string(gnoSourceCode))
		if err != nil {
			// TODO: it could be discrepancy that if it compiled alright that it later failed.
			return
		}

		if !strings.Contains(gnoSrc, "func main()") {
			gnoSrc += "\n\nfunc main() {}"
		}
		memPkg := &gnovm.MemPackage{
			Name: string(fn.PkgName),
			Path: string(fn.Name),
			Files: []*gnovm.MemFile{
				{Name: "a.gno", Body: gnoSrc},
			},
		}
		if err := gnolang.TypeCheckMemPackage(memPkg, mockPackageGetter{}, false); err != nil {
			// Confirmed that the code won't comnpile in Go either. This step is crucial
			// as we were advised that before gnokey, we use this type checker.
			return
		}

		// 3. Add timings to ensure that if transpiling takes a long time
		// to run, that we report this as problematic.
		doneCh := make(chan bool, 1)
		readyCh := make(chan bool)

		go func() {
			close(readyCh)
			defer close(doneCh)
			_, _ = transpiler.Transpile(string(gnoSourceCode), "gno", "in.gno")
			doneCh <- true
		}()

		<-readyCh

		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("took more than 5 seconds to transpile\n\n%s", gnoSourceCode)
		case <-doneCh:
		}
	})
}

func addGnoExamplesAsSeedsToFuzzer(f *testing.F) {
	f.Helper()

	ffs := os.DirFS(filepath.Join(gnoenv.RootDir(), "examples"))
	fs.WalkDir(ffs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		if !strings.HasSuffix(path, ".gno") {
			return nil
		}
		file, err := ffs.Open(path)
		if err != nil {
			panic(err)
		}
		blob, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			panic(err)
		}
		f.Add(blob)
		return nil
	})
}

func FuzzTypecheckThenGnoRunMemPackageVsCompileGo(f *testing.F) {
	if testing.Short() {
		f.Skip("Running in -short mode")
	}

	// 1. Derive the seeds from our seedGnoFiles.
	addGnoExamplesAsSeedsToFuzzer(f)

	// 2. Run the fuzzers.
	f.Fuzz(func(t *testing.T, gnoSourceCode []byte) {
		gnoSrc := string(gnoSourceCode)
		// 3. Add timings to ensure that if transpiling takes a long time
		// to run, that we report this as problematic.
		isGnoTypeCheckError := false
		defer func() {
			goRunErr := checkIfGoCompilesProgram(t, gnoSrc)
			r := recover()
			if r == nil {
				if goRunErr != nil {
					if !isGnoTypeCheckError {
						panic(fmt.Sprintf("Runs alright in Gno but fails in Go:\n%v\n%s", goRunErr, gnoSrc))
					}
				}
				return
			}

			sr := fmt.Sprintf("%s", r)

			// It recovered all good so one way that we can sort out common
			// syntax issues is by substring matches between Go and Gno and if
			// we have a strong match, we can return.
			if goRunErr != nil {
				// Normalize between Go and Gno as Go at times prefixes with:
				// * "syntax error: "
				goRunErrOutput := strings.ReplaceAll(goRunErr.output, "syntax error: ", "")

				// Gno at times adds the code.
				if index := strings.Index(sr, "(code="); index >= 0 {
					sr = sr[:index]
				}

				goRunErrOutput = strings.TrimSpace(goRunErrOutput)
				sr = strings.TrimSpace(sr)
				if sr == goRunErrOutput { // We've got exact matching errors so can exit.
					return
				}
			}

			// Otherwise we could not find matches between Go and Gno,
			// mos def custom syntax errors.

			switch {
			// Legitimate invalid syntax, compile problems that are common between
			// Go and Gno.
			case strings.Contains(sr, "invalid line number "),
				strings.Contains(sr, "not defined in fileset with files"),
				strings.Contains(sr, "unknown import path"),
				strings.Contains(sr, "redeclared in this block"),
				strings.Contains(sr, "invalid recursive type"),
				strings.Contains(sr, "does not have a body but is not natively defined"),
				strings.Contains(sr, "invalid operation: division by zero"),
				strings.Contains(sr, "not declared"),
				strings.Contains(sr, "not defined in fileset with"),
				strings.Contains(sr, "literal not terminated"),
				strings.Contains(sr, "illegal character"),
				strings.Contains(sr, "expected 1 expression"),
				strings.Contains(sr, "expected 'IDENT', found "),
				strings.Contains(sr, "expected declaration, found"),
				strings.Contains(sr, "expected 'package', found"),
				strings.Contains(sr, "expected type, found newline"),
				strings.Contains(sr, "illegal UTF-8 encoding"),
				strings.Contains(sr, "in octal literal"),
				strings.Contains(sr, "missing import path"),
				strings.Contains(sr, "expected type, found"),
				strings.Contains(sr, "expected ')', found newline"),
				strings.Contains(sr, "missing parameter name"),
				strings.Contains(sr, "literal has no digits"),
				strings.Contains(sr, "constant definition loop with"),
				strings.Contains(sr, "missing ',' in parameter list"),
				strings.Contains(sr, "missing ',' in argument list"),
				strings.Contains(sr, "redeclarations for identifier"),
				strings.Contains(sr, "required in 3-index slice"),
				strings.Contains(sr, "comment not terminated"),
				strings.Contains(sr, "missing field"),
				strings.Contains(sr, "expected operand, found"),
				strings.Contains(sr, "expected statement, found"),
				strings.Contains(sr, "m.NumValues <= 0"),
				strings.Contains(sr, "missing ',' in composite literal"),
				strings.Contains(sr, "no new variables on left side of"),
				strings.Contains(sr, "expected boolean or range expression, found assignment (missing parentheses around composite"),
				strings.Contains(sr, "must separate successive digits"),
				strings.Contains(sr, "runtime error: invalid memory address") && strings.Contains(gnoSrc, " int."),
				strings.Contains(sr, "expected '{', found "),
				strings.Contains(sr, "ast.FuncDecl has missing receiver"),
				strings.Contains(sr, "expected '}', found "),
				strings.Contains(sr, "dot imports not allowed"),
				strings.Contains(sr, "expected '(', found "),
				strings.Contains(sr, "expected ')', found "),
				strings.Contains(sr, "expected '[', found "),
				strings.Contains(sr, "expected ']', found "),
				strings.Contains(sr, "invalid digit"),
				strings.Contains(sr, "missing ',' before newline in argument list"),
				strings.Contains(sr, "import path must be a string"),
				strings.Contains(sr, "cannot indirect"),
				strings.Contains(sr, "invalid radix point in"),
				strings.Contains(sr, "invalid column number"),
				strings.Contains(sr, "unknown Go type *ast.IndexListExpr"),
				strings.Contains(sr, "expected selector or type assertion"),
				strings.Contains(sr, "cannot take address of"),
				strings.Contains(sr, "unexpected selector expression type"),
				strings.Contains(sr, "hexadecimal mantissa requires a 'p' exponent"),
				strings.Contains(sr, "invalid operation: operator"),
				strings.Contains(sr, "operator") && strings.Contains(sr, "not defined on"),
				strings.Contains(sr, "illegal rune literal"),
				strings.Contains(sr, "DeclaredType method named"),
				strings.Contains(sr, "invalid line number"),
				strings.Contains(sr, "missing ',' before newline in parameter list"),
				strings.Contains(sr, "expected ';', found "),
				strings.Contains(sr, "goroutines are not permitted"),
				strings.Contains(sr, "method has no receiver"),
				strings.Contains(sr, "more than one index"),
				strings.Contains(sr, "unknown escape sequence"),
				strings.Contains(sr, "expected '==', found '='"),
				strings.Contains(sr, "curly quotation mark") && strings.Contains(sr, `(use neutral '"')`),
				strings.Contains(sr, "invalid package name"),
				strings.Contains(sr, " exponent requires hexadecimal mantissa"),
				strings.Contains(sr, " exponent requires decimal mantissa"),
				strings.Contains(sr, "label ") && strings.Contains(sr, "undefined"),
				strings.Contains(sr, "imaginaries are not supported"),
				strings.Contains(sr, "missing parameter type"),
				strings.Contains(sr, "exponent has no digits"),
				strings.Contains(sr, "expected ':', found "),
				strings.Contains(sr, "expected if statement or block"),
				strings.Contains(sr, "illegal label declaration"),
				strings.Contains(sr, "expected boolean or range expression"),
				strings.Contains(sr, "expected type argument list"),
				strings.Contains(sr, "unexpected comma; expecting"),
				strings.Contains(sr, "expected ~ term or type"),
				strings.Contains(sr, "expected boolean expression, found assignment"),
				strings.Contains(sr, "missing ',' before newline in composite literal"),
				strings.Contains(sr, "cannot parenthesize type in composite literal"),
				strings.Contains(sr, "expression in defer must be function call"),
				strings.Contains(sr, "expected switch expression, found"),
				strings.Contains(sr, "missing ',' in type"),
				strings.Contains(sr, "escape sequence is invalid Unicode code"),
				strings.Contains(sr, "unexpected semicolon or newline before"),
				strings.Contains(sr, "missing type parameter name"),
				strings.Contains(sr, "empty type parameter"),
				strings.Contains(sr, "mixed named and unnamed"),
				strings.Contains(sr, "expression in go must be function call"),
				strings.Contains(sr, "unexpected newline, expecting"),
				strings.Contains(sr, "imports must appear before"),
				strings.Contains(sr, "missing condition in"),
				strings.Contains(sr, "method has multiple receivers"),
				strings.Contains(sr, "expected 'chan', found"),
				strings.Contains(sr, "missing type constraint"),
				strings.Contains(sr, "expected at most 2 expressions"),
				strings.Contains(sr, "expression in defer must not be parenthesized"),
				strings.Contains(sr, "not yet implemented"),
				strings.Contains(sr, "builtin identifiers cannot be shadowed"),
				strings.Contains(sr, "var declaration not allowed in if"),
				strings.Contains(sr, "EmptyStmt not supported"),
				strings.Contains(sr, "expected boolean expression, found"),
				strings.Contains(sr, "cannot parenthesize embedded type"),
				strings.Contains(sr, "method must have no type parameters"),
				strings.Contains(sr, "function type must have no type parameters"),
				strings.Contains(sr, "missing ',' before newline in type argument list"),
				strings.Contains(sr, "expected declaration, found con"),
				strings.Contains(sr, "escape sequence not terminated"):
				return

			default:
				if goRunErr == nil {
					panic(fmt.Sprintf("Discrepancy; runs alright in Go, fails in Gno:\n%s\n\033[33m%s\033[00m\n", r, gnoSrc))
				}

				panic(fmt.Sprintf("%s\n\nfailure due to:\n\033[31m%s\033[00m", sr, gnoSrc))
			}
		}()

		// Next run the code to see if it can be ran.
		var m *gnolang.Machine
		fn, err := m.ParseFile("main.go", string(gnoSourceCode))
		if err != nil {
			// TODO: it could be discrepancy that if it compiled alright that it later failed.
			panic(err)
		}

		if !strings.Contains(gnoSrc, "func main()") {
			gnoSrc += "\n\nfunc main() {}"
		}

		db := memdb.NewMemDB()
		baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
		iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
		store := gnolang.NewStore(nil, baseStore, iavlStore)
		m = gnolang.NewMachine(string(fn.PkgName), store)
		memPkg := &gnovm.MemPackage{
			Name: string(fn.PkgName),
			Path: string(fn.Name),
			Files: []*gnovm.MemFile{
				{Name: "a.gno", Body: gnoSrc},
			},
		}
		if err := gnolang.TypeCheckMemPackage(memPkg, mockPackageGetter{}, false); err != nil {
			isGnoTypeCheckError = true
			return
		}
		m.RunMemPackage(memPkg, true)
	})
}

type mockPackageGetter []*gnovm.MemPackage

func (mi mockPackageGetter) GetMemPackage(path string) *gnovm.MemPackage {
	for _, pkg := range mi {
		if pkg.Path == path {
			return pkg
		}
	}
	return nil
}

func checkIfGoCompilesProgram(tb testing.TB, src string) *comparisonError {
	tb.Helper()
	dir := tb.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte(src), 0o755); err != nil {
		// Swallow up this error.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "tool", "compile", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Re-Trim the input to remove the path prefix.
		normalizedOutput := strings.TrimSpace("main.go" + strings.TrimPrefix(string(output), path))
		if strings.Contains(normalizedOutput, "not a main package") {
			return nil
		}
		return &comparisonError{err, normalizedOutput, src}
	}
	return nil
}

type comparisonError struct {
	cmdErr error
	output string
	src    string
}

func (ce *comparisonError) Error() string {
	return fmt.Sprintf("%v\n%s\n%s", ce.cmdErr, ce.output, ce.src)
}
