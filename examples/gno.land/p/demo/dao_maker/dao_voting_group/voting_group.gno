package dao_voting_group

import (
	"std"
	"strconv"
	"strings"

	dao_interfaces "gno.land/p/demo/dao_maker/dao_interfaces"
	"gno.land/p/demo/dao_maker/havl"
	"gno.land/p/demo/dao_maker/jsonutil"
	"gno.land/p/demo/json"
)

type Member struct {
	Address std.Address
	Power   uint64
}

func (m Member) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"address": jsonutil.AddressNode(m.Address),
		"power":   jsonutil.Uint64Node(m.Power),
	})
}

func (m *Member) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	m.Address = jsonutil.MustAddress(obj["address"])
	m.Power = jsonutil.MustUint64(obj["power"])
}

type VotingGroup struct {
	dao_interfaces.IVotingModule

	powerByAddr *havl.Tree // std.Address -> uint64
	totalPower  *havl.Tree // "" -> uint64
	memberCount *havl.Tree // "" -> uint32
}

func NewVotingGroup() *VotingGroup {
	return &VotingGroup{
		powerByAddr: havl.NewTree(),
		totalPower:  havl.NewTree(),
		memberCount: havl.NewTree(),
	}
}

func (v *VotingGroup) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "gno.land/p/demo/dao_maker/dao_voting_group",
		Version: "0.1.0",
	}
}

func (v *VotingGroup) ConfigJSON() string {
	return json.ObjectNode("", map[string]*json.Node{
		"totalPower": jsonutil.Uint64Node(v.TotalPowerAtHeight(havl.Latest, []string{})),
		"members":    jsonutil.Uint32Node(v.MemberCount(havl.Latest)),
	}).String()
}

func (v *VotingGroup) GetMembersJSON(start, end string, limit uint64, height int64) string {
	members := v.GetMembers(start, end, limit, height)
	membersJSON := make([]*json.Node, len(members))
	for i, m := range members {
		membersJSON[i] = m.ToJSON()
	}
	return json.ArrayNode("", membersJSON).String()
}

func (v *VotingGroup) VotingPowerAtHeight(addr std.Address, height int64, resources []string) uint64 {
	_ = resources
	p, ok := v.powerByAddr.Get(addr.String(), height)
	if !ok {
		return 0
	}

	return p.(uint64)
}

func (v *VotingGroup) TotalPowerAtHeight(height int64, resources []string) uint64 {
	_ = resources
	p, ok := v.totalPower.Get("", height)
	if !ok {
		return 0
	}

	return p.(uint64)
}

func (g *VotingGroup) SetMemberPower(addr std.Address, power uint64) {
	if power == 0 {
		g.RemoveMember(addr)
		return
	}

	iprevious, ok := g.powerByAddr.Get(addr.String(), havl.Latest)
	if !ok {
		g.memberCount.Set("", g.MemberCount(havl.Latest)+1)
	}

	previous := uint64(0)
	if ok {
		previous = iprevious.(uint64)
	}

	if power == previous {
		return
	}

	g.powerByAddr.Set(addr.String(), power)

	ipreviousTotal, ok := g.totalPower.Get("", havl.Latest)
	previousTotal := uint64(0)
	if ok {
		previousTotal = ipreviousTotal.(uint64)
	}

	g.totalPower.Set("", (previousTotal+power)-previous)
}

func (g *VotingGroup) RemoveMember(addr std.Address) (uint64, bool) {
	p, removed := g.powerByAddr.Remove(addr.String())
	if !removed {
		return 0, false
	}

	g.memberCount.Set("", g.MemberCount(havl.Latest)-1)
	power := p.(uint64)
	g.totalPower.Set("", g.TotalPowerAtHeight(havl.Latest, []string{})-power)
	return power, true
}

func (g *VotingGroup) UpdateMembersHandler() dao_interfaces.MessageHandler {
	return &updateMembersHandler{vg: g}
}

func (g *VotingGroup) MemberCount(height int64) uint32 {
	val, ok := g.memberCount.Get("", height)
	if !ok {
		return 0
	}

	return val.(uint32)
}

func (g *VotingGroup) GetMembers(start, end string, limit uint64, height int64) []Member {
	var members []Member
	g.powerByAddr.Iterate(start, end, height, func(k string, v interface{}) bool {
		if limit > 0 && uint64(len(members)) >= limit {
			return true
		}

		members = append(members, Member{
			Address: std.Address(k),
			Power:   v.(uint64),
		})

		return false
	})
	return members
}

func (v *VotingGroup) Render(path string) string {
	sb := strings.Builder{}
	sb.WriteString("Member count: ")
	sb.WriteString(strconv.FormatUint(uint64(v.MemberCount(havl.Latest)), 10))
	sb.WriteString("\n\n")
	sb.WriteString("Total power: ")
	sb.WriteString(strconv.FormatUint(v.TotalPowerAtHeight(havl.Latest, []string{}), 10))
	sb.WriteString("\n\n")
	sb.WriteString("Members:\n")
	v.powerByAddr.Iterate("", "", havl.Latest, func(k string, v interface{}) bool {
		sb.WriteString("- ")
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(strconv.FormatUint(v.(uint64), 10))
		sb.WriteRune('\n')
		return false
	})

	sb.WriteRune('\n')
	return sb.String()
}
