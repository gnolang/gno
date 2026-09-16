package authz

import (
	"chain"
	"errors"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestNewWithCurrent(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	auth := NewWithMembers(cur.Address())

	// Check that the current authority is a MemberAuthority
	memberAuth, ok := auth.Authority().(*MemberAuthority)
	uassert.True(t, ok, "expected MemberAuthority")

	// Check that the caller is a member
	uassert.True(t, memberAuth.Has(alice), "caller should be a member")

	// Check string representation
	uassert.True(t, strings.Contains(auth.String(), alice.String()))
}

func TestNewWithAuthority(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	memberAuth := NewMemberAuthority(alice)

	auth := NewWithAuthority(memberAuth)

	// Check that the current authority is the one we provided
	uassert.True(t, auth.Authority() == memberAuth, "expected provided authority")
}

func TestAuthorizerAuthorize(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	auth := NewWithMembers(cur.Address())

	// Test successful action with args
	executed := false
	args := []any{"test_arg", 123}
	err := auth.DoByCurrent(0, cur, "test_action", func() error {
		executed = true
		return nil
	}, args...)

	uassert.True(t, err == nil, "expected no error")
	uassert.True(t, executed, "action should have been executed")

	// Test unauthorized action with args
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("bob")))

	executed = false
	err = auth.DoByCurrent(0, cur, "test_action", func() error {
		executed = true
		return nil
	}, "unauthorized_arg")

	uassert.True(t, err != nil, "expected error")
	uassert.False(t, executed, "action should not have been executed")

	// Test action returning error
	testing.SetRealm(testing.NewUserRealm(alice))
	expectedErr := errors.New("test error")

	err = auth.DoByCurrent(0, cur, "test_action", func() error {
		return expectedErr
	})

	uassert.True(t, err == expectedErr, "expected specific error")
}

func TestAuthorizerTransfer(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	auth := NewWithMembers(cur.Address())

	// Test transfer to new member authority
	bob := testutils.TestAddress("bob")
	newAuth := NewMemberAuthority(bob)

	var err error
	func(cur realm) { err = auth.Transfer(0, cur, newAuth) }(cross(cur))
	uassert.True(t, err == nil, "expected no error")
	uassert.True(t, auth.Authority() == newAuth, "expected new authority")

	// Test unauthorized transfer: principal is not a member of newAuth.
	carol := testutils.TestAddress("carol")
	testing.SetRealm(testing.NewUserRealm(carol))

	func(cur realm) { err = auth.Transfer(0, cur, NewMemberAuthority(alice)) }(cross(cur))
	uassert.True(t, err != nil, "expected error")

	// Test transfer to contract authority — bob is the current authority.
	testing.SetRealm(testing.NewUserRealm(bob))
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		return action()
	})

	func(cur realm) { err = auth.Transfer(0, cur, contractAuth) }(cross(cur))
	uassert.True(t, err == nil, "expected no error")
	uassert.True(t, auth.Authority() == contractAuth, "expected contract authority")
}

func TestAuthorizerTransferChain(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	// Create a chain of transfers
	auth := NewWithMembers(cur.Address())

	// First transfer to a new member authority
	bob := testutils.TestAddress("bob")
	memberAuth := NewMemberAuthority(bob)

	var err error
	func(cur realm) { err = auth.Transfer(0, cur, memberAuth) }(cross(cur))
	uassert.True(t, err == nil, "unexpected error in first transfer")

	// Then transfer to a contract authority — bob is now the authority.
	testing.SetRealm(testing.NewUserRealm(bob))
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		return action()
	})
	func(cur realm) { err = auth.Transfer(0, cur, contractAuth) }(cross(cur))
	uassert.True(t, err == nil, "unexpected error in second transfer")

	// Finally transfer to an auto-accept authority. The authority is now
	// the contract authority, whose default proposer is the contract
	// itself, so the transfer's caller —
	// cur.Previous() — must be the contract. Set the outer realm to the
	// contract, then cross into a closure so cur.Previous() resolves to
	// it.
	autoAuth := NewAutoAcceptAuthority()
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	func(cur realm) { err = auth.Transfer(0, cur, autoAuth) }(cross(cur))
	uassert.True(t, err == nil, "unexpected error in final transfer")
	uassert.True(t, auth.Authority() == autoAuth, "expected auto-accept authority")
}

