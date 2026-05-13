package authz

import (
	"chain"
	"chain/runtime"
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
	err := auth.Do(0, cur, "test_action", func() error {
		executed = true
		return nil
	}, args...)

	uassert.True(t, err == nil, "expected no error")
	uassert.True(t, executed, "action should have been executed")

	// Test unauthorized action with args
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("bob")))

	executed = false
	err = auth.Do(0, cur, "test_action", func() error {
		executed = true
		return nil
	}, "unauthorized_arg")

	uassert.True(t, err != nil, "expected error")
	uassert.False(t, executed, "action should not have been executed")

	// Test action returning error
	testing.SetRealm(testing.NewUserRealm(alice))
	expectedErr := errors.New("test error")

	err = auth.Do(0, cur, "test_action", func() error {
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
	func(cur realm) { err = auth.Transfer(0, cur, newAuth) }(cross)
	uassert.True(t, err == nil, "expected no error")
	uassert.True(t, auth.Authority() == newAuth, "expected new authority")

	// Test unauthorized transfer: principal is not a member of newAuth.
	carol := testutils.TestAddress("carol")
	testing.SetRealm(testing.NewUserRealm(carol))

	func(cur realm) { err = auth.Transfer(0, cur, NewMemberAuthority(alice)) }(cross)
	uassert.True(t, err != nil, "expected error")

	// Test transfer to contract authority — bob is the current authority.
	testing.SetRealm(testing.NewUserRealm(bob))
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		return action()
	})

	func(cur realm) { err = auth.Transfer(0, cur, contractAuth) }(cross)
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
	func(cur realm) { err = auth.Transfer(0, cur, memberAuth) }(cross)
	uassert.True(t, err == nil, "unexpected error in first transfer")

	// Then transfer to a contract authority — bob is now the authority.
	testing.SetRealm(testing.NewUserRealm(bob))
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error {
		return action()
	})
	func(cur realm) { err = auth.Transfer(0, cur, contractAuth) }(cross)
	uassert.True(t, err == nil, "unexpected error in second transfer")

	// Finally transfer to an auto-accept authority — must come from the
	// contract realm so ContractAuthority's wrappedAction CurrentRealm
	// check passes. Use the test frame's cur directly (SetRealm mutates
	// it in place); a crossing closure would push a new frame whose cur
	// is the test package, breaking the runtime.CurrentRealm match.
	autoAuth := NewAutoAcceptAuthority()
	codeRealm := testing.NewCodeRealm("gno.land/r/test")
	testing.SetRealm(codeRealm)
	err = auth.Transfer(0, cur, autoAuth)
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
	func(cur realm) { err = auth.Transfer(0, cur, attackerAuth) }(cross)

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
	func(cur realm) { err = auth.Transfer(0, cur, NewDroppedAuthority()) }(cross)
	uassert.True(t, err == nil, "expected no error")

	// Try to execute action
	err = auth.Do(0, cur, "test_action", func() error {
		return nil
	})
	uassert.True(t, err != nil, "expected error from dropped authority")

	// Try to transfer again
	func(cur realm) { err = auth.Transfer(0, cur, NewMemberAuthority(alice)) }(cross)
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
	func(cur realm) { err = auth.Transfer(0, cur, NewAutoAcceptAuthority()) }(cross)
	uassert.True(t, err == nil, "transfer should succeed")
	uassert.True(t, auth.Authority() != nil, "current authority should not be nil after transfer")
}

