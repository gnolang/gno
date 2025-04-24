package std

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ExecContext struct {
	ChainID         string
	ChainDomain     string
	Height          int64
	Timestamp       int64 // seconds
	TimestampNano   int64 // nanoseconds, only used for testing.
	OriginCaller    crypto.Bech32Address
	OriginPkgAddr   crypto.Bech32Address
	OriginSend      std.Coins
	OriginSendSpent *std.Coins // mutable
	Banker          BankerInterface
	Params          ParamsInterface
	EventLogger     *sdk.EventLogger
}

// GetContext returns the execution context.
// This is used to allow extending the exec context using interfaces,
// for instance when testing.
func (e ExecContext) GetExecContext() ExecContext {
	return e
}

var _ ExecContexter = ExecContext{}

// ExecContexter is a type capable of returning the parent [ExecContext]. When
// using these standard libraries, m.Context should always implement this
// interface. This can be obtained by embedding [ExecContext].
type ExecContexter interface {
	GetExecContext() ExecContext
}

// NOTE: In order to make this work by simply embedding ExecContext in another
// context (like TestExecContext), the method needs to be something other than
// the field name.

// GetContext returns the context from the Gno machine.
func GetContext(m *gno.Machine) ExecContext {
	return m.Context.(ExecContexter).GetExecContext()
}

// TestContext provides a testing environment for Gno contracts.
// It embeds ExecContext and adds testing-specific functionality.
type TestContext struct {
	*ExecContext
	EventLogger *sdk.EventLogger
}

// NewTestContext creates a new TestContext with initialized EventLogger.
func NewTestContext() *TestContext {
	return &TestContext{
		ExecContext: &ExecContext{
			EventLogger: sdk.NewEventLogger(),
		},
	}
}

// EventVerifier defines the interface for event verification.
// It provides methods to verify events against expected values.
type EventVerifier interface {
	Verify(events []GnoEvent) bool
	Error() string
}

// EventVerifierImpl implements EventVerifier interface.
// It holds the expected event type, attributes, and verification settings.
type EventVerifierImpl struct {
	expectedType  string
	expectedAttrs map[string]string
	err           string
	eventIndex    int
	partialMatch  bool
}

// WithEventIndex sets the index of the event to verify.
// Negative index (-1) means the last event.
func (ev *EventVerifierImpl) WithEventIndex(idx int) EventVerifier {
	ev.eventIndex = idx
	return ev
}

// ExepectEventType creates a new EventVerifier for the given event type.
func ExepectEventType(evtType string) EventVerifier {
	return &EventVerifierImpl{
		expectedType:  evtType,
		expectedAttrs: make(map[string]string),
	}
}

// WithAttribute adds an expected attribute to the verifier.
func (ev *EventVerifierImpl) WithAttribute(k, v string) EventVerifier {
	ev.expectedAttrs[k] = v
	return ev
}

// WithPartialMatch enables partial matching of attributes.
// When enabled, only specified attributes are verified, ignoring others.
func (ev *EventVerifierImpl) WithPartialMatch() EventVerifier {
	ev.partialMatch = true
	return ev
}

// Verify checks if the given events match the expected values.
// Returns true if verification passes, false otherwise.
func (ev *EventVerifierImpl) Verify(events []GnoEvent) bool {
	if len(events) == 0 {
		ev.err = "no events emitted"
		return false
	}

	var targetEvent GnoEvent
	if ev.eventIndex < 0 {
		targetEvent = events[len(events)-1]
	} else if ev.eventIndex >= len(events) {
		ev.err = fmt.Sprintf("event index %d out of range", ev.eventIndex)
		return false
	} else {
		targetEvent = events[ev.eventIndex]
	}

	if targetEvent.Type != ev.expectedType {
		ev.err = fmt.Sprintf("expected event type %s, got %s", ev.expectedType, targetEvent.Type)
		return false
	}

	attrs := make(map[string]string)
	for _, attr := range targetEvent.Attributes {
		attrs[attr.Key] = attr.Value
	}

	if !ev.partialMatch && len(attrs) != len(ev.expectedAttrs) {
		ev.err = fmt.Sprintf("attribute count mismatch: expected %d, got %d", len(ev.expectedAttrs), len(attrs))
		return false
	}

	for k, v := range ev.expectedAttrs {
		if actual, exists := attrs[k]; !exists || actual != v {
			ev.err = fmt.Sprintf("expected attribute %s=%s, got %s=%s", k, v, k, actual)
			return false
		}
	}

	return true
}

func (ev *EventVerifierImpl) Error() string {
	return ev.err
}

// WithTestContext executes a function with a test context.
// The original context is restored after the function returns.
func WithTestContext(m *gno.Machine, f func(ctx *TestContext)) {
	ctx := NewTestContext()
	oldCtx := m.Context
	m.Context = ctx
	defer func() {
		m.Context = oldCtx
	}()
	f(ctx)
}
