package modifiersdemo

import (
	"gno.land/p/ursulovic/functionmodifiers"
)

var (
	Modifiers functionmodifiers.ModifierRegistry
)

func init() {
}

func ProposeNewOwner(address string) {
	Modifiers.ProposeNewOwner(address)
}

func ResignOwnership() {
	Modifiers.ResignOwnership()
}

func RegisterModifier(name string, approvalPercentage float64) {
	Modifiers.RegisterModifier(name, approvalPercentage)
}

func ApproveModifier(name string) {
	Modifiers.ApproveModifier(name)
}

func AssertIsApproved(name string) {
	Modifiers.AssertIsApproved(name)
}

func EditModifierApprovalPercentage(name string, newApprovalPercentage float64) {
	Modifiers.EditModifierApprovalPercentage(name, newApprovalPercentage)
}

func Render(s string) string {
	return ""
}
