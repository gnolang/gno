package daocond

import (
	"testing"

	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
)

/*
Example 1:
T1 100 members --> 300 VP, 3 votes per member
T2 100 members --> 200 VP, 2 votes per member
T3 100 members --> 100 VP, 1 votes per member
Example 2:

T1 100 members --> 300 VP, 3 votes per member
T2 50 members --> 100 VP, 2 votes per member *
T3 10 members --> 10 VP, 1 votes per member *
Example 3:

T1 100 members --> 300 VP, 3 votes per member
T2 200 members --> 200 VP, 1 votes per member *
T3 100 members --> 100 VP, 1 votes per member
Example 4:

T1 100 members --> 300 VP, 3 votes per member
T2 200 members --> 200 VP, 1 votes per member *
T3 1000 members --> 100 VP, 0.1 votes per member
*/
func TestComputeVotingPowers(t *testing.T) {
	type gnoloveDaoComposition struct {
		t1s                int
		t2s                int
		t3s                int
		abstainT3          bool
		expectedTotalPower float64
		expectedPowers     map[string]float64
	}
	tests := map[string]gnoloveDaoComposition{
		"example 1": {
			t1s:                100,
			t2s:                100,
			t3s:                100,
			abstainT3:          false,
			expectedTotalPower: 600,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 2.0,
				"T3": 1.0,
			},
		},
		"example 2": {
			t1s:                100,
			t2s:                50,
			t3s:                10,
			abstainT3:          false,
			expectedTotalPower: 410,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 2.0,
				"T3": 1.0,
			},
		},
		"example 3": {
			t1s:                100,
			t2s:                200,
			t3s:                100,
			abstainT3:          false,
			expectedTotalPower: 600,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 1.0,
				"T3": 1.0,
			},
		},
		"example 4": {
			t1s:                100,
			t2s:                200,
			t3s:                1000,
			abstainT3:          false,
			expectedTotalPower: 600,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 1.0,
				"T3": 0.1,
			},
		},
		"0 -T1s": {
			t1s:                0,
			t2s:                100,
			t3s:                100,
			abstainT3:          false,
			expectedTotalPower: 0,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 0.0,
				"T3": 0.0,
			},
		},
		"100 T1, 1 T2, 1 T3": {
			t1s:                100,
			t2s:                1,
			t3s:                1,
			abstainT3:          false,
			expectedTotalPower: 303,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 2.0,
				"T3": 1.0,
			},
		},
		"T3 Abstaining": {
			t1s:                100,
			t2s:                100,
			t3s:                100,
			expectedTotalPower: 500,
			abstainT3:          true,
			expectedPowers: map[string]float64{
				"T1": 3.0,
				"T2": 2.0,
			},
		},
	}

	for name, composition := range tests {
		t.Run(name, func(t *testing.T) {
			dao := newMockDAO()
			for i := 0; i < composition.t1s; i++ {
				dao.addUser(ufmt.Sprintf("%d_T1", i), []string{"T1"})
			}
			for i := 0; i < composition.t2s; i++ {
				dao.addUser(ufmt.Sprintf("%d_T2", i), []string{"T2"})
			}
			for i := 0; i < composition.t3s; i++ {
				dao.addUser(ufmt.Sprintf("%d_T3", i), []string{"T3"})
			}

			roles := []string{"T1", "T2"}
			if !composition.abstainT3 {
				roles = append(roles, "T3")
			}

			cond := &gnolovDaoCondThreshold{
				threshold:            0.6,
				hasRoleFn:            dao.hasRole,
				roles:                roles,
				usersWithRoleCountFn: dao.usersWithRoleCount,
			}
			votingPowers, totalPower := cond.computeVotingPowers()
			for tier, expectedPower := range composition.expectedPowers {
				if votingPowers[tier] != expectedPower {
					t.Fail()
				}
			}

			if totalPower != composition.expectedTotalPower {
				t.Fail()
			}
		})
	}
}

