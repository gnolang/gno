package security

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/uassert/v0"
)

// resetChecks replaces the checks tree with a fresh one and returns
// a function to restore the original. Call: defer resetChecks(t)()
func resetChecks(t *testing.T) func() {
	t.Helper()
	saved := checks
	checks = *new(avl.Tree)
	return func() { checks = saved }
}

func TestRender_PreSeededChecks(t *testing.T) {
	out := Render("")

	// Header
	uassert.True(t, strings.Contains(out, "# Security Dashboard"))

	// IDs link to their realm
	uassert.True(t, strings.Contains(out, "[r/gnoland/boards2/v1:realm_locked](/r/gnoland/boards2/v1)"))
	uassert.True(t, strings.Contains(out, "[r/sys/validators/v2:count](/r/sys/validators/v2)"))
	uassert.True(t, strings.Contains(out, "[r/sys/cla:enforcement](/r/sys/cla)"))

	// Check labels
	uassert.True(t, strings.Contains(out, "Realm Locked"))
	uassert.True(t, strings.Contains(out, "Members Locked"))
	uassert.True(t, strings.Contains(out, "Validator Count"))
	uassert.True(t, strings.Contains(out, "CLA Enforcement"))

	// Not-queryable section
	uassert.True(t, strings.Contains(out, "Not Queryable On-Chain"))
	uassert.True(t, strings.Contains(out, "syscla_pkgpath"))
	uassert.True(t, strings.Contains(out, "restricted_denoms"))
	uassert.True(t, strings.Contains(out, "unrestricted_addrs"))
	uassert.True(t, strings.Contains(out, "default_deposit"))
	uassert.True(t, strings.Contains(out, "storage_price"))
	uassert.True(t, strings.Contains(out, "fee_collector"))
	uassert.True(t, strings.Contains(out, "storage_fee_collector"))
}

func TestRender_HelpLink(t *testing.T) {
	out := Render("")
	uassert.True(t, strings.Contains(out, "[How to manage checks](/r/sys/security:help)"))
}

func TestAddCheck_WarnOnMismatch(t *testing.T) {
	defer resetChecks(t)()

	setCheck("r/test:mismatch", "test label", "expected", "some risk", func() string {
		return "actual"
	})

	out := Render("")
	uassert.True(t, strings.Contains(out, "test label"))
	// Current value differs from expected — should be bold
	uassert.True(t, strings.Contains(out, "**actual**"))
	uassert.True(t, strings.Contains(out, "expected"))
}

func TestAddCheck_OKOnMatch(t *testing.T) {
	defer resetChecks(t)()

	setCheck("r/test:match", "ok check", "42", "bad if wrong", func() string {
		return "42"
	})

	out := Render("")
	uassert.True(t, strings.Contains(out, "ok check"))
	// Current matches expected — should NOT be bold
	uassert.True(t, strings.Contains(out, "| 42 |"))
	uassert.False(t, strings.Contains(out, "**42**"))
}

func TestAddCheck_Gt0Expected(t *testing.T) {
	defer resetChecks(t)()

	setCheck("r/test:gt0_ok", "count", "> 0", "none", func() string {
		return "5"
	})
	out := Render("")
	// 5 > 0, matches — not bold
	uassert.True(t, strings.Contains(out, "| 5 |"))
	uassert.False(t, strings.Contains(out, "**5**"))

	checks = *new(avl.Tree)
	setCheck("r/test:gt0_warn", "count", "> 0", "none", func() string {
		return "0"
	})
	out = Render("")
	// 0 is not > 0, mismatch — bold
	uassert.True(t, strings.Contains(out, "**0**"))
}

func TestRemoveCheck(t *testing.T) {
	defer resetChecks(t)()

	setCheck("r/test:remove", "to remove", "val", "risk", func() string {
		return "val"
	})
	uassert.Equal(t, 1, checks.Size())

	checks.Remove("r/test:remove")
	uassert.Equal(t, 0, checks.Size())

	out := Render("")
	uassert.True(t, strings.Contains(out, "(none registered)"))
}

