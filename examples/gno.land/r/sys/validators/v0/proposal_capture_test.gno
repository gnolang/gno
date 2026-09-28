package validators

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/init/v0"
	sysparams "gno.land/r/sys/params"
)

// TestValidatorProposalExecutesTheApprovedChanges covers the stored proposal executing the approved entry.
func TestValidatorProposalExecutesTheApprovedChanges(cur realm, t *testing.T) {
	resetValset(t)
	resetCache()

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	changes := []ValoperChange{{OperatorAddress: opA, Power: 5}}
	pr := NewValidatorProposalRequest(cur, changes, "add opA", "")
	uassert.True(t, contains(pr.Description(), opA.String()+": add (power 5)"))

	member := testutils.TestAddress("gov-member")
	testing.SetRealm(testing.NewUserRealm(member))
	ini.InitWithUsers(cross(cur), member)

	id := dao.MustCreateProposal(cross(cur), pr)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(id)))

	// The caller rewrites its array after the vote.
	changes[0] = ValoperChange{OperatorAddress: opB, Power: 99}

	dao.ExecuteProposal(cross(cur), id)

	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 1, len(effective))
	uassert.Equal(t, pubKeyA, effective[0].PubKey)
	uassert.Equal(t, uint64(5), effective[0].VotingPower)
}
