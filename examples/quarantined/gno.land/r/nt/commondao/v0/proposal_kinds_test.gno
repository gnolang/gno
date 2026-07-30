package commondao

import (
	"testing"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = manageKindsProposal{}
	_ commondao.Validable          = manageKindsProposal{}
	_ commondao.Executable         = manageKindsProposal{}
)

func TestManageKindsKindNew(t *testing.T) {
	// The test DAO registers the full catalog; deregister the text kind so
	// the table has an absent catalog kind to register and to reject
	// deregistering.
	dao := newTestDAO(t, "KindsHost")
	urequire.NoError(t, dao.DeregisterKind(kindText))

	cases := []struct {
		name string
		args manageKindsProposal
		err  string
	}{
		// Register by name.
		{"register unknown kind", manageKindsProposal{dao: dao, name: "unknown"}, "unknown proposal kind"},
		{"register a registered kind", manageKindsProposal{dao: dao, name: kindCouncilUpdate}, "proposal kind is already registered"},
		{"register a catalog kind", manageKindsProposal{dao: dao, name: kindText}, ""},
		{"register the opt-in execution kind", manageKindsProposal{dao: dao, name: kindExecution}, ""},
		// Deregister by name.
		{"deregister an absent kind", manageKindsProposal{dao: dao, remove: true, name: kindText}, "proposal kind is not registered: " + kindText},
		{"deregister the manage-kinds kind", manageKindsProposal{dao: dao, remove: true, name: kindManageKinds}, "the manage-kinds kind cannot be deregistered"},
		{"deregister a registered kind", manageKindsProposal{dao: dao, remove: true, name: kindCouncilUpdate}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def, err := manageKindsKind{}.New(dao.Readonly(), tc.args)

			if tc.err != "" {
				urequire.Error(t, err)
				uassert.Equal(t, tc.err, err.Error())
				return
			}

			urequire.NoError(t, err)
			// New returns the args value unchanged as the definition (the
			// collapsed args/definition type): the two must be equal.
			d, ok := def.(manageKindsProposal)
			urequire.True(t, ok, "expect a manageKindsProposal")
			uassert.True(t, d.dao == dao, "expect the definition to keep the host DAO")
			uassert.Equal(t, tc.args.remove, d.remove)
			uassert.Equal(t, tc.args.name, d.name)
		})
	}
}

// TestManageKindsValidateSelfBrick pins the second (Execute-time) self-brick
// layer: Validate rejects deregistering the manage-kinds kind even though
// the /p/ package no longer locks any name. Validate reruns inside Execute.
func TestManageKindsValidateSelfBrick(t *testing.T) {
	dao := newTestDAO(t, "ValidateHost")

	// A manage-kinds deregister slips past New (constructed directly here);
	// Validate is the backstop that still rejects it.
	brick := manageKindsProposal{dao: dao, remove: true, name: kindManageKinds}
	err := brick.Validate()
	urequire.Error(t, err)
	uassert.Equal(t, "the manage-kinds kind cannot be deregistered", err.Error())

	// Deregistering any other kind validates cleanly.
	ok := manageKindsProposal{dao: dao, remove: true, name: kindText}
	uassert.NoError(t, ok.Validate())

	// A register never trips the self-brick regardless of name.
	reg := manageKindsProposal{dao: dao, name: kindManageKinds}
	uassert.NoError(t, reg.Validate())
}