func TestAuthorizerTransferUnauthorizedRejected(cur realm, t *testing.T) {
	// Regression for the address-parameter forgery: previously Transfer
	// took a caller-supplied `caller address` that an attacker realm could
	// set to the real owner. After the IsCurrent + rlm.Previous() fix, the
	// principal is the captured cur's previous and cannot be forged.
	admin := testutils.TestAddress("admin")
	attacker := testutils.TestAddress("attacker")

	testing.SetRealm(testing.NewUserRealm(admin))
	auth := NewWithMembers(cur.Address()) // admin is the initial authority

	initialAuth, ok := auth.Authority().(*MemberAuthority)
	uassert.True(t, ok)
	uassert.True(t, initialAuth.Has(admin))
	uassert.False(t, initialAuth.Has(attacker))

	// Attacker context: cur.Previous() inside the closure will be attacker.
	testing.SetRealm(testing.NewUserRealm(attacker))
	attackerAuth := NewMemberAuthority(attacker)
	var err error
	func(cur realm) { err = auth.Transfer(0, cur, attackerAuth) }(cross(cur))

	uassert.True(t, err != nil, "attacker transfer must be rejected")

	// Authority unchanged: still the initial admin-only MemberAuthority.
	finalAuth, ok := auth.Authority().(*MemberAuthority)
	uassert.True(t, ok)
	uassert.True(t, finalAuth == initialAuth, "authority must not have changed")
	uassert.True(t, finalAuth.Has(admin))
	uassert.False(t, finalAuth.Has(attacker))
}

func TestAuthorizerWithDroppedAuthority(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	auth := NewWithMembers(cur.Address())

	// Transfer to dropped authority
	var err error
	func(cur realm) { err = auth.Transfer(0, cur, NewDroppedAuthority()) }(cross(cur))
	uassert.True(t, err == nil, "expected no error")

	// Try to execute action
	err = auth.DoByCurrent(0, cur, "test_action", func() error {
		return nil
	})
	uassert.True(t, err != nil, "expected error from dropped authority")

	// Try to transfer again
	func(cur realm) { err = auth.Transfer(0, cur, NewMemberAuthority(alice)) }(cross(cur))
	uassert.True(t, err != nil, "expected error when transferring from dropped authority")
}

func TestContractAuthorityHandlerExecutionOnce(cur realm, t *testing.T) {
	attempts := 0
	executed := 0

	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		// Try to execute the action twice in the same handler
		if err := action(); err != nil {
			return err
		}
		attempts++

		// Second execution should fail
		if err := action(); err != nil {
			return err
		}
		attempts++
		return nil
	})

	// Set caller to contract address
	codeRealm := testing.NewCodeRealm("gno.land/r/test")
	testing.SetRealm(codeRealm)
	code := codeRealm.Address()

	testArgs := []any{"proposal_id", 42, "metadata", map[string]string{"key": "value"}}
	err := contractAuth.Authorize(code, "test_action", func() error {
		executed++
		return nil
	}, testArgs...)

	uassert.True(t, err == nil, "handler execution should succeed")
	uassert.True(t, attempts == 2, "handler should have attempted execution twice")
	uassert.True(t, executed == 1, "handler should have executed once")
}

func TestContractAuthorityExecutionTwice(cur realm, t *testing.T) {
	executed := 0

	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		return action()
	})

	// Set caller to contract address
	codeRealm := testing.NewCodeRealm("gno.land/r/test")
	testing.SetRealm(codeRealm)
	code := codeRealm.Address()
	testArgs := []any{"proposal_id", 42, "metadata", map[string]string{"key": "value"}}

	err := contractAuth.Authorize(code, "test_action", func() error {
		executed++
		return nil
	}, testArgs...)

	uassert.True(t, err == nil, "handler execution should succeed")
	uassert.True(t, executed == 1, "handler should have executed once")

	// A new action, even with the same title, should be executed
	err = contractAuth.Authorize(code, "test_action", func() error {
		executed++
		return nil
	}, testArgs...)

	uassert.True(t, err == nil, "handler execution should succeed")
	uassert.True(t, executed == 2, "handler should have executed twice")
}

func TestContractAuthorityWithProposer(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	memberAuth := NewMemberAuthority(alice)

	handlerCalled := false
	actionExecuted := false

	contractAuth := NewRestrictedContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		handlerCalled = true
		// Set caller to contract address before executing action
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
		return action()
	}, memberAuth)

	// Test authorized member
	testArgs := []any{"proposal_metadata", "test value"}
	err := contractAuth.Authorize(alice, "test_action", func() error {
		actionExecuted = true
		return nil
	}, testArgs...)

	uassert.True(t, err == nil, "authorized member should be able to propose")
	uassert.True(t, handlerCalled, "contract handler should be called")
	uassert.True(t, actionExecuted, "action should be executed")

	// Reset flags for unauthorized test
	handlerCalled = false
	actionExecuted = false

	// Test unauthorized proposer
	bob := testutils.TestAddress("bob")
	err = contractAuth.Authorize(bob, "test_action", func() error {
		actionExecuted = true
		return nil
	}, testArgs...)

	uassert.True(t, err != nil, "unauthorized member should not be able to propose")
	uassert.False(t, handlerCalled, "contract handler should not be called for unauthorized proposer")
	uassert.False(t, actionExecuted, "action should not be executed for unauthorized proposer")
}

