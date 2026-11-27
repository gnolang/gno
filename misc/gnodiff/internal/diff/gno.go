package diff

import (
	"fmt"
	"io"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// RunGno executes src on an in-process GnoVM machine, the same shape gno run
// and the multiarch-determinism driver use: a prod store, ParseFile, RunFiles,
// RunMain. A gas meter bounds non-termination. Gas is recorded on the result
// and is not part of the comparison.
func RunGno(src string, opt Options) (res Result) {
	opt, err := opt.resolve()
	if err != nil {
		return Result{Kind: KindInternal, Detail: err.Error()}
	}

	var out cappedBuf
	out.init(opt.OutputLimit)
	output := test.OutputWithError(&out, io.Discard)

	var (
		meter store.GasMeter
		m     *gnolang.Machine
	)
	defer func() {
		if r := recover(); r != nil {
			res = resultFromRecover(r)
			if res.Kind == KindPanic {
				res.Stdout = NormalizeOutput(out.String())
			}
			res.Truncated = out.truncated
		}
		if meter != nil {
			res.Gas = meter.GasConsumed()
		}
		if m != nil {
			m.Release()
		}
	}()

	_, gnoStore := test.ProdStore(opt.RootDir, output, nil)
	meter = store.NewGasMeter(opt.GasLimit)

	m = gnolang.NewMachineWithOptions(gnolang.MachineOptions{
		PkgPath:       "main",
		Output:        output,
		Store:         gnoStore,
		MaxAllocBytes: opt.MaxAlloc,
		Context:       test.Context("", "main", nil),
		GasMeter:      meter,
	})

	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		res = Result{Kind: KindReject, Detail: clip(err.Error(), detailLimit), Truncated: out.truncated}
		return res
	}
	m.RunFiles(fn)
	m.RunMain()
	res = Result{
		Kind:      KindOK,
		Stdout:    NormalizeOutput(out.String()),
		Truncated: out.truncated,
	}
	return res
}

func resultFromRecover(r any) Result {
	switch e := r.(type) {
	case store.OutOfGasError:
		return Result{Kind: KindNonTerm, Detail: e.Error()}
	case *store.OutOfGasError:
		return Result{Kind: KindNonTerm, Detail: e.Error()}
	case gnolang.UnhandledPanicError:
		return Result{Kind: KindPanic, Panic: firstLine(e.Error())}
	case *gnolang.Exception:
		return Result{Kind: KindPanic, Panic: firstLine(fmt.Sprint(e))}
	case string:
		if strings.HasPrefix(e, "allocation limit exceeded") {
			return Result{Kind: KindNonTerm, Detail: e}
		}
		if looksLikeCompile(e) {
			return Result{Kind: KindReject, Detail: clip(e, detailLimit)}
		}
		return Result{Kind: KindInternal, Detail: clip(e, detailLimit)}
	case error:
		msg := e.Error()
		if looksLikeCompile(msg) {
			return Result{Kind: KindReject, Detail: clip(msg, detailLimit)}
		}
		return Result{Kind: KindInternal, Detail: clip(msg, detailLimit)}
	default:
		return Result{Kind: KindInternal, Detail: clip(fmt.Sprint(r), detailLimit)}
	}
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func looksLikeCompile(msg string) bool {
	needles := []string{
		"invalid operation",
		"expected ",
		"undefined:",
		"cannot ",
		"mismatched",
		"not enough arguments",
		"too many arguments",
		"declared and not used",
		"imported and not used",
		"missing return",
		"is not a type",
		"unknown field",
		"unknown method",
		"redeclared",
		"syntax",
		"parse error",
		"illegal",
		"overflows",
		"does not support",
		"invalid type",
		"cannot use",
		"cannot convert",
		"assignment mismatch",
		"multiple-value",
		"evaluated but not used",
		"no new variables",
		".gno:",
		".go:",
	}
	low := strings.ToLower(msg)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