func TestContractAuthorityValidation(cur realm, t *testing.T) {
	/*
		// Test empty path - should panic
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			NewContractAuthority("", nil)
		}()
		uassert.True(t, panicked, "expected panic for empty path")
	*/

	// Test nil handler - should return error on Authorize
	auth := NewContractAuthority("gno.land/r/test", nil)
	code := testing.NewCodeRealm("gno.land/r/test").Address()
	err := auth.Authorize(code, "test", func() error {
		return nil
	})
	uassert.True(t, err != nil, "nil handler authority should fail to authorize")

	// Test valid configuration
	handler := func(title string, action PrivilegedAction) error {
		return nil
	}
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
	func(cur realm) { err = auth.Transfer(0, cur, autoAuth) }(cross)
	uassert.True(t, err == nil, "transfer should succeed")
	str = auth.String()
	uassert.Equal(t, str, "auto_accept_authority")

	// Test custom authority — auto-accept lets anyone transfer.
	customAuth := &mockAuthority{}
	func(cur realm) { err = auth.Transfer(0, cur, customAuth) }(cross)
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

	// ContractAuthority
	contractAuth := NewContractAuthority("gno.land/r/test", func(title string, action PrivilegedAction) error { return nil })
	contractStr := contractAuth.String()
	expectedContractStr := "contract_authority[contract=gno.land/r/test]"
	uassert.Equal(t, contractStr, expectedContractStr)

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

func TestContractAuthorityUnauthorizedCaller(cur realm, t *testing.T) {
	contractPath := "gno.land/r/testcontract"
	contractAddr := chain.PackageAddress(contractPath)
	unauthorizedAddr := testutils.TestAddress("unauthorized")

	// Handler that checks the caller before proceeding
	handlerExecutedCorrectly := false // Tracks if handler logic ran correctly
	handlerErrorMsg := "handler: caller is not the contract"
	contractHandler := func(title string, action PrivilegedAction) error {
		caller := runtime.CurrentRealm().Address()
		if caller != contractAddr {
			return errors.New(handlerErrorMsg)
		}
		// Only execute action and mark success if caller is correct
		handlerExecutedCorrectly = true
		return action()
	}

	contractAuth := NewContractAuthority(contractPath, contractHandler)
	authorizer := NewWithAuthority(contractAuth) // Start with ContractAuthority

	actionExecuted := false
	privilegedAction := func() error {
		actionExecuted = true
		return nil
	}

	// 1. Attempt action from unauthorized user
	testing.SetRealm(testing.NewUserRealm(unauthorizedAddr))
	err := authorizer.Do(0, cur, "test_action_unauthorized", privilegedAction)

	// Assertions for unauthorized call
	uassert.Error(t, err, "DoByCurrent should return an error for unauthorized caller")
	uassert.ErrorContains(t, err, handlerErrorMsg, "Error should originate from the handler check")
	uassert.False(t, handlerExecutedCorrectly, "Handler should not have executed successfully for unauthorized caller")
	uassert.False(t, actionExecuted, "Privileged action should not have executed for unauthorized caller")

	// 2. Attempt action from the correct contract
	handlerExecutedCorrectly = false // Reset flag
	actionExecuted = false           // Reset flag
	testing.SetRealm(testing.NewCodeRealm(contractPath))
	err = authorizer.Do(0, cur, "test_action_authorized", privilegedAction)

	// Assertions for authorized call
	uassert.NoError(t, err, "DoByCurrent should succeed for authorized contract caller")
	uassert.True(t, handlerExecutedCorrectly, "Handler should have executed successfully for authorized caller")
	uassert.True(t, actionExecuted, "Privileged action should have executed for authorized caller")
}

// TestAuthorizerDoByPrevious verifies the "previous realm authorizes"
// pattern: caller invokes auth.Do with cur.Previous() so the authority
// check sees the realm that crossed into us, not us.
func TestAuthorizerDoByPrevious(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	auth := NewWithMembers(alice)

	// Simulate: alice EOA crossed into r/test/test. From r/test/test's
	// perspective cur.Previous() is alice — an authorized member.
	const testPkgPath = "gno.land/r/test/test"
	aliceRealm := testing.MakeRealm(alice, "", testing.OriginRealm())
	aliceCrossed := testing.MakeRealm(chain.PackageAddress(testPkgPath), testPkgPath, aliceRealm)

	executed := false
	args := []any{"test_arg", 123}
	err := auth.Do(0, aliceCrossed.Previous(), "test_action", func() error {
		executed = true
		return nil
	}, args...)
	uassert.NoError(t, err, "expected no error")
	uassert.True(t, executed, "action should have been executed")

	expectedErr := errors.New("test error")
	err = auth.Do(0, aliceCrossed.Previous(), "test_action", func() error {
		return expectedErr
	})
	uassert.ErrorContains(t, err, expectedErr.Error(), "expected error")

	// Simulate: bob (not a member) crossed into r/test/test. Authorize
	// must reject.
	bobRealm := testing.MakeRealm(bob, "", testing.OriginRealm())
	bobCrossed := testing.MakeRealm(chain.PackageAddress(testPkgPath), testPkgPath, bobRealm)

	executed = false
	err = auth.Do(0, bobCrossed.Previous(), "test_action", func() error {
		executed = true
		return nil
	}, "unauthorized_arg")
	uassert.ErrorContains(t, err, "unauthorized", "expected error")
	uassert.False(t, executed, "action should not have been executed")
}