func TestEval(t *testing.T) {
	type gnoloveDaoVotes struct {
		votesT1      []Vote
		votesT2      []Vote
		votesT3      []Vote
		threshold    float64
		expectedEval bool
		expectedYes  float64
		abstainT3    bool
		panic        bool
	}
	tests := map[string]gnoloveDaoVotes{
		"2/6% Yes": { //0.3333
			votesT1:      []Vote{VoteNo},                             // 3 voting power
			votesT2:      []Vote{VoteYes, VoteYes, VoteYes, VoteYes}, // 2 voting power combined
			votesT3:      []Vote{VoteNo},
			expectedEval: false,
			abstainT3:    false,
			threshold:    0.45,
			expectedYes:  2.0 / 6.0,
			panic:        false,
		},
		"50% Yes": {
			votesT1:      []Vote{VoteNo},                             // 3 voting power
			votesT2:      []Vote{VoteYes, VoteYes, VoteYes, VoteYes}, // 2 voting power combined
			votesT3:      []Vote{VoteYes},
			expectedEval: true,
			abstainT3:    false,
			threshold:    0.45,
			expectedYes:  3.0 / 6.0,
			panic:        false,
		},
		"several T2 & T3": {
			votesT1: []Vote{VoteNo, VoteNo}, // 6 voting power
			// 10 T2 total power (2/3) powerT1 = 4, 4/10 0.4 each
			votesT2: []Vote{VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes},
			// 4 T3 total power (1/3) powerT1 = 2, 2/4 0.5 each
			votesT3:      []Vote{VoteYes, VoteNo, VoteYes, VoteNo},
			expectedEval: false,
			abstainT3:    false,
			threshold:    0.42, //total power = 6+4+2 T3yes = 1, T2yes = 4 T1yes = 0 totalYes = 0.41666666666 (5/12)
			expectedYes:  5.0 / 12.0,
			panic:        false,
		},
		"several T2 & T3 eval true": {
			votesT1: []Vote{VoteNo, VoteNo}, // 6 voting power
			// 10 T2 total power (2/3) powerT1 = 4, 4/10 0.4 each
			votesT2: []Vote{VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes, VoteYes},
			// 4 T3 total power (1/3) powerT1 = 2, 2/4 0.5 each
			votesT3:      []Vote{VoteYes, VoteYes, VoteYes, VoteNo},
			expectedEval: true,
			abstainT3:    false,
			threshold:    0.42, //total power = 6+4+2 T3yes = 1.5, T2yes = 4 T1yes = 0 totalYes = 0.45833333333 (5.5/12)
			expectedYes:  5.5 / 12.0,
			panic:        false,
		},
		"only T1s": {
			votesT1:      []Vote{VoteYes, VoteNo}, // 6 voting power
			expectedEval: false,
			abstainT3:    false,
			threshold:    0.6,
			expectedYes:  3.0 / 6.0,
			panic:        false,
		},
		"only T3s": { // as T2 & T3 power is capped as power of T1, in this case the power will be 0 everywhere
			votesT3:      []Vote{VoteYes, VoteYes, VoteYes, VoteYes}, // voting power = 0, 0 each
			expectedEval: false,
			abstainT3:    false,
			threshold:    0.6,
			expectedYes:  0.0,
		},
		"T3 abstaining": {
			votesT1:      []Vote{VoteYes, VoteNo}, // 6 voting power
			votesT2:      []Vote{VoteYes, VoteYes},
			votesT3:      []Vote{},
			expectedEval: true,
			abstainT3:    true,
			threshold:    0.6,
			expectedYes:  7.0 / 10.0,
			panic:        false,
		},
		"T3 abstaining w/ votes": {
			votesT1:      []Vote{VoteYes, VoteNo}, // 6 voting power
			votesT2:      []Vote{VoteYes, VoteYes},
			votesT3:      []Vote{VoteYes, VoteNo},
			expectedEval: true,
			abstainT3:    true,
			threshold:    0.6,
			expectedYes:  7.0 / 10.0,
			panic:        true, // a user with T3 when T3 is abstaining should panic
		},
	}

	for name, tdata := range tests {
		t.Run(name, func(t *testing.T) {
			if tdata.panic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}
			votes := map[string]Vote{}
			dao := newMockDAO()
			for i, vote := range tdata.votesT1 {
				userID := ufmt.Sprintf("%d_T1", i)
				dao.addUser(userID, []string{"T1"})
				votes[userID] = vote
			}

			for i, vote := range tdata.votesT2 {
				userID := ufmt.Sprintf("%d_T2", i)
				dao.addUser(userID, []string{"T2"})
				votes[userID] = vote
			}

			for i, vote := range tdata.votesT3 {
				userID := ufmt.Sprintf("%d_T3", i)
				dao.addUser(userID, []string{"T3"})
				votes[userID] = vote
			}
			ballot := NewBallot()
			for userId, vote := range votes {
				ballot.Vote(userId, vote)
			}
			roles := []string{"T1", "T2"}
			if !tdata.abstainT3 {
				roles = append(roles, "T3")
			}
			cond := GnoloveDAOCondThreshold(tdata.threshold, roles, dao.hasRole, dao.usersWithRoleCount)
			// Get percent of total yes
			urequire.Equal(t, tdata.expectedEval, cond.Eval(ballot))
		})
	}
}

type mockDAO struct {
	members map[string][]string
	roles   map[string][]string
}

func newMockDAO() *mockDAO {
	return &mockDAO{
		members: map[string][]string{},
		roles:   map[string][]string{}, // roles to users
	}
}

func (m *mockDAO) addUser(memberId string, roles []string) {
	m.members[memberId] = roles
	for _, memberRole := range roles {
		m.roles[memberRole] = append(m.roles[memberRole], memberId)
	}
}
func (m *mockDAO) usersWithRoleCount(role string) uint32 {
	return uint32(len(m.roles[role]))
}

func (m *mockDAO) hasRole(memberId string, role string) bool {
	roles, ok := m.members[memberId]
	if !ok {
		return false
	}
	for _, memberRole := range roles {
		if memberRole == role {
			return true
		}
	}
	return false
}
