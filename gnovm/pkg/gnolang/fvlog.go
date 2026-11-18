package gnolang

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Foreign-value probe: with GNO_FV_LOG=<file> set, log every struct or array
// value of a realm's type born or copied while another realm is current.
// Diagnostic only: it changes no behaviour and charges no gas.

var (
	fvOnce sync.Once
	fvMu   sync.Mutex
	fvFile *os.File
)

// fvOn reports whether the probe is enabled, opening the log file once.
func fvOn() bool {
	fvOnce.Do(func() {
		p := os.Getenv("GNO_FV_LOG")
		if p == "" {
			return
		}
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			panic(fmt.Sprintf("GNO_FV_LOG: %v", err))
		}
		fvFile = f
	})
	return fvFile != nil
}

// fvLog logs a birth/copy/arg/ret of a value of type t; see fvlog.go header.
// site "" means: name the first Go caller outside the Copy methods.
func fvLog(kind string, t Type, alloc *Allocator, site string) {
	if !fvOn() || t == nil || alloc == nil {
		return
	}
	cur, code, caller := alloc.currentRealmPath, "-", "-"
	if alloc.fvCtx != nil {
		code, caller = alloc.fvCtx()
	}
	var refs []string
	if kind == "arg" || kind == "ret" {
		refs = []string{caller}
	} else {
		refs = []string{cur, code}
	}
	if !fvForeign(t, refs) {
		return
	}
	if site == "" {
		site = fvCopySite()
	}
	if cur == "" {
		cur = "-"
	}
	line := fmt.Sprintf("fv %s %s in %s code=%s caller=%s at %s\n",
		kind, t.String(), cur, code, caller, site)
	fvMu.Lock()
	fvFile.WriteString(line)
	fvMu.Unlock()
}

// fvForeign reports whether t, or a type it holds inline (array element,
// struct field), is declared by a realm other than one of refs ("-" = unknown).
func fvForeign(t Type, refs []string) bool {
	differs := func(path string, pid PkgID) bool {
		if !pid.IsRealmPkg() {
			return false
		}
		for _, r := range refs {
			if r != "-" && r != path {
				return true
			}
		}
		return false
	}
	switch tt := t.(type) {
	case *DeclaredType:
		return differs(tt.PkgPath, tt.GetPkgID()) || fvForeign(tt.Base, refs)
	case *StructType:
		if differs(tt.PkgPath, tt.GetPkgID()) {
			return true
		}
		for _, f := range tt.Fields {
			if fvForeign(f.Type, refs) {
				return true
			}
		}
	case *ArrayType:
		return fvForeign(tt.Elt, refs)
	}
	return false
}

// fvCopySite names the first Go caller that is not a Copy method or the probe.
func fvCopySite() string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		fr, more := frames.Next()
		name := fr.Function[strings.LastIndexByte(fr.Function, '/')+1:]
		name = strings.TrimPrefix(name, "gnolang.")
		if !strings.HasSuffix(name, ".Copy") && !strings.HasPrefix(name, "fv") {
			return name
		}
		if !more {
			return "-"
		}
	}
}

// fvFrames returns the package of the innermost call frame and of the one below.
func (m *Machine) fvFrames() (code, caller string) {
	code, caller = "-", "-"
	found := false
	for i := len(m.Frames) - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if !fr.IsCall() || fr.Func == nil {
			continue
		}
		p := fr.Func.PkgPath
		if p == "" {
			p = "-"
		}
		if found {
			return code, p
		}
		code, found = p, true
	}
	return code, caller
}

// fvLogValues logs each struct/array value of tvs as kind (arg or ret).
func (m *Machine) fvLogValues(kind string, tvs []TypedValue, site string) {
	for _, tv := range tvs {
		if hiv, ok := tv.V.(*HeapItemValue); ok {
			tv = hiv.Value
		}
		if tv.T == nil {
			continue
		}
		if k := baseOf(tv.T).Kind(); k == StructKind || k == ArrayKind {
			fvLog(kind, tv.T, m.Alloc, site)
		}
	}
}

// fvLogResults logs the results of the innermost call frame before it pops.
func (m *Machine) fvLogResults(site string, fromBlock bool) {
	if !fvOn() {
		return
	}
	cfr := m.PeekCallFrame(1)
	if cfr == nil || cfr.Func == nil {
		return
	}
	ft := cfr.Func.GetType(m.Store)
	n := len(ft.Results)
	if fromBlock && cfr.NumBlocks < len(m.Blocks) {
		np := len(ft.Params)
		vals := m.Blocks[cfr.NumBlocks].Values
		if np+n <= len(vals) {
			m.fvLogValues("ret", vals[np:np+n], site)
		}
		return
	}
	if !fromBlock && n <= len(m.Values) {
		m.fvLogValues("ret", m.PeekValues(n), site)
	}
}
