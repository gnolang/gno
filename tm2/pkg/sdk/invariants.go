package sdk

import (
	"fmt"
	"strings"
)

// An Invariant is a function which tests a particular invariant.
// The invariant returns a descriptive message about what happened
// and a boolean indicating whether the invariant has been broken.
// The simulator will then halt and print the logs.
type Invariant func(ctx Context) (string, bool)

// Invariants defines a group of invariants
type Invariants []Invariant

// expected interface for registering invariants
type InvariantRegistry interface {
	RegisterRoute(moduleName, route string, invar Invariant)
}

// FormatInvariant returns a standardized invariant message.
func FormatInvariant(module, name, msg string) string {
	return fmt.Sprintf("%s: %s invariant\n%s\n", module, name, msg)
}

// InvariantReport accumulates the findings of one invariant.
//
// Findings are counted in full but only the first reportCap are formatted. An
// invariant runs over whole keyspaces, so a genuinely corrupt store can produce
// millions of findings — building that message would exhaust memory inside the
// check that exists to diagnose the corruption.
type InvariantReport struct {
	module, name string
	b            strings.Builder
	found        int
}

// reportCap is how many findings are described individually. The rest are counted.
const reportCap = 10

// Addf records one finding. It is the only method a check body needs; Guard
// builds the report and renders it.
func (r *InvariantReport) Addf(format string, args ...any) {
	r.found++
	if r.found <= reportCap {
		fmt.Fprintf(&r.b, "\t"+format+"\n", args...)
	}
}

// result renders the report in the form Invariant returns.
func (r *InvariantReport) result() (string, bool) {
	if r.found == 0 {
		return FormatInvariant(r.module, r.name, "no violations found"), false
	}
	msg := fmt.Sprintf("%d violation(s) found", r.found)
	if r.found > reportCap {
		msg += fmt.Sprintf(", first %d shown", reportCap)
	}
	return FormatInvariant(r.module, r.name, msg+":\n"+r.b.String()), true
}

// Guard turns a check body into an Invariant, converting a panic into a broken
// result rather than letting it escape.
//
// This is a backstop, not the mechanism. An invariant inspects exactly the state
// its ordinary callers panic on, so the checks are written to read raw state with
// non-panicking helpers; see the bank and auth invariants. The guard exists because
// a check that panics takes the node down with a stack trace instead of producing
// the report it exists for — and because a bug in a check must be louder than a
// violation, not quieter, which is why the recovered value is reported as broken.
//
// It cannot catch everything: an out-of-memory allocation is fatal and unrecoverable,
// which is why the checks bound their own allocations rather than relying on this.
func Guard(module, name string, body func(Context, *InvariantReport)) Invariant {
	return func(ctx Context) (string, bool) {
		rep := &InvariantReport{module: module, name: name}
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					rep.Addf("check panicked (this is a bug in the check, or state it cannot read): %v", rec)
				}
			}()
			body(ctx, rep)
		}()
		return rep.result()
	}
}
