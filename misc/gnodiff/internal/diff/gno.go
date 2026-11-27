package diff

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type gnoStage uint8

const (
	stageSetup gnoStage = iota
	stageLoad
	stageRun
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
		stage = stageSetup
	)
	defer func() {
		if r := recover(); r != nil {
			res = resultFromRecover(r, stage)
			if res.Kind == KindPanic || res.Kind == KindNonTerm {
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

	stage = stageLoad
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		res = Result{Kind: KindReject, Detail: clip(err.Error(), detailLimit), Truncated: out.truncated}
		return res
	}
	m.RunFiles(fn)
	stage = stageRun
	m.RunMain()
	res = Result{
		Kind:      KindOK,
		Stdout:    NormalizeOutput(out.String()),
		Truncated: out.truncated,
	}
	return res
}

func resultFromRecover(r any, stage gnoStage) Result {
	switch e := r.(type) {
	case store.OutOfGasError:
		return Result{Kind: KindNonTerm, Detail: e.Error()}
	case *store.OutOfGasError:
		return Result{Kind: KindNonTerm, Detail: e.Error()}
	case *gnolang.PreprocessError:
		return resultFromPreprocess(e)
	case gnolang.UnhandledPanicError:
		return Result{Kind: KindPanic, Panic: firstLine(e.Error())}
	case *gnolang.Exception:
		return Result{Kind: KindPanic, Panic: firstLine(fmt.Sprint(e))}
	case string:
		if strings.HasPrefix(e, "allocation limit exceeded") {
			return Result{Kind: KindNonTerm, Detail: e}
		}
		if stage == stageLoad && looksLikeCompile(e) {
			return Result{Kind: KindReject, Detail: clip(e, detailLimit)}
		}
		return Result{Kind: KindInternal, Detail: clip(e, detailLimit)}
	case error:
		if oog, ok := asOutOfGas(e); ok {
			return Result{Kind: KindNonTerm, Detail: oog}
		}
		var rtErr runtime.Error
		if errors.As(e, &rtErr) {
			return Result{Kind: KindInternal, Detail: clip(e.Error(), detailLimit)}
		}
		msg := e.Error()
		if strings.HasPrefix(msg, "allocation limit exceeded") {
			return Result{Kind: KindNonTerm, Detail: clip(msg, detailLimit)}
		}
		if stage == stageLoad && looksLikeCompile(msg) {
			return Result{Kind: KindReject, Detail: clip(msg, detailLimit)}
		}
		return Result{Kind: KindInternal, Detail: clip(msg, detailLimit)}
	default:
		return Result{Kind: KindInternal, Detail: clip(fmt.Sprint(r), detailLimit)}
	}
}

func resultFromPreprocess(pe *gnolang.PreprocessError) Result {
	if oog, ok := asOutOfGas(pe); ok {
		return Result{Kind: KindNonTerm, Detail: oog}
	}
	var rtErr runtime.Error
	if errors.As(pe, &rtErr) {
		return Result{Kind: KindInternal, Detail: clip(pe.Error(), detailLimit)}
	}
	inner := pe.Unwrap()
	msg := pe.Error()
	if inner != nil {
		msg = inner.Error()
	}
	if strings.Contains(msg, "allocation limit exceeded") {
		return Result{Kind: KindNonTerm, Detail: clip(msg, detailLimit)}
	}
	if strings.Contains(msg, "should not happen") {
		return Result{Kind: KindInternal, Detail: clip(msg, detailLimit)}
	}
	return Result{Kind: KindReject, Detail: clip(msg, detailLimit)}
}

func asOutOfGas(err error) (string, bool) {
	var oog store.OutOfGasError
	if errors.As(err, &oog) {
		return oog.Error(), true
	}
	var oogPtr *store.OutOfGasError
	if errors.As(err, &oogPtr) && oogPtr != nil {
		return oogPtr.Error(), true
	}
	return "", false
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
		"unknown import path",
		"redeclared",
		"syntax",
		"parse error",
		"illegal",
		"overflows",
		"underflows",
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
		if strings.Contains(low, n) {
			return true
		}
	}
	return false
}