func TestUpdateCheck(t *testing.T) {
	defer resetChecks(t)()

	setCheck("r/test:update", "old label", "old", "old risk", func() string {
		return "old"
	})

	// Update it (same ID, new values)
	setCheck("r/test:update", "new label", "new", "new risk", func() string {
		return "new"
	})

	uassert.Equal(t, 1, checks.Size())
	out := Render("")
	uassert.True(t, strings.Contains(out, "new label"))
	uassert.True(t, strings.Contains(out, "[r/test:update](/r/test)"))
	// Current matches expected — not bold
	uassert.True(t, strings.Contains(out, "| new |"))
	uassert.False(t, strings.Contains(out, "**new**"))
}

func TestSafeCall_Normal(t *testing.T) {
	result := safeCall(func() string { return "hello" })
	uassert.Equal(t, "hello", result)
}

func TestSafeCall_Panic(t *testing.T) {
	result := safeCall(func() string { panic("boom") })
	uassert.True(t, strings.Contains(result, "ERROR:"))
	uassert.True(t, strings.Contains(result, "boom"))
}

func TestNewAddCheckProposalRequest_Validation(t *testing.T) {
	uassert.PanicsWithMessage(t, "id must not be empty", func() {
		NewAddCheckProposalRequest("desc", "", "label", "exp", "risk", func() string { return "" })
	})
	uassert.PanicsWithMessage(t, "id must be formatted as realm/path:flag_name", func() {
		NewAddCheckProposalRequest("desc", "nocolon", "label", "exp", "risk", func() string { return "" })
	})
	uassert.PanicsWithMessage(t, "label must not be empty", func() {
		NewAddCheckProposalRequest("desc", "r/x:y", "", "exp", "risk", func() string { return "" })
	})
	uassert.PanicsWithMessage(t, "valueFn must not be nil", func() {
		NewAddCheckProposalRequest("desc", "r/x:y", "label", "exp", "risk", nil)
	})
	// Can't add a check that already exists (pre-seeded)
	uassert.PanicsWithMessage(t, "check already exists: r/gnoland/boards2/v1:realm_locked", func() {
		NewAddCheckProposalRequest("desc", "r/gnoland/boards2/v1:realm_locked", "dup", "exp", "risk", func() string { return "" })
	})
}

func TestNewUpdateCheckProposalRequest_Validation(t *testing.T) {
	uassert.PanicsWithMessage(t, "check not found: r/nonexistent:flag", func() {
		NewUpdateCheckProposalRequest("desc", "r/nonexistent:flag", "label", "exp", "risk", func() string { return "" })
	})
	// Pre-seeded checks can be updated
	uassert.NotPanics(t, func() {
		NewUpdateCheckProposalRequest("desc", "r/gnoland/boards2/v1:realm_locked", "Realm Locked", "false", "Frozen", func() string { return "false" })
	})
}

func TestNewRemoveCheckProposalRequest_NotFound(t *testing.T) {
	uassert.PanicsWithMessage(t, "check not found: r/nonexistent:flag", func() {
		NewRemoveCheckProposalRequest("desc", "r/nonexistent:flag")
	})
}

func TestRender_HelpSubpage(t *testing.T) {
	out := Render("help")
	uassert.True(t, strings.Contains(out, "# Managing Security Checks"))
	uassert.True(t, strings.Contains(out, "[Back to Dashboard](/r/sys/security)"))
	uassert.True(t, strings.Contains(out, "NewAddCheckProposalRequest"))
	uassert.True(t, strings.Contains(out, "NewUpdateCheckProposalRequest"))
	uassert.True(t, strings.Contains(out, "NewRemoveCheckProposalRequest"))
	uassert.True(t, strings.Contains(out, "dao.MustCreateProposal"))
	uassert.True(t, strings.Contains(out, "realm/path:flag_name"))
}

func TestRealmFromID(t *testing.T) {
	uassert.Equal(t, "r/gnoland/boards2/v1", realmFromID("r/gnoland/boards2/v1:realm_locked"))
	uassert.Equal(t, "r/sys/cla", realmFromID("r/sys/cla:enforcement"))
	uassert.Equal(t, "nocolon", realmFromID("nocolon"))
}

func TestMatchExpected(t *testing.T) {
	uassert.True(t, matchExpected("true", "true"))
	uassert.False(t, matchExpected("false", "true"))
	uassert.True(t, matchExpected("5", "> 0"))
	uassert.True(t, matchExpected("1", "> 0"))
	uassert.False(t, matchExpected("0", "> 0"))
	uassert.False(t, matchExpected("-1", "> 0"))
	uassert.False(t, matchExpected("abc", "> 0"))
}
