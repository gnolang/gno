package proxy

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

// --- a minimal mock that satisfies MyApp ---

type mockApp struct {
	v     string
	field string
}

func (m *mockApp) Render() string  { return "mock(" + m.field + ")" }
func (m *mockApp) Do(arg string)   { m.field = arg }
func (m *mockApp) Version() string { return m.v }

// --- tests ---

func TestRender_NoImplementation(t *testing.T) {
	// Ensure a clean slate
	app = nil
	got := Render("")
	uassert.Equal(t, got, "No implementation currently available.")
}

func TestSetImplementation_NilPanics(t *testing.T) {
	r := revive(func() {
		SetImplementation(cross, nil)
	})
	if r == nil {
		t.Fatal("expected panic on nil implementation")
	}
}

func TestSetImplementation_Unauthorized(t *testing.T) {
	stranger := testutils.TestAddress("stranger")
	testing.SetOriginCaller(stranger)

	r := revive(func() {
		SetImplementation(cross, &mockApp{v: "vX"})
	})
	uassert.NotNil(t, r, "expected unauthorized panic")
}

func TestSetImplementation_Authorized_AndRouting(t *testing.T) {
	// The owner in your proxy is hardcoded; mirror it here:
	ownerAddr := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	testing.SetOriginCaller(ownerAddr)

	m := &mockApp{v: "v1", field: "hello"}
	SetImplementation(cross, m)

	uassert.True(t, IsImplSet(), "app should be set")

	// Render routes to impl
	uassert.Equal(t, Render(""), "mock(hello)")

	// Do routes to impl and mutates impl state
	Do(cross, "world")
	uassert.Equal(t, Render(""), "mock(world)")

	// Version routes to impl
	uassert.Equal(t, Version(), "v1")
}