func TestAutoAcceptAuthority(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	auth := NewAutoAcceptAuthority()

	// Test that any action is authorized
	executed := false
	err := auth.Authorize(alice, "test_action", func() error {
		executed = true
		return nil
	})

	uassert.True(t, err == nil, "auto-accept should not return error")
	uassert.True(t, executed, "action should have been executed")

	// Test with different caller
	random := testutils.TestAddress("random")
	executed = false
	err = auth.Authorize(random, "test_action", func() error {
		executed = true
		return nil
	})

	uassert.True(t, err == nil, "auto-accept should not care about caller")
	uassert.True(t, executed, "action should have been executed")
}

func TestAutoAcceptAuthorityWithArgs(cur realm, t *testing.T) {
	auth := NewAutoAcceptAuthority()
	anyuser := testutils.TestAddress("anyuser")

	// Test that any action is authorized with args
	executed := false
	testArgs := []any{"arg1", 42, "arg3"}
	err := auth.Authorize(anyuser, "test_action", func() error {
		executed = true
		return nil
	}, testArgs...)

	uassert.True(t, err == nil, "auto-accept should not return error")
	uassert.True(t, executed, "action should have been executed")
}

func TestMemberAuthorityMultipleMembers(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")

	// Create authority with multiple members
	auth := NewMemberAuthority(alice, bob)

	// Test that both members can execute actions
	for _, member := range []address{alice, bob} {
		err := auth.Authorize(member, "test_action", func() error {
			return nil
		})
		uassert.True(t, err == nil, "member should be authorized")
	}

	// Test that non-member cannot execute
	err := auth.Authorize(carol, "test_action", func() error {
		return nil
	})
	uassert.True(t, err != nil, "non-member should not be authorized")

	// Test Tree() functionality
	tree := auth.Tree()
	uassert.True(t, tree.Size() == 2, "tree should have 2 members")

	// Verify both members are in the tree
	found := make(map[address]bool)
	tree.Iterate("", "", func(key string, _ any) bool {
		found[address(key)] = true
		return false
	})
	uassert.True(t, found[alice], "alice should be in the tree")
	uassert.True(t, found[bob], "bob should be in the tree")
	uassert.False(t, found[carol], "carol should not be in the tree")

	// Test read-only nature of the tree
	defer func() {
		r := recover()
		uassert.True(t, r != nil, "modifying read-only tree should panic")
	}()
	tree.Set(string(carol), nil) // This should panic
}

func TestAuthorizerCurrentNeverNil(cur realm, t *testing.T) {
	auth := NewWithMembers(cur.Address())

	// Authority should never be nil after initialization
	uassert.True(t, auth.Authority() != nil, "current authority should not be nil")

	// Authority should not be nil after transfer
	var err error
	func(cur realm) { err = auth.Transfer(0, cur, NewAutoAcceptAuthority()) }(cross(cur))
	uassert.True(t, err == nil, "transfer should succeed")
	uassert.True(t, auth.Authority() != nil, "current authority should not be nil after transfer")
}

func TestContractAuthorityValidation(cur realm, t *testing.T) {
	handler := func(title string, action PrivilegedAction) error {
		return nil
	}

	// Empty path panics (consistent with NewRestrictedContractAuthority).
	uassert.PanicsWithMessage(t, cur, "contract path cannot be empty", func() {
		NewContractAuthority("", handler)
	})

	// A nil handler panics at CONSTRUCTION rather than being tolerated and
	// surfaced at Authorize time. Tolerating it produced a permanent brick:
	// see TestNilHandlerWouldBeUnrotatable below.
	uassert.PanicsWithMessage(t, cur, "contract handler cannot be nil", func() {
		NewContractAuthority("gno.land/r/test", nil)
	})

	// The Authorize-time guard stays for a zero-value ContractAuthority,
	// which the constructors cannot produce but a struct literal can.
	code := testing.NewCodeRealm("gno.land/r/test").Address()
	zero := &ContractAuthority{contractPath: "gno.land/r/test", contractAddr: code}
	err := zero.Authorize(code, "test", func() error {
		return nil
	})
	uassert.True(t, err != nil, "nil handler authority should fail to authorize")

	// Test valid configuration
	contractAuth := NewContractAuthority("gno.land/r/test", handler)
	err = contractAuth.Authorize(code, "test", func() error {
		return nil
	})
	uassert.True(t, err == nil, "valid contract authority should authorize successfully")
}

