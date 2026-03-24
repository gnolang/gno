// Package security provides a read-only dashboard displaying security-critical
// variables across production realms, their current state, and risk explanations.
//
// All checks live in a single registry pre-seeded with known production flags.
// GovDAO proposals can add, update, or remove any check — including the
// pre-seeded ones — so the dashboard stays accurate if a realm upgrades its API.
package security

import (
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gov/dao"
	"gno.land/r/sys/cla"
	names "gno.land/r/sys/names"
	validators "gno.land/r/sys/validators/v2"
)

// Check represents a single security check in the registry.
// The check ID encodes the realm: "realm/path:flag_name".
type Check struct {
	Label    string        // human-readable flag name
	ValueFn  func() string // closure evaluated at render time
	Expected string        // expected value for OK status
	Risk     string        // risk description if the value is wrong
}

// checks is the single registry. Keys are human-readable IDs.
var checks avl.Tree // string -> *Check

// Pre-seed the registry with known production checks.
func init() {
	checks.Set("r/gnoland/boards2/v1:realm_locked", &Check{
		Label: "Realm Locked",
		ValueFn: func() string {
			return strconv.FormatBool(boards2.IsRealmLocked())
		},
		Expected: "false",
		Risk:     "All boards frozen",
	})

	checks.Set("r/gnoland/boards2/v1:members_locked", &Check{
		Label: "Members Locked",
		ValueFn: func() string {
			return strconv.FormatBool(boards2.AreRealmMembersLocked())
		},
		Expected: "false",
		Risk:     "No membership changes possible",
	})

	checks.Set("r/sys/validators/v2:count", &Check{
		Label: "Validator Count",
		ValueFn: func() string {
			return strconv.Itoa(len(validators.GetValidators()))
		},
		Expected: "> 0",
		Risk:     "No block production",
	})

	// TODO: update when #5080 is merged
	checks.Set("r/sys/names:enabled", &Check{
		Label: "Namespace Enforcement",
		ValueFn: func() string {
			return strconv.FormatBool(names.IsEnabled())
		},
		Expected: "true",
		Risk:     "Namespace hijacking",
	})

	checks.Set("r/sys/cla:enforcement", &Check{
		Label: "CLA Enforcement",
		ValueFn: func() string {
			// HasValidSignature returns true when requiredHash == "" (disabled).
			// Passing the zero address: if true, enforcement is off.
			return strconv.FormatBool(!cla.HasValidSignature(address("")))
		},
		Expected: "true",
		Risk:     "Unsigned code can be committed",
	})
}

// ---------------------------------------------------------------------------
// GovDAO proposals — single API for all checks
// ---------------------------------------------------------------------------

// NewAddCheckProposalRequest creates a GovDAO proposal to register a new check.
// The id must be formatted as "realm/path:flag_name" (e.g. "r/example/myrealm:my_flag").
// The valueFn closure is called at render time to fetch the current value.
func NewAddCheckProposalRequest(description, id, label, expected, risk string, valueFn func() string) dao.ProposalRequest {
	validateCheckArgs(id, label, valueFn)
	if checks.Has(id) {
		panic("check already exists: " + id)
	}

	cb := func(cur realm) error {
		setCheck(id, label, expected, risk, valueFn)
		return nil
	}

	return dao.NewProposalRequest(
		"Add Security Check: "+label,
		checkDesc(description, "Add", id, label, expected, risk),
		dao.NewSimpleExecutor(cb, "Add check ["+id+"] to the security dashboard"),
	)
}

// NewUpdateCheckProposalRequest creates a GovDAO proposal to update an
// existing check. This allows adapting to realm API changes without
// redeploying the security dashboard.
func NewUpdateCheckProposalRequest(description, id, label, expected, risk string, valueFn func() string) dao.ProposalRequest {
	validateCheckArgs(id, label, valueFn)
	if !checks.Has(id) {
		panic("check not found: " + id)
	}

	cb := func(cur realm) error {
		setCheck(id, label, expected, risk, valueFn)
		return nil
	}

	return dao.NewProposalRequest(
		"Update Security Check: "+label,
		checkDesc(description, "Update", id, label, expected, risk),
		dao.NewSimpleExecutor(cb, "Update check ["+id+"] on the security dashboard"),
	)
}

