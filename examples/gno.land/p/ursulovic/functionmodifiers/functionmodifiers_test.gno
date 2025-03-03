package functionmodifiers

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
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

	modifierRegistry.InitializeOwners(owners, 0.3)

	//Test proposing new owner when not an owner
	uassert.PanicsWithMessage(t, "You are not the owner!", func() {
		std.TestSetOriginCaller(std.Address(nonOwner))
		modifierRegistry.ProposeNewOwner(nonOwner)
	})

	



}