func TestAuthorizerString(cur realm, t *testing.T) {
	auth := NewWithMembers(cur.Address())
	addr := cur.Address()

	// Test initial string representation
	str := auth.String()
	uassert.Equal(t, str, "member_authority["+string(addr)+"]")

	// Test string after transfer — caller is the current member (cur).
	autoAuth := NewAutoAcceptAuthority()
	var err error
	func(cur realm) { err = auth.Transfer(0, cur, autoAuth) }(cross(cur))
	uassert.True(t, err == nil, "transfer should succeed")
	str = auth.String()
	uassert.Equal(t, str, "auto_accept_authority")

	// Test custom authority — auto-accept lets anyone transfer.
	customAuth := &mockAuthority{}
	func(cur realm) { err = auth.Transfer(0, cur, customAuth) }(cross(cur))
	uassert.True(t, err == nil, "transfer should succeed")
	str = auth.String()
	uassert.Equal(t, str, "custom_authority[mock]")
}

type mockAuthority struct{}

func (c mockAuthority) String() string { return "mock" }
func (a mockAuthority) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	// autoaccept
	return action()
}

func TestAuthorityString(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	// MemberAuthority
	memberAuth := NewMemberAuthority(alice)
	memberStr := memberAuth.String()
	expectedMemberStr := "member_authority[g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh]"
	uassert.Equal(t, memberStr, expectedMemberStr)

	// ContractAuthority — the proposer is rendered, not just the path.
	// Without it the two constructors below are indistinguishable, which
	// is what made consumer-level assertions on this string blind to an
	// authority being swapped for a wide-open one.
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error { return nil })
	contractStr := contractAuth.String()
	expectedContractStr := "contract_authority[contract=gno.land/r/test,proposer=contract-identity]"
	uassert.Equal(t, contractStr, expectedContractStr)

	// Same path, same handler, explicit open proposer — must render
	// differently from the gated default above.
	openAuth := NewRestrictedContractAuthority(
		"gno.land/r/test",
		func(title string, action PrivilegedAction) error { return nil },
		NewAutoAcceptAuthority(),
	)
	uassert.Equal(t,
		"contract_authority[contract=gno.land/r/test,proposer=auto_accept_authority]",
		openAuth.String())
	uassert.NotEqual(t, contractStr, openAuth.String())

	// AutoAcceptAuthority
	autoAuth := NewAutoAcceptAuthority()
	autoStr := autoAuth.String()
	expectedAutoStr := "auto_accept_authority"
	uassert.Equal(t, autoStr, expectedAutoStr)

	// DroppedAuthority
	droppedAuth := NewDroppedAuthority()
	droppedStr := droppedAuth.String()
	expectedDroppedStr := "dropped_authority"
	uassert.Equal(t, droppedStr, expectedDroppedStr)
}

// TestContractAuthorityUnauthorizedCaller verifies the
// fix: a ContractAuthority's default proposer is the contract itself, so
// any other caller is rejected UPSTREAM (at the proposer) and the handler
// and privileged action never run. Previously the default proposer was
// AutoAcceptAuthority, and the only guard was a handler-side
// `unsafe.CurrentRealm() == contractAddr` check — which was bypassable
// and which many consumers (e.g. r/gnops/valopers) omitted entirely,
// leaving no caller check at all.
func TestContractAuthorityUnauthorizedCaller(cur realm, t *testing.T) {
	contractPath := "gno.land/r/testcontract"
	contractAddr := chain.PackageAddress(contractPath)
	unauthorizedAddr := testutils.TestAddress("unauthorized")

	// A permissive handler that simply runs the action — the realistic
	// shape a consumer registers. The package, not the handler, must be
	// the one that rejects an unauthorized caller.
	handlerExecuted := false
	contractHandler := func(title string, action PrivilegedAction) error {
		handlerExecuted = true
		return action()
	}
	contractAuth := NewContractAuthority(contractPath, contractHandler)

	actionExecuted := false
	privilegedAction := func() error {
		actionExecuted = true
		return nil
	}

	// 1. Unauthorized caller: rejected by the contract-identity proposer
	// before the handler or the action can run.
	err := contractAuth.Authorize(unauthorizedAddr, "test_action", privilegedAction)
	uassert.Error(t, err, "unauthorized caller must be rejected")
	uassert.ErrorContains(t, err, "unauthorized", "rejection must come from the contract-identity proposer")
	uassert.False(t, handlerExecuted, "handler must not run for an unauthorized caller")
	uassert.False(t, actionExecuted, "privileged action must not run for an unauthorized caller")

	// 2. The contract itself is authorized.
	err = contractAuth.Authorize(contractAddr, "test_action", privilegedAction)
	uassert.NoError(t, err, "the contract itself must be authorized")
	uassert.True(t, handlerExecuted, "handler must run when the contract is the caller")
	uassert.True(t, actionExecuted, "privileged action must run when the contract is the caller")
}

