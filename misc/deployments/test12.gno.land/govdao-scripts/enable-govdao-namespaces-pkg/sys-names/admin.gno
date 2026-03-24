package names

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
)

// NewNamespacePropRequest returns a ProposalRequest to register a custom namespace for an address.
// Submit the result to r/gov/dao via MsgRun: dao.MustCreateProposal(cross, pr).
func NewNamespacePropRequest(namespace string, addr address) dao.ProposalRequest {
	if namespace == "" {
		panic("namespace cannot be empty")
	}
	if !addr.IsValid() {
		panic("invalid address")
	}
	cb := func(_ realm) error {
		namespaceStore.Set(namespace, addr)
		return nil
	}
	desc := ufmt.Sprintf("Register namespace %q for address %s in r/sys/names/v2.", namespace, addr.String())
	return dao.NewProposalRequest("Register Namespace", desc, dao.NewSimpleExecutor(cb, ""))
}

// NewNamespaceRemovalPropRequest returns a ProposalRequest to remove a registered namespace.
// Submit the result to r/gov/dao via MsgRun: dao.MustCreateProposal(cross, pr).
func NewNamespaceRemovalPropRequest(namespace string) dao.ProposalRequest {
	if namespace == "" {
		panic("namespace cannot be empty")
	}
	if !namespaceStore.Has(namespace) {
		panic("namespace not registered: " + namespace)
	}
	cb := func(_ realm) error {
		if _, removed := namespaceStore.Remove(namespace); !removed {
			return errors.New("namespace not found: " + namespace)
		}
		return nil
	}
	desc := ufmt.Sprintf("Remove namespace %q from r/sys/names/v2.", namespace)
	return dao.NewProposalRequest("Remove Namespace", desc, dao.NewSimpleExecutor(cb, ""))
}
