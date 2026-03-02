package daocond_test

import (
	"errors"
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/p/samcrew/daocond"
)

func TestCondition(t *testing.T) {
	dao := newMockDAO()

	// leaf conditions
	membersMajority := daocond.MembersThreshold(0.6, dao.isMember, dao.membersCount)
	publicRelationships := daocond.RoleCount(1, "public-relationships", dao.hasRole)
	financeOfficer := daocond.RoleCount(1, "finance-officer", dao.hasRole)

	urequire.Equal(t, "60% of members", membersMajority.Render())
	urequire.Equal(t, "1 public-relationships", publicRelationships.Render())
	urequire.Equal(t, "1 finance-officer", financeOfficer.Render())

	// ressource expressions
	ressources := map[string]daocond.Condition{
		"social.post":    daocond.And(publicRelationships, membersMajority),
		"finance.invest": daocond.Or(financeOfficer, membersMajority),
	}

	urequire.Equal(t, "[1 public-relationships AND 60% of members]", ressources["social.post"].Render())
	urequire.Equal(t, "[1 finance-officer OR 60% of members]", ressources["finance.invest"].Render())
}

func TestEval(t *testing.T) {
	setups := []struct {
		name  string
		setup func(dao *mockDAO)
	}{
		{name: "basic", setup: func(dao *mockDAO) {
			membersMajority := daocond.MembersThreshold(0.6, dao.isMember, dao.membersCount)
			publicRelationships := daocond.RoleCount(1, "public-relationships", dao.hasRole)
			financeOfficer := daocond.RoleCount(1, "finance-officer", dao.hasRole)
			dao.resources = map[string]daocond.Condition{
				"social.post":    daocond.And(publicRelationships, membersMajority),
				"finance.invest": daocond.Or(financeOfficer, membersMajority),
			}
		}},
	}

	cases := []struct {
		name     string
		resource string
		phases   []testPhase
	}{
		{
			name:     "post with public-relationships",
			resource: "social.post",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "yes",
					"bob":   "yes",
					"eve":   "no",
				},
				result: true,
			}},
		},
		{
			name:     "post without public-relationships",
			resource: "social.post",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "yes",
					"bob":   "no",
					"eve":   "yes",
				},
				result: false,
			}},
		},
		{
			name:     "post after public-relationships changes",
			resource: "social.post",
			phases: []testPhase{
				{
					votes: map[string]daocond.Vote{
						"alice": "yes",
						"bob":   "yes",
						"eve":   "no",
					},
					result: true,
				},
				{
					changes: func(dao *mockDAO) {
						dao.unassignRole("bob", "public-relationships")
					},
					result: false,
				},
				{
					changes: func(dao *mockDAO) {
						dao.assignRole("alice", "public-relationships")
					},
					result: true,
				},
			},
		},
		{
			name:     "post public-relationships alone",
			resource: "social.post",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "no",
					"bob":   "yes",
					"eve":   "no",
				},
				result: false,
			}},
		},
		{
			name:     "invest with finance officer",
			resource: "finance.invest",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "yes",
					"bob":   "no",
					"eve":   "no",
				},
				result: true,
			}},
		},
		{
			name:     "invest without finance officer",
			resource: "finance.invest",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "no",
					"bob":   "yes",
					"eve":   "yes",
				},
				result: true,
			}},
		},
		{
			name:     "invest alone",
			resource: "finance.invest",
			phases: []testPhase{{
				votes: map[string]daocond.Vote{
					"alice": "no",
					"bob":   "no",
					"eve":   "yes",
				},
				result: false,
			}},
		},
	}

	for _, tc := range cases {
		for _, s := range setups {
			t.Run(tc.name+" "+s.name, func(t *testing.T) {
				dao := newMockDAO()
				s.setup(dao)

				resource, ok := dao.resources[tc.resource]
				urequire.True(t, ok)

				ballot := daocond.NewBallot()
				for _, phase := range tc.phases {
					if phase.changes != nil {
						phase.changes(dao)
					}
					if phase.votes != nil {
						for memberId, vote := range phase.votes {
							ballot.Vote(memberId, vote)
						}
					}
					result := resource.Eval(ballot)
					if phase.result != result {
						println("State:", resource.RenderWithVotes(ballot))
					}
					urequire.Equal(t, phase.result, result)
				}
			})
		}
	}
}

type testPhase struct {
	changes func(dao *mockDAO)
	votes   map[string]daocond.Vote
	result  bool
}

type mockDAO struct {
	members   map[string][]string
	roles     map[string][]string
	resources map[string]daocond.Condition
}

func newMockDAO() *mockDAO {
	return &mockDAO{
		members: map[string][]string{
			"alice": []string{"finance-officer"},
			"bob":   []string{"public-relationships"},
			"eve":   []string{},
		},
		roles: map[string][]string{
			"finance-officer":      []string{"alice"},
			"public-relationships": []string{"bob"},
		}, // roles to users
		resources: make(map[string]daocond.Condition),
	}
}

func (m *mockDAO) assignRole(userId string, role string) {
	roles, ok := m.members[userId]
	if !ok {
		panic(errors.New("unknown member"))
	}
	m.members[userId], ok = strsadd(roles, role)
}

func (m *mockDAO) unassignRole(userId string, role string) {
	roles, ok := m.members[userId]
	if !ok {
		panic(errors.New("unknown member"))
	}
	m.members[userId], ok = strsrm(roles, role)
}

func (m *mockDAO) isMember(memberId string) bool {
	_, ok := m.members[memberId]
	return ok
}

func (m *mockDAO) membersCount() uint64 {
	return uint64(len(m.members))
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

func strsrm(strs []string, val string) ([]string, bool) {
	removed := false
	res := []string{}
	for _, str := range strs {
		if str == val {
			removed = true
			continue
		}
		res = append(res, str)
	}
	return res, removed
}

func strsadd(strs []string, val string) ([]string, bool) {
	for _, str := range strs {
		if str == val {
			return strs, false
		}
	}
	return append(strs, val), true
}