// TestAuthorizerDoByPrevious verifies the "calling realm authorizes"
// pattern: a function (the inner crossing closure) invokes
// DoByPrevious so the authority check sees cur.Previous() — the realm
// that crossed into it — not the function's own realm.
//
// Each scenario crosses into the inner closure via cross(cur) after
// SetRealm — inside the closure, cur is the fresh live cur and
// cur.Previous() is the SetRealm'd outer realm. This is the only way
// to exercise DoByPrevious correctly under the IsCurrent guard, which
// rejects synthetic realm values (testing.MakeRealm) and stored
// stale captures.
func TestAuthorizerDoByPrevious(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	auth := NewWithMembers(alice)

	// alice (member) crosses in: cur.Previous() == alice inside the inner closure.
	testing.SetRealm(testing.NewUserRealm(alice))
	executed := false
	args := []any{"test_arg", 123}
	func(cur realm) {
		err := auth.DoByPrevious(0, cur, "test_action", func() error {
			executed = true
			return nil
		}, args...)
		uassert.NoError(t, err, "expected no error")
		uassert.True(t, executed, "action should have been executed")
	}(cross(cur))

	expectedErr := errors.New("test error")
	func(cur realm) {
		err := auth.DoByPrevious(0, cur, "test_action", func() error {
			return expectedErr
		})
		uassert.ErrorContains(t, err, expectedErr.Error(), "expected error")
	}(cross(cur))

	// bob (not a member) crosses in: Authorize must reject.
	testing.SetRealm(testing.NewUserRealm(bob))
	executed = false
	func(cur realm) {
		err := auth.DoByPrevious(0, cur, "test_action", func() error {
			executed = true
			return nil
		}, "unauthorized_arg")
		uassert.ErrorContains(t, err, "unauthorized", "expected error")
		uassert.False(t, executed, "action should not have been executed")
	}(cross(cur))
}

// ---------------------------------------------------------------------------
// Contract-identity gate: regression suite.
//
// A plain NewContractAuthority now defaults its proposer to the contract
// ITSELF (MemberAuthority of the contract's own package address) instead of
// AutoAcceptAuthority. These tests pin the security contract: only the
// contract may drive privileged actions, an external caller can neither
// drive nor Transfer the authority, the contract can still rotate its own
// authority, and the escape hatch for a genuinely-open proposer still
// exists.
// ---------------------------------------------------------------------------

// The default proposer accepts the contract itself and rejects everyone
// else, before the handler or action can run.
func TestDefaultProposerIsContractOnly(cur realm, t *testing.T) {
	const path = "gno.land/r/defcontract"
	contractAddr := chain.PackageAddress(path)

	ran := 0
	ca := NewContractAuthority(path, func(_ string, action PrivilegedAction) error {
		return action()
	})

	// The contract itself: authorized.
	err := ca.Authorize(contractAddr, "action", func() error { ran++; return nil })
	uassert.NoError(t, err, "the contract itself must be authorized")

	// Anyone else: rejected at the proposer, action never runs.
	outsider := testutils.TestAddress("outsider")
	err = ca.Authorize(outsider, "action", func() error { ran++; return nil })
	uassert.Error(t, err, "a non-contract caller must be rejected")
	uassert.ErrorContains(t, err, "unauthorized", "rejection must come from the contract-identity proposer")

	uassert.Equal(t, 1, ran, "only the contract's action must have run")
}

