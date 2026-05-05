package users

import (
	"chain"
	"chain/runtime"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/gov/dao"
)

const initControllerPath = "gno.land/r/sys/users/init"

var controllers = addrset.Set{} // caller whitelist

func init() {
	// auto-whitelist the init controller for bootstrapping for testing chain.
	if chainID := runtime.ChainID(); chainID == "dev" {
		controllers.Add(chain.PackageAddress(initControllerPath))
	}
}

// AddControllerAtGenesis allows adding a controller during chain genesis (height 0).
// This is mostly useful for testing.
func AddControllerAtGenesis(_ realm, addr address) {
	height := runtime.ChainHeight()
	if height > 0 {
		panic("AddControllerAtGenesis can only be called at genesis (height 0)")
	}

	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	controllers.Add(addr)
}

// ProposeNewController allows GovDAO to add a whitelisted caller
func ProposeNewController(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return addToWhitelist(addr)
	}

	desc := "This proposal adds " + addr.String() + " to `sys/users` realm's callers whitelist."
	return dao.NewProposalRequest("Add Whitelisted Caller to \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerRemoval allows GovDAO to add a whitelisted caller
func ProposeControllerRemoval(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return deleteFromWhitelist(addr)
	}

	desc := "This proposal removes " + addr.String() + " from `sys/users` realm's callers whitelist."
	return dao.NewProposalRequest("Remove Whitelisted Caller From \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerAdditionAndRemoval allows GovDAO to add a new caller and remove an old caller in the same proposal.
func ProposeControllerAdditionAndRemoval(toAdd, toRemove address) dao.ProposalRequest {
	if !toAdd.IsValid() || !toRemove.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return applyControllerSwap(toAdd, toRemove)
	}

	desc := ufmt.Sprintf(
		"This proposal adds %s and removes %s from `sys/users` realm's callers whitelist.",
		toAdd,
		toRemove,
	)
	return dao.NewProposalRequest("Add and Remove Whitelisted Callers From \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// applyControllerSwap is the callback body of ProposeControllerAdditionAndRemoval,
// extracted so it can be unit-tested without driving the full GovDAO flow.
//
// The desired end state is "toAdd is in the whitelist AND toRemove is out".
// Both operations are made idempotent so the swap doesn't silently no-op when
// the chain state has drifted between proposal creation and execution:
//
//   - If toAdd is already whitelisted, treat addToWhitelist's
//     ErrAlreadyWhitelisted as benign and continue to the remove step.
//   - If toRemove is already absent, treat deleteFromWhitelist's
//     ErrNotWhitelisted as benign and return success.
//
// Without this idempotency, the original code returned early on an "already
// whitelisted" toAdd and skipped the remove entirely — a swap proposal could
// pass governance and silently leave the old controller active. (audit
// finding #6)
func applyControllerSwap(toAdd, toRemove address) error {
	if err := addToWhitelist(toAdd); err != nil && err != ErrAlreadyWhitelisted {
		return err
	}
	if err := deleteFromWhitelist(toRemove); err != nil {
		if _, alreadyOut := err.(ErrNotWhitelisted); !alreadyOut {
			return err
		}
	}
	return nil
}

// ProposeRegisterUser allows GovDAO to register a name without checking
// controllers. The executor closure runs with ignoreCanonical=true (decision
// #3): DAO grants always bypass canonical-collision detection. Voters see
// any collision in the proposal description and can vote NO if unintended.
func ProposeRegisterUser(name string, addr address) dao.ProposalRequest {
	// Validate the name and address now, even though registerUser will validate again
	if err := validateName(name); err != nil {
		panic(err.Error())
	}
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	desc := "This proposal registers " + name + " with address " + addr.String() + " in `sys/users`."
	if existing, taken := IsCanonicalTaken(name); taken && existing != name {
		desc += "\n\nCANONICAL COLLISION: this name's canonical form matches the existing registration of `" +
			existing + "`. DAO grants bypass canonical-collision detection — the proposal will succeed if voted in. " +
			"If the collision is unintended, vote NO."
	}

	cb := func(cur realm) error {
		return registerUser(cur, name, addr, true) // bypass canonical (decision #3)
	}

	return dao.NewProposalRequest("Register User to \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeUpdateName allows GovDAO to update a name with an alias without
// checking controllers. Like ProposeRegisterUser, the executor runs with
// ignoreCanonical=true (decision #3).
func ProposeUpdateName(addr address, newName string) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}
	if err := validateName(newName); err != nil {
		panic(err.Error())
	}

	desc := "This proposal updates address " + addr.String() + " with alias " + newName + " in `sys/users`."
	if existing, taken := IsCanonicalTaken(newName); taken && existing != newName {
		desc += "\n\nCANONICAL COLLISION: the new alias's canonical form matches the existing registration of `" +
			existing + "`. DAO grants bypass canonical-collision detection — the proposal will succeed if voted in. " +
			"If the collision is unintended, vote NO."
	}

	cb := func(cur realm) error {
		data := ResolveAddress(addr)
		if data == nil {
			return ErrUserNotExistOrDeleted
		}
		return data.updateName(newName, true) // bypass canonical (decision #3)
	}

	return dao.NewProposalRequest("Update Name Alias in \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeDeleteUser allows GovDAO to delete a user without checking controllers
func ProposeDeleteUser(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		data := ResolveAddress(addr)
		if data == nil {
			return ErrUserNotExistOrDeleted
		}
		return data.delete()
	}

	desc := "This proposal deletes the user with address " + addr.String() + " in `sys/users`."
	return dao.NewProposalRequest("Delete User in \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// IsController reports whether the given address is currently in the
// controller whitelist. Returns the same boolean that gating checks
// (RegisterUser, UpdateName, Delete) use internally — useful for
// off-chain monitoring and for governance proposals to inspect state
// before voting.
func IsController(addr address) bool {
	return controllers.Has(addr)
}

// Controllers returns a snapshot of the current controller whitelist.
// The returned slice is a fresh copy; mutating it does not affect realm
// state. Order is the iteration order of the underlying address set.
//
// Audit finding #20: without this getter, the controller whitelist was
// opaque from outside the package — operators had to read source or
// replay every governance proposal to know who could write to the
// registry. This is the read-only API that closes that gap.
func Controllers() []address {
	out := make([]address, 0, controllers.Size())
	controllers.IterateByOffset(0, controllers.Size(), func(a address) bool {
		out = append(out, a)
		return false
	})
	return out
}

// Helpers

func deleteFromWhitelist(addr address) error {
	if !controllers.Has(addr) {
		return NewErrNotWhitelisted()
	}

	if ok := controllers.Remove(addr); !ok {
		return ErrWhitelistRemoveFailed
	}

	return nil
}

func addToWhitelist(newCaller address) error {
	if !controllers.Add(newCaller) {
		return ErrAlreadyWhitelisted
	}

	return nil
}
