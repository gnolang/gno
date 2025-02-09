package authority

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	alice       = std.Address("g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh")
	bob         = std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")
	testPkgPath = "gno.land/r/demo/test"
)

func TestNew(t *testing.T) {
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	a := New()
	if a.Owner() != alice.String() {
		t.Errorf("New() owner = %v, want %v", a.Owner(), alice)
	}
	if a.typeOf != OwnerTypeWallet {
		t.Errorf("New() typeOf = %v, want %v", a.typeOf, OwnerTypeWallet)
	}
}

func TestNewWithAddress(t *testing.T) {
	a := NewWithAddress(alice)
	if a.Owner() != alice.String() {
		t.Errorf("NewWithAddress() owner = %v, want %v", a.Owner(), alice)
	}
	if a.typeOf != OwnerTypeWallet {
		t.Errorf("NewWithAddress() typeOf = %v, want %v", a.typeOf, OwnerTypeWallet)
	}
}

func TestDo_WalletType(t *testing.T) {
	a := NewWithAddress(alice)

	// Test successful action as admin
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	var executed bool
	err := a.Do("TestAction", func() error {
		executed = true
		return nil
	})
	uassert.NoError(t, err)
	uassert.True(t, executed, "action should be executed for wallet admin")

	// Test unauthorized action
	realm = std.NewUserRealm(bob)
	std.TestSetOrigCaller(bob)
	std.TestSetRealm(realm)

	executed = false
	err = a.Do("TestAction", func() error {
		executed = true
		return nil
	})
	uassert.Error(t, err)
	uassert.False(t, executed, "action should not be executed for unauthorized caller")
}

func TestDo_ContractType(t *testing.T) {
	a := NewWithAddress(alice)

	std.TestSetOrigCaller(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetRealm(realm)

	var handlerCalled bool
	var receivedTitle string

	handler := func(title string, action PrivilegedAction) error {
		handlerCalled = true
		receivedTitle = title
		return action()
	}

	// First transfer ownership to contract
	err := a.TransferToContract(testPkgPath, handler)
	uassert.NoError(t, err)

	realm = std.NewCodeRealm(testPkgPath)
	std.TestSetRealm(realm)
	std.TestSetOrigCaller(realm.Addr())

	var executed bool
	err = a.Do("ContractAction", func() error {
		executed = true
		return nil
	})

	uassert.NoError(t, err)
	uassert.True(t, handlerCalled, "handler should be called for contract authority")
	if receivedTitle != "ContractAction" {
		t.Errorf("handler received title = %v, want %v", receivedTitle, "ContractAction")
	}
	uassert.True(t, executed, "action should be executed")

}

func TestTransferToWallet(t *testing.T) {
	a := NewWithAddress(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	// Test successful transfer
	err := a.TransferToWallet(bob)
	uassert.NoError(t, err)
	if a.Owner() != bob.String() {
		t.Errorf("Owner after transfer = %v, want %v", a.Owner(), bob)
	}
	if a.typeOf != OwnerTypeWallet {
		t.Errorf("typeOf after transfer = %v, want %v", a.typeOf, OwnerTypeWallet)
	}

	// Test transfer with invalid address
	err = a.TransferToWallet("")
	uassert.Error(t, err)
}

func TestTransferToContract(t *testing.T) {
	a := NewWithAddress(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	handler := func(title string, action PrivilegedAction) error { return nil }

	// Test successful transfer
	err := a.TransferToContract(testPkgPath, handler)
	uassert.NoError(t, err)
	if a.Owner() != testPkgPath {
		t.Errorf("Owner after transfer = %v, want %v", a.Owner(), testPkgPath)
	}
	if a.typeOf != OwnerTypeContract {
		t.Errorf("typeOf after transfer = %v, want %v", a.typeOf, OwnerTypeContract)
	}

	// Test transfer with invalid package path
	err = a.TransferToContract("", handler)
	uassert.Error(t, err)

	// Test transfer with nil handler
	err = a.TransferToContract(testPkgPath, nil)
	uassert.Error(t, err)
}

func TestDrop(t *testing.T) {
	a := NewWithAddress(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	err := a.Drop()
	uassert.NoError(t, err)
	uassert.Empty(t, a.Owner())

	// Check against the correct type constant
	if a.typeOf != OwnerTypeDropped { // OwnerTypeNone should be 0
		t.Errorf("typeOf after drop = %v, want %v", a.typeOf, 0)
	}
}

func TestNilReceiver(t *testing.T) {
	var a *Authority

	// Test Owner()
	uassert.Empty(t, a.Owner())

	// Test Do()
	err := a.Do("TestAction", func() error { return nil })
	uassert.Error(t, err)
}

func TestDo_ActionError(t *testing.T) {
	a := NewWithAddress(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)

	expectedErr := errors.New("action failed")
	err := a.Do("TestAction", func() error {
		return expectedErr
	})
	uassert.Error(t, err)
	if err != expectedErr {
		t.Errorf("Do() error = %v, want %v", err, expectedErr)
	}
}

func TestDo_ContractActionError(t *testing.T) {
	a := NewWithAddress(alice)
	realm := std.NewUserRealm(alice)
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(realm)
	expectedErr := errors.New("action failed")

	var handlerExecuted bool
	handler := func(title string, action PrivilegedAction) error {
		handlerExecuted = true
		return action()
	}

	// First transfer ownership to contract
	err := a.TransferToContract(testPkgPath, handler)
	uassert.NoError(t, err)

	// Use testPkgPath directly for realm creation
	realm = std.NewCodeRealm(testPkgPath)
	std.TestSetOrigCaller(std.Address(testPkgPath))
	std.TestSetRealm(realm)

	err = a.Do("TestAction", func() error {
		return expectedErr
	})

	// We expect the exact error returned from the action
	uassert.Error(t, err)
	uassert.True(t, handlerExecuted, "handler should have been executed")
}

func TestNewWithContract(t *testing.T) {
	handler := func(title string, action PrivilegedAction) error { return nil }

	// Test successful creation
	a, err := NewWithContract(testPkgPath, handler)
	uassert.NoError(t, err)
	if a.Owner() != testPkgPath {
		t.Errorf("NewWithContract() owner = %v, want %v", a.Owner(), testPkgPath)
	}
	if a.typeOf != OwnerTypeContract {
		t.Errorf("NewWithContract() typeOf = %v, want %v", a.typeOf, OwnerTypeContract)
	}

	// Test with empty package path
	a, err = NewWithContract("", handler)
	uassert.Error(t, err)
	uassert.Nil(t, a)

	// Test with nil handler
	a, err = NewWithContract(testPkgPath, nil)
	uassert.Error(t, err)
	uassert.Nil(t, a)
}
