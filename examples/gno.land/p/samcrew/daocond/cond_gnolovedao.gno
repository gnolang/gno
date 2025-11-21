package daocond

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt"
)

type gnolovDaoCondThreshold struct {
	threshold            float64
	roles                []string
	hasRoleFn            func(memberId string, role string) bool
	usersWithRoleCountFn func(role string) uint32
}

var roleWeights = []float64{3.0, 2.0, 1.0}

// Creates a weighted voting condition based on role tiers with dynamic power calculation.
// Example: GnoloveDAOCondThreshold(0.5, ["admin", "mod", "user"], hasRoleFn, countFn)
// This is our implementation of the govdao condition following Jae Kwon's specification: https://gist.github.com/jaekwon/918ad325c4c8f7fb5d6e022e33cb7eb3
func GnoloveDAOCondThreshold(threshold float64, roles []string, hasRoleFn func(memberId string, role string) bool, usersWithRoleCountFn func(role string) uint32) Condition {
	if threshold <= 0 || threshold > 1 {
		panic(errors.New("invalid threshold"))
	}
	if usersWithRoleCountFn == nil {
		panic(errors.New("nil usersWithRoleCountFn"))
	}
	if hasRoleFn == nil {
		panic(errors.New("nil hasRoleFn"))
	}
	if len(roles) > 3 {
		panic("the gnolove dao condition handles at most 3 roles")
	}
	return &gnolovDaoCondThreshold{
		threshold:            threshold,
		roles:                roles,
		hasRoleFn:            hasRoleFn,
		usersWithRoleCountFn: usersWithRoleCountFn,
	}
}

// Checks if weighted voting power meets the threshold.
func (c *gnolovDaoCondThreshold) Eval(ballot Ballot) bool {
	return c.voteRatio(ballot, VoteYes) >= c.threshold
}

// Returns progress toward meeting the weighted threshold between 0.0 and 1.0.
func (c *gnolovDaoCondThreshold) Signal(ballot Ballot) float64 {
	return math.Min(c.voteRatio(ballot, VoteYes)/c.threshold, 1)
}

// Displays the condition with role weights as text.
// Example output: "50% of total voting power | admin => 3.00 power | mod => 2.00 power"
func (c *gnolovDaoCondThreshold) Render() string {
	rolePowers := []string{}
	for i, role := range c.roles {
		weight := strconv.FormatFloat(roleWeights[i], 'f', 2, 64) // ufmt.Sprintf("%.2f", ...) is not working
		rolePowers = append(rolePowers, ufmt.Sprintf("%s => %s power", role, weight))
	}
	return ufmt.Sprintf("%g%% of total voting power | %s", c.threshold*100, strings.Join(rolePowers, " | "))
}

// Displays the condition with current weighted vote counts and breakdown.
// Shows detailed voting power analysis and vote distribution across roles.
func (c *gnolovDaoCondThreshold) RenderWithVotes(ballot Ballot) string {
	vPowers, totalPower := c.computeVotingPowers()
	rolePowers := []string{}
	for _, role := range c.roles {
		weight := strconv.FormatFloat(vPowers[role], 'f', 2, 64) // ufmt.Sprintf("%.2f", ...) is not working
		rolePowers = append(rolePowers, ufmt.Sprintf("%s => %s power", role, weight))
	}
	s := ""
	s += ufmt.Sprintf("%g%% of total voting power | %s\n\n", c.threshold*100, strings.Join(rolePowers, " | "))
	s += ufmt.Sprintf("Threshold needed: %g%% of total voting power\n\n", c.threshold*100)
	s += ufmt.Sprintf("Yes: %d/%d\n\n", c.voteRatio(ballot, VoteYes), totalPower)
	s += ufmt.Sprintf("No: %d/%d\n\n", c.voteRatio(ballot, VoteNo), totalPower)
	s += ufmt.Sprintf("Abstain: %d/%d\n\n", c.voteRatio(ballot, VoteAbstain), totalPower)
	s += ufmt.Sprintf("Voting power needed: %g%% of total voting power\n\n", c.threshold*totalPower)
	return s
}

var _ Condition = (*gnolovDaoCondThreshold)(nil)

func (c *gnolovDaoCondThreshold) voteRatio(ballot Ballot, vote Vote) float64 {
	var total float64
	votingPowersByTier, totalPower := c.computeVotingPowers()
	// Case when there are zero T1s
	if totalPower == 0.0 {
		return totalPower
	}

	ballot.Iterate(func(voter string, v Vote) bool {
		if v != vote {
			return false
		}
		tier := c.getUserRole(voter)
		total += votingPowersByTier[tier]
		return false
	})
	return total / totalPower
}

func (c *gnolovDaoCondThreshold) getUserRole(userID string) string {
	for _, role := range c.roles {
		if c.hasRoleFn(userID, role) {
			return role
		}
	}
	panic("No role found for user")
}

func (c *gnolovDaoCondThreshold) computeVotingPowers() (map[string]float64, float64) {
	votingPowers := make(map[string]float64)
	totalPower := 0.0
	countsMembersPerRole := make(map[string]float64)

	for _, role := range c.roles {
		countsMembersPerRole[role] = float64(c.usersWithRoleCountFn(role))
	}

	for i, role := range c.roles {
		if i == 0 {
			votingPowers[role] = roleWeights[0] // Highest tier always gets max power (3.0)
		} else {
			votingPowers[role] = computePower(countsMembersPerRole[c.roles[0]], countsMembersPerRole[role], roleWeights[i])
		}
		totalPower += votingPowers[role] * countsMembersPerRole[role]
	}

	return votingPowers, totalPower
}

// max power here is the number of votes each tier gets when we have
// the same number of member on each tier
// T2 = 2.0 and T1 = 1.0 with the ration T1/Tn
// we compute the actual ratio
func computePower(T1, Tn, maxPower float64) float64 {
	// If there are 0 Tn (T2, T3) just return the max power
	// we could also return 0.0 as voting power
	if Tn <= 0.0 {
		return maxPower
	}

	computedPower := (T1 / Tn) * maxPower
	if computedPower >= maxPower {
		// If computed power is bigger than the max, this happens if Tn is lower than T1
		// cap the max power to max power.
		return maxPower
	}
	return computedPower
}
