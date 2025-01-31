package modifiersdemo

import (
	"gno.land/p/ursulovic/functionmodifiers"
)

var (
	modifiers functionmodifiers.ModifierRegistry
)

func init() {
	owners := []string{"g1sz9vzfgthtu370l8w3gyl2rg4qmz8txn2tw5vp", "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7", "g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"}
	newOwnerPercentage := 0.5
	modifiers.InitializeOwners(owners, newOwnerPercentage)
}

func RegisterModifier(name string, approvalPercentage float64) {
	modifiers.RegisterModifier(name, approvalPercentage)
}

func ApproveModifier(name string) {
	modifiers.ApproveModifier(name)
}

func AssertIsApproved(name string) {
	modifiers.AssertIsApproved(name)
}

func EditModifierApprovalPercentage(name string, newApprovalPercentage float64) {
	modifiers.EditModifierApprovalPercentage(name, newApprovalPercentage)
}

func Render(s string) string {
	return ""
}
