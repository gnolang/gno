package functionmodifiers

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Modifier struct {
	Name               string
	ApprovalPercentage float64
	Approvers          *avl.Tree
}

type OwnershipProposalDetails struct {
	NewOwnerPercentage float64
	OwnerProposals     *avl.Tree
}

type ModiferRegistry struct {
	Owners    *avl.Tree
	Modifiers *avl.Tree
	Proposals *OwnershipProposalDetails
}

func init() {
}

func (registry *ModiferRegistry) InitializeOwners(owners []string, newOwnerPercentage float64) {
	if len(owners) == 0 {
		panic("You must provide atleast one owner address!")
	}
	registry.initMultiSigVariables(newOwnerPercentage)

	for _, addr := range owners {
		address := std.Address(addr)
		if !address.IsValid() {
			panic(ufmt.Sprintf("Address %s is not valid!", addr))
		}
		registry.Owners.Set(addr, true)
	}
}

func (registry *ModiferRegistry) ProposeNewOwner(address string) {
	addr := std.Address(address)
	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	if !addr.IsValid() {
		panic(ufmt.Sprintf("Address %s is not valid!", addr))
	}

	v, exists := registry.Proposals.OwnerProposals.Get(addr.String())

	if !exists {
		v = avl.NewTree()
		registry.Proposals.OwnerProposals.Set(addr.String(), v)
	}

	addressApprovals := v.(*avl.Tree)

	if addressApprovals.Has(caller.String()) {
		panic("You have already approved this new owner proposal address!")
	}

	addressApprovals.Set(caller.String(), true)

	approvalsNeeded := registry.Proposals.NewOwnerPercentage * float64(registry.Owners.Size())

	if float64(addressApprovals.Size()) >= approvalsNeeded {
		registry.Proposals.OwnerProposals.Remove(addr.String())
		registry.Owners.Set(addr.String(), true)
	}
}

func (registry *ModiferRegistry) ResignOwnership() {
	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	if registry.Owners.Size() == 1 {
		panic("You cannot give up ownership because there must be atleast one owner!")
	}

	registry.Owners.Remove(caller.String())
}

func (registry *ModiferRegistry) RegisterModifier(name string, approvalPercentage float64) {
	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	if name == "" {
		panic("Must provide a non-empty name")
	}
	if registry.Modifiers.Has(name) {
		panic("Modifier with provided name already exists!")
	}
	if approvalPercentage <= 0 || approvalPercentage > 1 {
		panic("Approval must be between 0 and 1")
	}
	modifier := Modifier{
		Name:               name,
		ApprovalPercentage: approvalPercentage,
		Approvers:          avl.NewTree(),
	}
	registry.Modifiers.Set(name, modifier)
}

func (registry *ModiferRegistry) DisableModifier(name string) {
	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	if name == "" {
		panic("Must provide a non-empty name")
	}

	v, exists := registry.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)
	modifier.ApprovalPercentage = 0

	registry.Modifiers.Set(name, *modifier)
}

func (registry *ModiferRegistry) ApproveModifier(name string) {
	if name == "" {
		panic("Name of the modifier cannot be empty!")
	}

	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	v, exists := registry.Modifiers.Get(name)

	if !exists {
		panic("Modifer with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)

	if modifier.Approvers.Has(caller.String()) {
		panic("You have already approved this modifier!")
	}

	modifier.Approvers.Set(caller.String(), true)
}

func (registry *ModiferRegistry) AssertIsApproved(name string) {
	caller := std.GetOrigCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialised()

	v, exists := registry.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)
	totalOwners := registry.Owners.Size()

	approvalsPercentage := float64(modifier.Approvers.Size()) / float64(totalOwners)

	if approvalsPercentage < modifier.ApprovalPercentage {
		panic(ufmt.Sprintf("Required approval percentage: %.2f, required : %.2f", modifier.ApprovalPercentage, approvalsPercentage))
	}
}

func (registry *ModiferRegistry) initMultiSigVariables(newOwnerPercentage float64) {
	registry.Owners = avl.NewTree()
	registry.Modifiers = avl.NewTree()
	registry.Proposals = &OwnershipProposalDetails{
		NewOwnerPercentage: newOwnerPercentage,
		OwnerProposals:     avl.NewTree(),
	}
}

func (registry *ModiferRegistry) AssertIsOwner(address std.Address) {
	if !registry.Owners.Has(address.String()) {
		panic("You are not the owner!")
	}
}

func (registry *ModiferRegistry) AssertOwnersInitialised() {
	if registry.Owners.Size() == 0 {
		panic("Owners must be initialised before making any action!")
	}
}
