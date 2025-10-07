package validators

import (
	"chain"
	"testing"
	"time"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
	pVals "gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/v3/init" // ensures govdao initializer runs
	susers "gno.land/r/sys/users"
)

var testUserAddr = testutils.TestAddress("g1user")

func init() {
	daoinit.InitWithUsers(testUserAddr)
	registerTestUsers(testUserAddr)
}

const usersV1RealmPath = "gno.land/r/gnoland/users/v1"

// Registers test users in the users realm for proposal voting.
func registerTestUsers(addrs ...address) {
	testing.SetRealm(testing.NewCodeRealm(usersV1RealmPath))
	for _, addr := range addrs {
		err := susers.RegisterUser(cross, addr.String()[1:], addr)
		if err != nil {
			panic("failed to register test user: " + err.Error() + " : " + addr.String())
		}
	}
}

// Test suite for validator proposal logic.
func TestValidatorProposals(t *testing.T) {
	const (
		minRegisterFee int64 = 20 * 1_000_000 // minimum gnot to register validator
		minProposalFee int64 = 100 * 1_000_000

		testMoniker     = "moniker"
		testDescription = "description"
		testPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	)

	// Helper to set up coins for registration and proposal.
	setupTestCoins := func() {
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", minRegisterFee)})
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", minProposalFee)})
	}

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(testUserAddr))

	t.Run("Proposal creation succeeds", func(t *testing.T) {
		// Test: Successfully create and execute a validator proposal.
		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				return []pVals.Validator{
					{
						Address:     testutils.TestAddress("valoper1"),
						PubKey:      "valoper1_pubkey",
						VotingPower: 10,
					},
				}
			}
			pr := NewPropRequest(changesFn, "Add validator 1", "Add validator 1 to set")
			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid)
		urequire.NoError(t, err, "proposal not found")
		urequire.Equal(t, "Add validator 1 to set", proposal.Description())

		setupTestCoins()
		urequire.NotPanics(t, func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: pid,
			})
			dao.ExecuteProposal(cross, pid)
		})
	})

	t.Run("Proposal fails due to cooldown", func(t *testing.T) {
		// Test: Proposal execution fails if cooldown period not elapsed.
		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				return []pVals.Validator{
					{
						Address:     testutils.TestAddress("valoper2"),
						PubKey:      "valoper2_pubkey",
						VotingPower: 5,
					},
				}
			}
			pr := NewPropRequest(changesFn, "Add validator 2", "Add validator 2 to set")
			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid)
		urequire.NoError(t, err, "proposal not found")
		urequire.Equal(t, "Add validator 2 to set", proposal.Description())

		setupTestCoins()
		urequire.AbortsWithMessage(t, errValsetUpdateCooldown, func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: pid,
			})
			dao.ExecuteProposal(cross, pid)
		})
	})

	t.Run("Proposal succeeds after cooldown", func(t *testing.T) {
		// Test: Proposal executes successfully after cooldown period.
		lastValsetUpdate = time.Now().Add(-valsetUpdateCooldown - time.Minute)
		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				return []pVals.Validator{
					{
						Address:     testutils.TestAddress("valoper3"),
						PubKey:      "valoper3_pubkey",
						VotingPower: 5,
					},
				}
			}
			pr := NewPropRequest(changesFn, "Add validator 3", "Add validator 3 to set")
			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid)
		urequire.NoError(t, err, "proposal not found")
		urequire.Equal(t, "Add validator 3 to set", proposal.Description())

		setupTestCoins()
		urequire.NotPanics(t, func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: pid,
			})
			dao.ExecuteProposal(cross, pid)
		})
	})

	t.Run("Proposal fails due to too many changes", func(t *testing.T) {
		// Test: Proposal execution fails if too many validators are changed.
		lastValsetUpdate = time.Now().Add(-valsetUpdateCooldown - time.Minute)
		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				return []pVals.Validator{
					{
						Address:     testutils.TestAddress("valoper4"),
						PubKey:      "valoper4_pubkey",
						VotingPower: 6,
					},
					{
						Address:     testutils.TestAddress("valoper3"),
						PubKey:      "valoper3_pubkey",
						VotingPower: 0,
					},
				}
			}
			pr := NewPropRequest(changesFn, "Add validators 4 and 5", "Attempt to add too many validators")
			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid)
		urequire.NoError(t, err, "proposal not found")
		urequire.Equal(t, "Attempt to add too many validators", proposal.Description())

		setupTestCoins()
		urequire.AbortsWithMessage(t, errTooManyChanges, func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: pid,
			})
			dao.ExecuteProposal(cross, pid)
		})
	})
}
