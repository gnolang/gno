package executor

import (
	"errors"

	"gno.land/p/demo/context"
)

type propContextKey string

func (k propContextKey) String() string { return string(k) }

const (
	statusContextKey = propContextKey("govdao-prop-status")
	approvedStatus   = "approved"
)

var errNotApproved = errors.New("not approved by govdao")

// CtxExecutor is an implementation of the dao.Executor interface,
// based on the given context.
// It utilizes the given context to assert the validity of the govdao call
type CtxExecutor struct {
	callbackCtx func(ctx context.Context) error // the callback ctx fn, if any
}

// NewCtxExecutor creates a new executor with the provided callback function.
func NewCtxExecutor(callback func(ctx context.Context) error) *CtxExecutor {
	return &CtxExecutor{
		callbackCtx: callback,
	}
}

// Execute runs the executor's callback function
func (exec *CtxExecutor) Execute() error {
	// Verify the executor is r/gov/dao
	assertCalledByGovdao()

	// Create the context
	ctx := context.WithValue(
		context.Empty(),
		statusContextKey,
		approvedStatus,
	)

	return exec.callbackCtx(ctx)
}

// IsApprovedByGovdaoContext asserts that the govdao approved the context
func IsApprovedByGovdaoContext(ctx context.Context) bool {
	v := ctx.Value(statusContextKey)
	if v == nil {
		return false
	}

	vs, ok := v.(string)

	return ok && vs == approvedStatus
}

// AssertContextApprovedByGovDAO asserts the given context
// was approved by GOVDAO
func AssertContextApprovedByGovDAO(ctx context.Context) {
	if IsApprovedByGovdaoContext(ctx) {
		return
	}

	panic(errNotApproved)
}
