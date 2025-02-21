package functionmodifiers

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/addrset"
)

type Modifier struct {
	Name                            string
	ApprovalPercentage              float64
	Approvers                       *addrset.Set
	EditApprovalPercentageProposals *avl.Tree
}

type OwnershipProposalDetails struct {
	NewOwnerPercentage float64
	OwnerProposals     *avl.Tree
}

type ModifierRegistry struct {
	Owners            *addrset.Set
	Modifiers         *avl.Tree
	NewOwnerProposals *OwnershipProposalDetails
}

func init() {
}

func (registry *ModifierRegistry) InitializeOwners(owners []string, newOwnerPercentage float64) {
	if len(owners) == 0 {
		panic("You must provide at least one owner address!")
	}
	if newOwnerPercentage <= 0 || newOwnerPercentage > 1 {
		panic("New owner percentage must be greater than 0 and less or equal to 1!")
	}
	registry.initVariables(newOwnerPercentage)

	for _, addr := range owners {
		address := std.Address(addr)
		if !address.IsValid() {
			panic(ufmt.Sprintf("Address %s is not valid!", addr))
		}
		if !registry.Owners.Add(address) {
			panic(ufmt.Sprintf("Address %s is already an owner!", addr))
		}
	}
}

func (registry *ModifierRegistry) ProposeNewOwner(address string) {
	addr := std.Address(address)
	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	if !addr.IsValid() {
		panic("Address " + address + " is not valid!")
	}

	if registry.Owners.Has(addr) {
		panic("Address is already an owner!")
	}

	v, exists := registry.NewOwnerProposals.OwnerProposals.Get(address)

	if !exists {
		v = avl.NewTree()
		registry.NewOwnerProposals.OwnerProposals.Set(address, v)
	}

	addressApprovals := v.(*avl.Tree)

	if addressApprovals.Has(caller.String()) {
		panic("You have already approved this new owner proposal address!")
	}

	addressApprovals.Set(caller.String(), true)

	approvalsNeeded := registry.NewOwnerProposals.NewOwnerPercentage * float64(registry.Owners.Size())

	if float64(addressApprovals.Size()) >= approvalsNeeded {
		registry.NewOwnerProposals.OwnerProposals.Remove(address)
		registry.Owners.Add(addr)
	}
}

func (registry *ModifierRegistry) ResignOwnership() {
	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	if registry.Owners.Size() == 1 {
		panic("You cannot give up ownership because there must be at least one owner!")
	}

	registry.Owners.Remove(caller)
}

func (registry *ModifierRegistry) RegisterModifier(name string, approvalPercentage float64) {
	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	if name == "" {
		panic("Must provide a non-empty name")
	}
	if registry.Modifiers.Has(name) {
		panic("Modifier with provided name already exists!")
	}
	if approvalPercentage <= 0 || approvalPercentage > 1 {
		panic("Approval must be between 0 and 1")
	}
	modifier := &Modifier{
		Name:                            name,
		ApprovalPercentage:              approvalPercentage,
		Approvers:                       &addrset.Set{},
		EditApprovalPercentageProposals: avl.NewTree(),
	}
	registry.Modifiers.Set(name, modifier)
}

func (registry *ModifierRegistry) EditModifierApprovalPercentage(name string, newApprovalPercentage float64) {
	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	if name == "" {
		panic("Must provide a non-empty name")
	}

	if newApprovalPercentage <= 0 || newApprovalPercentage > 1 {
		panic("New approval percentage value must be greater than 0 and less or equal to 1!")
	}

	v, exists := registry.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)

	approvals, exists := modifier.EditApprovalPercentageProposals.Get(strconv.FormatFloat(newApprovalPercentage, 'f', -1, 64))
	if !exists {
		approvals = avl.NewTree()
	}

	approvalsTree := approvals.(*avl.Tree)
	approvalsTree.Set(caller.String(), true)

	modifier.EditApprovalPercentageProposals.Set(strconv.FormatFloat(newApprovalPercentage, 'f', -1, 64), approvalsTree)

	totalOwners := registry.Owners.Size()
	approvalsNeeded := modifier.ApprovalPercentage * float64(totalOwners)

	if float64(approvalsTree.Size()) >= approvalsNeeded {
		modifier.ApprovalPercentage = newApprovalPercentage
		modifier.EditApprovalPercentageProposals.Remove(strconv.FormatFloat(newApprovalPercentage, 'f', -1, 64))
	}

	registry.Modifiers.Set(name, modifier)
}

func (registry *ModifierRegistry) ApproveModifier(name string) {
	if name == "" {
		panic("Name of the modifier cannot be empty!")
	}

	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	v, exists := registry.Modifiers.Get(name)

	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)

	if modifier.Approvers.Has(caller) {
		panic("You have already approved this modifier!")
	}

	modifier.Approvers.Add(caller)
}

func (registry *ModifierRegistry) AssertIsApproved(name string) {
	caller := std.OriginCaller()

	registry.AssertIsOwner(caller)
	registry.AssertOwnersInitialized()

	v, exists := registry.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)
	totalOwners := registry.Owners.Size()

	approvalsPercentage := float64(modifier.Approvers.Size()) / float64(totalOwners)

	if approvalsPercentage < modifier.ApprovalPercentage {
		panic(ufmt.Sprintf("Required approval percentage: %f, achieved: %f", modifier.ApprovalPercentage, approvalsPercentage))
	}
}

func (registry *ModifierRegistry) initVariables(newOwnerPercentage float64) {
	registry.Owners = &addrset.Set{}
	registry.Modifiers = avl.NewTree()
	registry.NewOwnerProposals = &OwnershipProposalDetails{
		NewOwnerPercentage: newOwnerPercentage,
		OwnerProposals:     avl.NewTree(),
	}
}

func (registry *ModifierRegistry) AssertIsOwner(address std.Address) {
	if !registry.Owners.Has(address) {
		panic("You are not the owner!")
	}
}

func (registry *ModifierRegistry) AssertOwnersInitialized() {
	if registry.Owners.Size() == 0 {
		panic("Owners must be initialized before making any action!")
	}
}
