package vm

import (
	"errors"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cmnerrors "github.com/gnolang/gno/tm2/pkg/errors"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	// maxBoundedBytes caps the rendered output of a single
	// recover-path string (panic value, error, etc.) — produced by
	// boundedString. Output ≤ maxBoundedBytes (or +3 on truncation).
	maxBoundedBytes = 1024

	// maxBoundedStringDepth bounds recursion via Unwrap() to avoid
	// stack/heap exhaustion on adversarial wrap chains.
	maxBoundedStringDepth = 8
)

// boundedString renders v as a string capped at maxBoundedBytes
// (with "..." suffix on truncation).
//
// PARANOID WHITELIST: only types we have verified are bounded
// invoke their .Error() / .String() methods. For anything else,
// returns "<%T>" (Go type name, source-bounded).
//
// Specifically does NOT call cmnError.Error() because that calls
// fmt.Sprintf("%v", err) which routes through cmnError.Format and
// renders msgtraces + stacktrace + fmt-of-data, all unbounded. For
// cmnError, we peek at Data() and use FmtError.Format() (returns
// raw format string, no Sprintf) when applicable.
//
// Pass depth=0 from external callers; recursive Unwrap calls
// increment up to maxBoundedStringDepth.
func boundedString(v any, depth int) string {
	if depth >= maxBoundedStringDepth {
		return "<unwrap-depth-exceeded>"
	}
	switch x := v.(type) {
	case nil:
		return "<nil>"
	case string:
		return truncate(x)
	case []byte:
		return truncate(string(x))

	// Gno-specific bounded types
	case *gno.Exception:
		// Bounded by BoundedSprintException + the BoundedPanicRender
		// flag on Machines that built this exception. m=nil here
		// because boundedString may be called from non-VM contexts;
		// composites render structurally (no user .Error() dispatch).
		return gno.BoundedSprintException(x, nil, maxBoundedBytes)
	case *gno.PreprocessError:
		// Bounded by the earlier PreprocessError.Stack() fix in
		// gnovm/pkg/gnolang/debug.go.
		return truncate(x.Error())
	case gno.UnhandledPanicError:
		// Value type — must precede the generic `error` arm.
		// Descriptor is bounded at construction when the constructing
		// Machine had BoundedPanicRender=true (op_call.go flag-true
		// branch). Validators set the flag on every m.
		return truncate(x.Descriptor)

	// tm2-specific bounded types
	case stypes.OutOfGasError:
		// Short message, source-bounded.
		return truncate(x.Error())

	case abci.Error:
		// All abci.Error implementers have AssertABCIError() and a
		// short Error() string. std error types (InsufficientCoinsError,
		// InternalError, etc.) return hardcoded constants. The only
		// risk is abci.StringError which wraps a raw string —
		// truncate() handles that.
		return truncate(x.Error())

	case cmnerrors.Error:
		// tm2/pkg/errors *cmnError. Avoid .Error() (which calls
		// Sprintf("%v", err) → Format → walks msgtraces + stacktrace).
		// Peek at .Data():
		//   - FmtError: use .Format() — returns raw format string, no
		//     Sprintf invocation, no expansion.
		//   - error:    recurse via errors.Unwrap.
		//   - other:    fall through to <error: %T>.
		if fe, ok := x.Data().(cmnerrors.FmtError); ok {
			return truncate(fe.Format())
		}
		if u := errors.Unwrap(x); u != nil {
			return boundedString(u, depth+1)
		}
		return fmt.Sprintf("<error: %T>", x)

	case error:
		if u := errors.Unwrap(x); u != nil {
			return boundedString(u, depth+1)
		}
		return fmt.Sprintf("<error: %T>", x)

	default:
		return fmt.Sprintf("<%T>", v)
	}
}

// truncate caps s at maxBoundedBytes; appends "..." on truncation.
func truncate(s string) string {
	if len(s) <= maxBoundedBytes {
		return s
	}
	return s[:maxBoundedBytes] + "..."
}
