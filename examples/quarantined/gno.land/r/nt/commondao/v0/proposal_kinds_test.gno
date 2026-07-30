package commondao

import (
	"testing"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*manageKindsPropDefinition)(nil)
	_ commondao.Validable          = (*manageKindsPropDefinition)(nil)
)

// probeKind is a foreign proposal kind with a configurable name, for the
// register-by-value cases.
type probeKind string

func (k probeKind) Name() string { return string(k) }
func (probeKind) New(_ commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
	return nil, nil
}

func TestManageKindsKindNew(t *testing.T) {
	// The test DAO registers the full catalog; deregister the text kind so
	// the table has an absent catalog kind to register and to reject
	// deregistering.
	dao := newTestDAO(t, "KindsHost")
	urequire.NoError(t, dao.DeregisterKind(kindText))

	cases := []struct {
		name string
		args manageKindsArgs
		err  string
	}{
		// Register by name.
		{"register unknown kind", manageKindsArgs{dao: dao, name: "unknown"}, "unknown proposal kind"},
		{"register an enabled kind", manageKindsArgs{dao: dao, name: kindCouncilUpdate}, "proposal kind is already enabled"},
		{"register a catalog kind", manageKindsArgs{dao: dao, name: kindText}, ""},
		{"register the opt-in execution kind", manageKindsArgs{dao: dao, name: kindExecution}, ""},
		// Register by value (foreign kind).
		{"register a foreign kind by value", manageKindsArgs{dao: dao, kind: probeKind("custom")}, ""},
		{"register a nil-named kind by value", manageKindsArgs{dao: dao, kind: probeKind("")}, "proposal kind requires a name"},
		{"register an already-enabled kind by value", manageKindsArgs{dao: dao, kind: probeKind(kindCouncilUpdate)}, "proposal kind is already enabled"},
		// Deregister by name.
		{"deregister an absent kind", manageKindsArgs{dao: dao, remove: true, name: kindText}, "proposal kind is not enabled"},
		{"deregister the manage-kinds kind", manageKindsArgs{dao: dao, remove: true, name: kindManageKinds}, "the manage-kinds kind cannot be deregistered"},
		{"deregister an enabled kind", manageKindsArgs{dao: dao, remove: true, name: kindCouncilUpdate}, ""},
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
			d, ok := def.(manageKindsPropDefinition)
			urequire.True(t, ok, "expect a manageKindsPropDefinition")
			uassert.True(t, d.dao == dao, "expect the definition to keep the host DAO")
			uassert.Equal(t, tc.args.remove, d.remove)
			uassert.Equal(t, tc.args.name, d.name)
			uassert.True(t, d.kind == tc.args.kind, "expect the definition to keep the supplied kind")
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
	brick := manageKindsPropDefinition{dao: dao, remove: true, name: kindManageKinds}
	err := brick.Validate()
	urequire.Error(t, err)
	uassert.Equal(t, "the manage-kinds kind cannot be deregistered", err.Error())

	// Deregistering any other kind validates cleanly.
	ok := manageKindsPropDefinition{dao: dao, remove: true, name: kindText}
	uassert.NoError(t, ok.Validate())

	// A register never trips the self-brick regardless of name.
	reg := manageKindsPropDefinition{dao: dao, name: kindManageKinds}
	uassert.NoError(t, reg.Validate())
}
