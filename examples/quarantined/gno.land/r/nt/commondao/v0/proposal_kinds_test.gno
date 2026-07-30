package commondao

import (
	"testing"

	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = (*setProposalKindPropDefinition)(nil)
)

func TestProposalKindsKindNew(t *testing.T) {
	// The test DAO registers the full catalog; deregister the text kind so
	// the table has an absent kind to enable and to reject disabling.
	dao := newTestDAO(t, "KindsHost")
	urequire.NoError(t, dao.DeregisterKind(kindText))

	cases := []struct {
		name     string
		kindName string
		enabled  bool
		err      string
	}{
		{"unknown kind", "unknown", true, "unknown proposal kind"},
		{"enable an enabled kind", kindCouncilUpdate, true, "proposal kind is already enabled"},
		{"disable an absent kind", kindText, false, "proposal kind is not enabled"},
		{"disable the governance kind", kindProposalKinds, false, "the proposal-kinds kind cannot be disabled"},
		{"enable", kindText, true, ""},
		{"disable", kindCouncilUpdate, false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def, err := proposalKindsKind{}.New(dao.Readonly(), setProposalKindArgs{dao, tc.kindName, tc.enabled})

			if tc.err != "" {
				urequire.Error(t, err)
				uassert.Equal(t, tc.err, err.Error())
				return
			}

			urequire.NoError(t, err)
			d, ok := def.(setProposalKindPropDefinition)
			urequire.True(t, ok, "expect a setProposalKindPropDefinition")
			uassert.True(t, d.dao == dao, "expect the definition to keep the host DAO")
			uassert.Equal(t, tc.kindName, d.kindName)
			uassert.Equal(t, tc.enabled, d.enabled)
		})
	}
}
