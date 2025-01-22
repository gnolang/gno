package multisig

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

type MultiSig struct {
	Owners    *avl.Tree
	Modifiers *avl.Tree
	Proposals *OwnershipProposalDetails
}

func init() {
}

func (multiSig *MultiSig) InitializeOwners(owners []string, newOwnerPercentage float64) {
	if len(owners) == 0 {
		panic("You must provide atleast one owner address!")
	}
	multiSig.initMultiSigVariables(newOwnerPercentage)

	for _, addr := range owners {
		address := std.Address(addr)
		if !address.IsValid() {
			panic(ufmt.Sprintf("Address %s is not valid!", addr))
		}
		multiSig.Owners.Set(addr, true)
	}
}

func (multiSig *MultiSig) ProposeNewOwner(address string) {
	addr := std.Address(address)
	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	if !addr.IsValid() {
		panic(ufmt.Sprintf("Address %s is not valid!", addr))
	}

	v, exists := multiSig.Proposals.OwnerProposals.Get(addr.String())

	if !exists {
		v = avl.NewTree()
		multiSig.Proposals.OwnerProposals.Set(addr.String(), v)
	}

	addressApprovals := v.(*avl.Tree)

	if addressApprovals.Has(caller.String()) {
		panic("You have already approved this new owner proposal address!")
	}

	addressApprovals.Set(caller.String(), true)

	approvalsNeeded := multiSig.Proposals.NewOwnerPercentage * float64(multiSig.Owners.Size())

	if float64(addressApprovals.Size()) >= approvalsNeeded {
		multiSig.Proposals.OwnerProposals.Remove(addr.String())
		multiSig.Owners.Set(addr.String(), true)
	}
}

func (multiSig *MultiSig) ResignOwnership() {
	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	if multiSig.Owners.Size() == 1 {
		panic("You cannot give up ownership because there must be atleast one owner!")
	}

	multiSig.Owners.Remove(caller.String())
}

func (multiSig *MultiSig) RegisterModifier(name string, approvalPercentage float64) {
	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	if name == "" {
		panic("Must provide a non-empty name")
	}
	if multiSig.Modifiers.Has(name) {
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
	multiSig.Modifiers.Set(name, modifier)
}

func (multiSig *MultiSig) DisableModifier(name string) {
	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	if name == "" {
		panic("Must provide a non-empty name")
	}

	v, exists := multiSig.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)
	modifier.ApprovalPercentage = 0

	multiSig.Modifiers.Set(name, *modifier)
}

func (multiSig *MultiSig) ApproveModifier(name string) {
	if name == "" {
		panic("Name of the modifier cannot be empty!")
	}

	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	v, exists := multiSig.Modifiers.Get(name)

	if !exists {
		panic("Modifer with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)

	if modifier.Approvers.Has(caller.String()) {
		panic("You have already approved this modifier!")
	}

	modifier.Approvers.Set(caller.String(), true)
}

func (multiSig *MultiSig) AssertIsApproved(name string) {
	caller := std.GetOrigCaller()

	multiSig.assertIsOwner(caller)
	multiSig.assertOwnersInitialised()

	v, exists := multiSig.Modifiers.Get(name)
	if !exists {
		panic("Modifier with provided name doesn't exist!")
	}

	modifier := v.(*Modifier)
	totalOwners := multiSig.Owners.Size()

	approvalsPercentage := float64(modifier.Approvers.Size()) / float64(totalOwners)

	if approvalsPercentage < modifier.ApprovalPercentage {
		panic(ufmt.Sprintf("Required approval percentage: %.2f, required : %.2f", modifier.ApprovalPercentage, approvalsPercentage))
	}
}

func (multiSig *MultiSig) initMultiSigVariables(newOwnerPercentage float64) {
	multiSig.Owners = avl.NewTree()
	multiSig.Modifiers = avl.NewTree()
	multiSig.Proposals = &OwnershipProposalDetails{
		NewOwnerPercentage: newOwnerPercentage,
		OwnerProposals:     avl.NewTree(),
	}
}

func (multiSig *MultiSig) assertIsOwner(address std.Address) {
	if !multiSig.Owners.Has(address.String()) {
		panic("You are not the owner!")
	}
}

func (multiSig *MultiSig) assertOwnersInitialised() {
	if multiSig.Owners.Size() == 0 {
		panic("Owners must be initialised before making any action!")
	}
}