// An external realm that holds another realm's Authorizer cannot Transfer
// it. The caller derived under IsCurrent is that external realm, not the
// contract, so the rotation is rejected and the authority is left intact.
func TestExternalTransferRejected(cur realm, t *testing.T) {
	authorizer := NewWithAuthority(
		NewContractAuthority("gno.land/r/victim", func(_ string, action PrivilegedAction) error {
			return action()
		}),
	)

	attacker := testutils.TestAddress("attacker")
	testing.SetRealm(testing.NewUserRealm(attacker))

	var err error
	func(cur realm) { err = authorizer.Transfer(0, cur, NewDroppedAuthority()) }(cross(cur))

	uassert.Error(t, err, "external Transfer must be rejected")
	uassert.ErrorContains(t, err, "unauthorized", "rejection must come from the contract-identity proposer")

	_, ok := authorizer.Authority().(*ContractAuthority)
	uassert.True(t, ok, "authority must be unchanged after a rejected rotation")
}

// Same shape via DoByPrevious (a privileged action instead of a
// Transfer): an external caller is rejected and the action never runs.
func TestExternalDoByPreviousRejected(cur realm, t *testing.T) {
	authorizer := NewWithAuthority(
		NewContractAuthority("gno.land/r/victim2", func(_ string, action PrivilegedAction) error {
			return action()
		}),
	)

	ran := false
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("randomuser")))
	var err error
	func(cur realm) {
		err = authorizer.DoByPrevious(0, cur, "privileged", func() error { ran = true; return nil })
	}(cross(cur))

	uassert.Error(t, err, "external DoByPrevious must be rejected")
	uassert.False(t, ran, "privileged action must not run for an external caller")
}

// The contract-identity model does NOT lock out legitimate governance: the
// contract can still rotate (Transfer) its own authority when the transfer
// is driven with the contract as the previous realm (caller == contractAddr).
func TestContractCanRotateItsOwnAuthority(cur realm, t *testing.T) {
	const path = "gno.land/r/selfgov"
	authorizer := NewWithAuthority(
		NewContractAuthority(path, func(_ string, action PrivilegedAction) error {
			return action()
		}),
	)

	// Drive the transfer with the contract as the caller: set the outer
	// realm to the contract, then cross into a closure so cur.Previous()
	// resolves to the contract.
	testing.SetRealm(testing.NewCodeRealm(path))
	var err error
	func(cur realm) { err = authorizer.Transfer(0, cur, NewAutoAcceptAuthority()) }(cross(cur))

	uassert.NoError(t, err, "the contract itself must be able to rotate its authority")
	_, ok := authorizer.Authority().(*AutoAcceptAuthority)
	uassert.True(t, ok, "authority should have rotated to auto-accept")
}

// Boundary test: the raw Authority.Authorize takes `caller` as a plain
// parameter, so a holder of the raw Authority can forge caller ==
// contractAddr. This proves that even so, a forged-caller raw Authorize
// only runs the action the caller itself supplies — it CANNOT mutate an
// Authorizer's installed authority, because the Transfer mutation lives
// only inside Authorizer.Transfer (which derives caller non-forgeably).
// The lesson encoded here: consumers must reach a ContractAuthority through
// the Authorizer wrapper and must not hand out the raw Authority.
//
// WHAT THIS TEST DOES AND DOES NOT FREEZE. Only the final assertion is a
// security property. That raw Authorize ACCEPTS a forged caller is
// current behaviour, not a guarantee — it is the consequence of `caller`
// being a parameter on an exported interface method. If someone later
// hardens that path so raw Authorize stops accepting a forged caller,
// this test going red means the test is stale, NOT that the change is
// wrong: delete the two intermediate assertions and keep the last one.
// (Recorded because asserting NoError here reads like the forgeability
// is wanted. It is tolerated, and only because of what follows.)
func TestForgedCallerCannotTransfer(cur realm, t *testing.T) {
	const path = "gno.land/r/boundary"
	contractAddr := chain.PackageAddress(path)

	authorizer := NewWithAuthority(
		NewContractAuthority(path, func(_ string, action PrivilegedAction) error {
			return action()
		}),
	)

	// Attacker obtains the raw Authority and forges the contract as caller.
	attackerRan := false
	err := authorizer.Authority().Authorize(contractAddr, "transfer_authority", func() error {
		attackerRan = true
		return nil
	})

	// Current behaviour, documented above — not a property to preserve.
	uassert.NoError(t, err, "raw Authorize currently accepts a forged caller (see comment)")
	uassert.True(t, attackerRan, "only the caller's own inert closure runs")

	// THE property: the installed authority is UNCHANGED. Raw Authorize
	// cannot perform a Transfer — the mutation closure is private to
	// Authorizer.Transfer, which derives caller non-forgeably.
	_, ok := authorizer.Authority().(*ContractAuthority)
	uassert.True(t, ok, "raw Authorize must not be able to transfer the authority")
}