// NewRemoveCheckProposalRequest creates a GovDAO proposal to remove a check.
func NewRemoveCheckProposalRequest(description, id string) dao.ProposalRequest {
	if !checks.Has(id) {
		panic("check not found: " + id)
	}

	cb := func(cur realm) error {
		checks.Remove(id)
		return nil
	}

	return dao.NewProposalRequest(
		"Remove Security Check: "+id,
		checkDesc(description, "Remove", id, "", "", ""),
		dao.NewSimpleExecutor(cb, "Remove check ["+id+"] from the security dashboard"),
	)
}

func validateCheckArgs(id, label string, valueFn func() string) {
	if id == "" {
		panic("id must not be empty")
	}
	if !strings.Contains(id, ":") {
		panic("id must be formatted as realm/path:flag_name")
	}
	if label == "" {
		panic("label must not be empty")
	}
	if valueFn == nil {
		panic("valueFn must not be nil")
	}
}

func setCheck(id, label, expected, risk string, valueFn func() string) {
	checks.Set(id, &Check{
		Label:    label,
		ValueFn:  valueFn,
		Expected: expected,
		Risk:     risk,
	})
}

func checkDesc(description, action, id, label, expected, risk string) string {
	out := ufmt.Sprintf("%s security dashboard check.\n\n", action)
	if description != "" {
		out += description + "\n\n"
	}
	out += "- **ID:** " + id + "\n"
	if label != "" {
		out += "- **Label:** " + label + "\n"
	}
	if expected != "" {
		out += "- **Expected:** " + expected + "\n"
	}
	if risk != "" {
		out += "- **Risk:** " + risk + "\n"
	}
	return out
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

// Render produces the security dashboard as markdown.
// Subpages: "help" — how to manage checks via GovDAO.
func Render(path string) string {
	if path == "help" {
		return renderHelp()
	}

	out := "# Security Dashboard\n\n"
	out += "Read-only view of security-critical flags across production realms.\n\n"

	out += renderChecks()

	out += "Checks can be added, updated, or removed via GovDAO proposals. "
	out += "See [How to manage checks](/r/sys/security:help) for instructions.\n\n"
	out += "---\n\n"

	out += renderNotQueryable()

	return out
}

func renderChecks() string {
	out := "## Checks\n\n"

	if checks.Size() == 0 {
		out += "(none registered)\n\n"
		return out
	}

	out += "| ID | Flag | Current | Expected | Risk |\n"
	out += "|----|------|---------|----------|------|\n"

	checks.Iterate("", "", func(key string, value any) bool {
		c := value.(*Check)
		current := safeCall(c.ValueFn)
		if !matchExpected(current, c.Expected) {
			current = "**" + current + "**"
		}
		realmPath := realmFromID(key)
		idLink := "[" + key + "](/" + realmPath + ")"
		out += "| " + idLink +
			" | " + c.Label +
			" | " + current +
			" | " + c.Expected +
			" | " + c.Risk +
			" |\n"
		return false
	})

	out += "\n"
	return out
}

func renderNotQueryable() string {
	out := "## Not Queryable On-Chain\n\n"
	out += "These flags are controlled at the VM/param level and cannot be read from Gno code.\n\n"
	out += "| Flag | How to Check | Risk |\n"
	out += "|------|-------------|------|\n"
	out += "| CLA Enforcement (VM) | `gnokey query params/vm:p:syscla_pkgpath` | CLA checks skipped if empty |\n"
	out += "| Restricted Denoms | `gnokey query params/bank:p:restricted_denoms` | Token transfers unrestricted if empty |\n"
	out += "| Unrestricted Addresses | `gnokey query params/auth:p:unrestricted_addrs` | Whitelisted addrs bypass transfer lock |\n"
	out += "| Default Deposit | `gnokey query params/vm:p:default_deposit` | Cheap spam deployments if too low |\n"
	out += "| Storage Price | `gnokey query params/vm:p:storage_price` | State bloat if too low |\n"
	out += "| Fee Collector | `gnokey query params/auth:p:fee_collector` | Fee diversion if changed |\n"
	out += "| Storage Fee Collector | `gnokey query params/vm:p:storage_fee_collector` | Storage fee diversion if changed |\n"
	out += "\n"
	return out
}

// ---------------------------------------------------------------------------
// Help subpage
// ---------------------------------------------------------------------------

func renderHelp() string {
	out := "# Managing Security Checks\n\n"
	out += "[Back to Dashboard](/r/sys/security)\n\n"
	out += "All checks (including pre-seeded ones) can be added, updated, or removed via GovDAO proposals.\n\n"
	out += "The check **id** encodes both the realm and the flag name, formatted as `realm/path:flag_name`.\n"
	out += "The realm portion is extracted automatically for display in the dashboard.\n\n"

	out += "## Add a new check\n\n"
	out += "```gno\n"
	out += "package main\n\n"
	out += "import (\n"
	out += "\t\"gno.land/r/gov/dao\"\n"
	out += "\t\"gno.land/r/sys/security\"\n"
	out += "\tmyrealm \"gno.land/r/example/myrealm\"\n"
	out += ")\n\n"
	out += "func main() {\n"
	out += "\treq := security.NewAddCheckProposalRequest(\n"
	out += "\t\t\"Track myrealm flag\",      // description for voters\n"
	out += "\t\t\"r/example/myrealm:flag\",  // id (realm/path:flag_name)\n"
	out += "\t\t\"My Flag\",                 // label\n"
	out += "\t\t\"true\",                    // expected value\n"
	out += "\t\t\"Bad things if wrong\",     // risk\n"
	out += "\t\tfunc() string { return myrealm.GetFlag() },\n"
	out += "\t)\n"
	out += "\tdao.MustCreateProposal(cross, req)\n"
	out += "}\n"
	out += "```\n\n"

	out += "## Update an existing check\n\n"
	out += "Use the same pattern with `NewUpdateCheckProposalRequest` and the existing check ID.\n"
	out += "This is useful when you need to change the label, expected value, risk, or valueFn.\n"
	out += "If the realm itself changes (e.g. v1 to v2), remove the old check and add a new one.\n\n"
	out += "```gno\n"
	out += "package main\n\n"
	out += "import (\n"
	out += "\t\"gno.land/r/gov/dao\"\n"
	out += "\t\"gno.land/r/sys/security\"\n"
	out += "\tmyrealm \"gno.land/r/example/myrealm\"\n"
	out += ")\n\n"
	out += "func main() {\n"
	out += "\treq := security.NewUpdateCheckProposalRequest(\n"
	out += "\t\t\"Update expected value\",   // description for voters\n"
	out += "\t\t\"r/example/myrealm:flag\",  // existing id\n"
	out += "\t\t\"My Flag (updated)\",       // updated label\n"
	out += "\t\t\"false\",                   // new expected\n"
	out += "\t\t\"Updated risk description\",// risk\n"
	out += "\t\tfunc() string { return myrealm.GetFlag() },\n"
	out += "\t)\n"
	out += "\tdao.MustCreateProposal(cross, req)\n"
	out += "}\n"
	out += "```\n\n"

	out += "## Remove a check\n\n"
	out += "```gno\n"
	out += "package main\n\n"
	out += "import (\n"
	out += "\t\"gno.land/r/gov/dao\"\n"
	out += "\t\"gno.land/r/sys/security\"\n"
	out += ")\n\n"
	out += "func main() {\n"
	out += "\treq := security.NewRemoveCheckProposalRequest(\n"
	out += "\t\t\"Realm deprecated\",        // description for voters\n"
	out += "\t\t\"r/example/myrealm:flag\",  // id to remove\n"
	out += "\t)\n"
	out += "\tdao.MustCreateProposal(cross, req)\n"
	out += "}\n"
	out += "```\n\n"

	out += "## Fields\n\n"
	out += "| Field | Description |\n"
	out += "|-------|-------------|\n"
	out += "| **description** | Context for GovDAO voters explaining why this change is needed |\n"
	out += "| **id** | `realm/path:flag_name` — encodes both the realm and the flag |\n"
	out += "| **label** | Short name shown in the dashboard |\n"
	out += "| **expected** | Value that means OK; mismatches are shown in **bold** |\n"
	out += "| **risk** | One-line explanation of the risk |\n"
	out += "| **valueFn** | `func() string` closure called at render time |\n"
	return out
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// matchExpected compares the current value to the expected string.
// Supports "> 0" as a special threshold expression.
func matchExpected(current, expected string) bool {
	if expected == "> 0" {
		n, err := strconv.Atoi(current)
		if err != nil {
			return false
		}
		return n > 0
	}
	return current == expected
}

// realmFromID extracts the realm path from a check ID formatted as
// "realm/path:flag_name". It returns everything before the last ':'.
func realmFromID(id string) string {
	i := strings.LastIndex(id, ":")
	if i < 0 {
		return id
	}
	return id[:i]
}

// safeCall invokes fn and recovers from panics, returning an error string.
func safeCall(fn func() string) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = ufmt.Sprintf("ERROR: %v", r)
		}
	}()
	return fn()
}
