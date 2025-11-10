package validators

import (
	"chain"
	"fmt"
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
						VotingPower: 1,
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
						VotingPower: 1,
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
						VotingPower: 1,
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
		for _, existing := range vp.GetValidators() {
			removeValidator(existing.Address)
		}
		legacy := make([]pVals.Validator, 0, 4)
		for i := 1; i <= 4; i++ {
			val := pVals.Validator{
				Address:     testutils.TestAddress(fmt.Sprintf("legacyTooMany%d", i)),
				PubKey:      fmt.Sprintf("legacyTooMany%d_pubkey", i),
				VotingPower: 1,
			}
			legacy = append(legacy, val)
			addValidator(val)
		}

		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				updates := []pVals.Validator{
					{
						Address:     legacy[0].Address,
						PubKey:      legacy[0].PubKey,
						VotingPower: 0,
					},
					{
						Address:     legacy[1].Address,
						PubKey:      legacy[1].PubKey,
						VotingPower: 0,
					},
					{
						Address:     legacy[2].Address,
						PubKey:      legacy[2].PubKey,
						VotingPower: 0,
					},
				}
				updates = append(updates, pVals.Validator{
					Address:     testutils.TestAddress("tooManyNew1"),
					PubKey:      "tooManyNew1_pubkey",
					VotingPower: 1,
				})
				updates = append(updates, pVals.Validator{
					Address:     testutils.TestAddress("tooManyNew2"),
					PubKey:      "tooManyNew2_pubkey",
					VotingPower: 1,
				})
				return updates
			}
			pr := NewPropRequest(changesFn, "Replace validator set", "Attempt to replace >2/3 of validators")
			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid)
		urequire.NoError(t, err, "proposal not found")
		urequire.Equal(t, "Attempt to replace >2/3 of validators", proposal.Description())

		setupTestCoins()
		urequire.AbortsWithMessage(t, errTooManyChanges, func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: pid,
			})
			dao.ExecuteProposal(cross, pid)
		})
	})

	t.Run("Proposal incorrectly allows majority-new validators", func(t *testing.T) {
		// Test: Proposal should reject when >2/3 of the resulting set are brand new validators.
		lastValsetUpdate = time.Now().Add(-valsetUpdateCooldown - time.Minute)

		// Reset any prior state so we can craft the scenario deterministically.
		for _, existing := range vp.GetValidators() {
			removeValidator(existing.Address)
		}

		legacyValidators := []pVals.Validator{
			{
				Address:     testutils.TestAddress("legacy1"),
				PubKey:      "legacy1_pubkey",
				VotingPower: 1,
			},
			{
				Address:     testutils.TestAddress("legacy2"),
				PubKey:      "legacy2_pubkey",
				VotingPower: 1,
			},
			{
				Address:     testutils.TestAddress("legacy3"),
				PubKey:      "legacy3_pubkey",
				VotingPower: 1,
			},
		}

		for _, val := range legacyValidators {
			addValidator(val)
		}

		setupTestCoins()

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(testUserAddr))
			changesFn := func() []pVals.Validator {
				return []pVals.Validator{
					{
						Address:     testutils.TestAddress("newval1"),
						PubKey:      "newval1_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval2"),
						PubKey:      "newval2_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval3"),
						PubKey:      "newval3_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval4"),
						PubKey:      "newval4_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval5"),
						PubKey:      "newval5_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval6"),
						PubKey:      "newval6_pubkey",
						VotingPower: 1,
					},
					{
						Address:     testutils.TestAddress("newval7"),
						PubKey:      "newval7_pubkey",
						VotingPower: 1,
					},
				}
			}
			pr := NewPropRequest(changesFn, "Add majority-new validators", "Should reject >2/3 new validators")
			pid = dao.MustCreateProposal(cross, pr)
		})

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