// Two contract authorities on different paths never cross-authorize.
func TestContractPathIsolation(cur realm, t *testing.T) {
	caB := NewContractAuthority("gno.land/r/pathb", func(_ string, action PrivilegedAction) error {
		return action()
	})
	addrA := chain.PackageAddress("gno.land/r/patha")

	ran := false
	err := caB.Authorize(addrA, "action", func() error { ran = true; return nil })
	uassert.Error(t, err, "contract A must not authorize contract B's authority")
	uassert.False(t, ran, "action must not run across contract identities")
}

// Escape hatch: a consumer that genuinely wants open proposals can still
// opt in via NewRestrictedContractAuthority with an AutoAcceptAuthority
// proposer — restoring the pre-fix "anyone can propose" behavior explicitly.
func TestRestrictedAutoAcceptEscapeHatch(cur realm, t *testing.T) {
	ca := NewRestrictedContractAuthority(
		"gno.land/r/open",
		func(_ string, action PrivilegedAction) error { return action() },
		NewAutoAcceptAuthority(),
	)

	ran := false
	err := ca.Authorize(testutils.TestAddress("anyone"), "action", func() error { ran = true; return nil })
	uassert.NoError(t, err, "an explicit AutoAccept proposer restores open proposals")
	uassert.True(t, ran, "action must run under an explicit AutoAccept proposer")
}

// Why the nil handler is now rejected at construction rather than
// tolerated: Authorize checks contractHandler == nil BEFORE consulting the
// proposer, and Transfer routes through Authorize, so there is no rotation
// path out. It is a permanent brick reachable by an ordinary deployer
// typo -- the same bricked-governance failure mode, by accident.
//
// The zero value is used because the constructor no longer allows it.
func TestNilHandlerWouldBeUnrotatable(cur realm, t *testing.T) {
	const path = "gno.land/r/nilbrick"
	addr := chain.PackageAddress(path)
	authorizer := NewWithAuthority(&ContractAuthority{
		contractPath: path,
		contractAddr: addr,
		// Both handler and proposer are nil: a struct literal is the only
		// way to reach this, and Authorize fails closed on either.
	})

	// Drive the rotation as the contract itself -- the only principal the
	// default proposer accepts. If this cannot escape, nobody can.
	testing.SetRealm(testing.NewCodeRealm(path))
	var err error
	func(cur realm) { err = authorizer.Transfer(0, cur, NewAutoAcceptAuthority()) }(cross(cur))

	uassert.Error(t, err, "a nil-handler authority must not be silently rotatable")
	_, stuck := authorizer.Authority().(*ContractAuthority)
	uassert.True(t, stuck, "nil-handler authority is unrotatable -- hence rejected at construction")
}

// A crossing closure declared HERE carries this package's identity, not
// the contract path's -- testing.SetRealm does not change that, because
// the VM mints a crossing frame's realm from the callee's declaring
// package. So the DoByCurrent shape that Example_contractAuthority teaches
// cannot be exercised from inside this package at all; it needs a real
// realm. It is covered by filetests/z_contract_authority_shape_filetest.gno.
//
// The same property is why an exported crossing closure leaks a realm's
// authority to any caller -- it is one mechanism seen from two
// sides. This test pins the half that is observable here: an unrelated
// package's frame is rejected.
func TestForeignFrameCannotDriveContractAuthority(cur realm, t *testing.T) {
	auth := NewWithAuthority(NewContractAuthority("gno.land/r/demo/dao", mockDAOHandler))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/example"))
	ran := false
	var err error
	func(cur realm) {
		err = auth.DoByCurrent(0, cur, "update_params", func() error {
			ran = true
			return nil
		})
	}(cross(cur))

	uassert.Error(t, err, "a frame from another package must not drive this authority")
	uassert.False(t, ran, "the action must not run")
}

