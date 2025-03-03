package functionmodifiers

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestInitializeOwners(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test initializing with valid owners and invalid percentage
	uassert.PanicsWithMessage(t, "New owner percentage must be greater than 0 and less or equal to 1!", func() {
		modifierRegistry.InitializeOwners(owners, 1.1)
	})
	uassert.PanicsWithMessage(t, "New owner percentage must be greater than 0 and less or equal to 1!", func() {
		modifierRegistry.InitializeOwners(owners, -0.1)
	})

	//Test initializing
	uassert.PanicsWithMessage(t, "You must provide at least one owner address!", func() {
		modifierRegistry.InitializeOwners([]string{}, 0.5)
	})
}

func TestProposeNewOwner(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test proposing new owner without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.ProposeNewOwner(nonOwner)
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test proposing new owner when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.ProposeNewOwner(nonOwner)
	})

	//Test proposing new owner when already an owner
	uassert.PanicsWithMessage(t, "Address is already an owner!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ProposeNewOwner(owners[0])
	})

	//Test proposing new owner which is already proposed
	uassert.PanicsWithMessage(t, "You have already approved this new owner proposal address!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ProposeNewOwner(nonOwner)
		modifierRegistry.ProposeNewOwner(nonOwner)
	})

	//Test if valid proposal is accepted
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[1]))
		modifierRegistry.ProposeNewOwner(nonOwner)
		urequire.Equal(t, modifierRegistry.Owners.Size(), 3)
	})

}

func TestResignOwnership(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test resigning ownership without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.ResignOwnership()
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test resigning ownership when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.ResignOwnership()
	})

	//Test valid resignation
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[1]))
		modifierRegistry.ResignOwnership()
		urequire.Equal(t, modifierRegistry.Owners.Size(), 1)
	})

	//Test resigning ownership when there is only one owner
	uassert.PanicsWithMessage(t, "You cannot give up ownership because there must be at least one owner!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ResignOwnership()
	})

}

func TestRegisterModifier(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test registering modifier without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.RegisterModifier("test", 0.15)
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test registering modifier when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.RegisterModifier("test", 0.15)
	})

	//Test registering modifier with empty name
	uassert.PanicsWithMessage(t, "Must provide a non-empty name", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.RegisterModifier("", 0.15)
	})

	//Test registering modifier with invalid approval percentage
	uassert.PanicsWithMessage(t, "Approval must be between 0 and 1", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.RegisterModifier("test", 1.1)
		modifierRegistry.RegisterModifier("test", -0.1)

	})

	//Test registering a valid modifier
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.RegisterModifier("test", 0.15)
		urequire.Equal(t, modifierRegistry.Modifiers.Size(), 1)
	})

	//Test registering a modifier with the same name
	uassert.PanicsWithMessage(t, "Modifier with provided name already exists!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.RegisterModifier("test", 0.15)
	})

}

func TestEditModifierApprovalPercentage(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test editing modifier approval percentage without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.EditModifierApprovalPercentage("test", 0.15)
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test editing modifier approval percentage when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.EditModifierApprovalPercentage("test", 0.15)
	})

	//Test editing modifier approval percentage with invalid approval percentage
	uassert.PanicsWithMessage(t, "New approval percentage value must be greater than 0 and less or equal to 1!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.EditModifierApprovalPercentage("test", 1.1)
		modifierRegistry.EditModifierApprovalPercentage("test", -0.1)
	})

	//Test editing modifier approval percentage with a non-existent modifier
	uassert.PanicsWithMessage(t, "Modifier with provided name doesn't exist!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.EditModifierApprovalPercentage("test2", 0.15)
	})

	modifierRegistry.RegisterModifier("test", 0.3)

	//Test editing modifier approval percentage with a valid approval percentage
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.EditModifierApprovalPercentage("test", 0.15)
		modifier, exists := modifierRegistry.Modifiers.Get("test")
		urequire.True(t, exists)
		urequire.Equal(t, modifier.(*Modifier).ApprovalPercentage, 0.15)
	})

}

func TestApproveModifier(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test approving modifier without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.ApproveModifier("test")
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test approving modifier when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.ApproveModifier("test")
	})

	//Test approving modifier with a non-existent modifier
	uassert.PanicsWithMessage(t, "Modifier with provided name doesn't exist!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ApproveModifier("test2")
	})

	modifierRegistry.RegisterModifier("test", 0.3)

	//Test approving modifier with a valid modifier
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ApproveModifier("test")
		modifier, exists := modifierRegistry.Modifiers.Get("test")
		urequire.True(t, exists)
		urequire.Equal(t, modifier.(*Modifier).Approvers.Size(), 1)
	})

	//Test approving modifier when already approved
	uassert.PanicsWithMessage(t, "You have already approved this modifier!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ApproveModifier("test")
	})

}

func TestAssertIsApproved(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	//Test asserting is approved without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.AssertIsApproved("test")
	})

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test asserting is approved when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.AssertIsApproved("test")
	})

	std.TestSetOriginCaller(std.Address(owners[0]))
	modifierRegistry.RegisterModifier("test", 0.3)

	//Test asserting is approved with a non-existent modifier
	uassert.PanicsWithMessage(t, "Modifier with provided name doesn't exist!", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.AssertIsApproved("test2")
	})

	//Test asserting is approved with a valid modifier but not enough approvals
	uassert.PanicsWithMessage(t, "Required approval percentage: 0.300000, achieved: 0.000000", func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.AssertIsApproved("test")
	})

	//Test asserting is approved with a valid modifier and enough approvals
	uassert.NotPanics(t, func() {
		std.TestSetOriginCaller(std.Address(owners[0]))
		modifierRegistry.ApproveModifier("test")
		modifierRegistry.AssertIsApproved("test")
	})
}

func TestAssertIsOwner(t *testing.T) {
	owners := []string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"}
	nonOwner := "g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp"

	modifierRegistry := ModifierRegistry{}

	modifierRegistry.InitializeOwners(owners, 0.6)

	//Test asserting is owner with a non-existent address
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		modifierRegistry.AssertIsOwner(std.Address(nonOwner))
	})

	//Test asserting is owner with a valid address
	uassert.NotPanics(t, func() {
		modifierRegistry.AssertIsOwner(std.Address(owners[0]))
	})

}

func TestAssertOwnersInitialized(t *testing.T) {
	modifierRegistry := ModifierRegistry{}

	//Test asserting owners are initialized without initialization
	uassert.PanicsWithMessage(t, "Owners must be initialized before making any action!", func() {
		modifierRegistry.AssertOwnersInitialized()
	})

	modifierRegistry.InitializeOwners([]string{"g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"}, 0.6)

	//Test asserting owners are initialized with a valid initialization
	uassert.NotPanics(t, func() {
		modifierRegistry.AssertOwnersInitialized()
	})
}