// A malformed contract path binds the authority to an address no realm can
// ever present. Because Transfer routes through the same gate, such an
// authority is also unrotatable: a permanent brick, and the same failure
// mode the nil-handler panic exists to prevent. Rejected at construction.
//
// Pre-fix these all constructed happily; the AutoAccept default made the
// typo harmless (and insecure), so nothing complained.
func TestContractPathValidationRejectsMalformed(cur realm, t *testing.T) {
	handler := func(_ string, action PrivilegedAction) error { return action() }

	for _, path := range []string{
		"gno.land/r/gov/dao ",  // trailing space
		" gno.land/r/gov/dao",  // leading space
		"gno.land/r/gov/dao\n", // trailing newline
		"gno.land/r/gov dao",   // embedded space
		" ",
		"not a path",
		"GNO.LAND/R/GOV/DAO", // uppercase: not a legal gno pkgpath
		"gno.land//r/gov/dao",
		"gno.land/r/gov/dao/",
		"singlesegment",
	} {
		uassert.PanicsContains(t, cur, "contract path", func() {
			NewContractAuthority(path, handler)
		}, "malformed path must be rejected: "+path)
		uassert.PanicsContains(t, cur, "contract path", func() {
			NewRestrictedContractAuthority(path, handler, NewAutoAcceptAuthority())
		}, "malformed path must be rejected by the restricted ctor too: "+path)
	}

	// Real paths still construct, including dashes, digits, underscores and
	// version suffixes.
	for _, path := range []string{
		"gno.land/r/gov/dao",
		"gno.land/r/gnops/valopers",
		"gno.land/p/moul/authz/v0",
		"gno.land/r/sys/validators/v0",
		"gno.land/r/some-user/my_pkg2",
	} {
		uassert.NotPanics(t, cur, func() {
			NewContractAuthority(path, handler)
		}, "legitimate path must construct: "+path)
	}
}

// spoofProposer is a wide-open Authority that LIES in String(), returning
// the same text the canonical contract-identity default renders.
type spoofProposer struct{}

func (spoofProposer) Authorize(caller address, title string, action PrivilegedAction, args ...any) error {
	return action()
}

func (spoofProposer) String() string { return "contract-identity" }

// A foreign proposer must not be able to impersonate the gated default in
// the rendered description. ContractAuthority.String() is promoted by this
// package as the one surface distinguishing a gated authority from a
// wide-open one, and consumers assert on it (r/gnops/valopers.Auth()). An
// impl that chooses its own String() text defeated that: a fully permissive
// authority rendered byte-identical to the default, so every string-based
// configuration pin stayed green while the gate was gone.
func TestSpoofedProposerCannotImpersonateContractIdentity(cur realm, t *testing.T) {
	const path = "gno.land/r/spoofprobe"
	handler := func(_ string, action PrivilegedAction) error { return action() }
	outsider := chain.PackageAddress("gno.land/r/outsider")

	gated := NewContractAuthority(path, handler)
	spoof := NewRestrictedContractAuthority(path, handler, spoofProposer{})

	// The behavioural difference the strings must reflect.
	uassert.Error(t, gated.Authorize(outsider, "t", func() error { return nil }),
		"the gated default must reject an outsider")
	uassert.NoError(t, spoof.Authorize(outsider, "t", func() error { return nil }),
		"the spoofing proposer is wide open -- that is the point of the test")

	// ...and they do.
	uassert.NotEqual(t, gated.String(), spoof.String(),
		"a wide-open authority must not render identically to the gated default")
	uassert.Equal(t,
		"contract_authority[contract="+path+",proposer=contract-identity]",
		gated.String())
	uassert.Equal(t,
		"contract_authority[contract="+path+",proposer=custom_authority[contract-identity]]",
		spoof.String(),
		"a non-canonical proposer must be wrapped, not trusted to name itself")

	// Same guarantee through the Authorizer wrapper, which is the shape a
	// consumer realm actually exposes.
	uassert.NotEqual(t,
		NewWithAuthority(gated).String(),
		NewWithAuthority(spoof).String())
}

// An explicit proposer REPLACES the contract-identity gate; it does not add
// to it. NewRestrictedContractAuthority(govdaoPath, h, member(alice)) means
// "alice, and not GovDAO" -- contractAddr is never consulted. Pinned because
// the rendered string reads like a conjunction and the old godoc said
// "widen", so a consumer could reasonably have expected "both".
func TestExplicitProposerReplacesIdentityGate(cur realm, t *testing.T) {
	const path = "gno.land/r/gov/dao"
	handler := func(_ string, action PrivilegedAction) error { return action() }
	govdao := chain.PackageAddress(path)
	alice := chain.PackageAddress("gno.land/r/alice")

	gated := NewContractAuthority(path, handler)
	replaced := NewRestrictedContractAuthority(path, handler, NewMemberAuthority(alice))

	uassert.NoError(t, gated.Authorize(govdao, "x", func() error { return nil }),
		"the default gate accepts the bound contract")
	uassert.Error(t, gated.Authorize(alice, "x", func() error { return nil }),
		"the default gate rejects everyone else")

	uassert.Error(t, replaced.Authorize(govdao, "x", func() error { return nil }),
		"an explicit proposer REPLACES the identity gate: govdao is no longer accepted")
	uassert.NoError(t, replaced.Authorize(alice, "x", func() error { return nil }),
		"only the explicit proposer's principal is accepted")
}
