package basedao

import (
	"chain"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/onbloc/json"
)

type MembersStore struct {
	Roles   *avl.Tree // role name -> *Role
	Members *avl.Tree // string -> *avl.Tree [roles -> struct{}]
}

type Member struct {
	Address string
	Roles   []string
}

type RoleInfo struct {
	Name        string
	Description string
}

type Role struct {
	Name        string
	Description string
	Members     *avl.Tree // string -> struct{}
}

const EventAddMember = "BaseDAOAddMember"
const EventRemoveMember = "BaseDAORemoveMember"

func NewMembersStore(initialRoles []RoleInfo, initialMembers []Member) *MembersStore {
	res := &MembersStore{
		Roles:   avl.NewTree(),
		Members: avl.NewTree(),
	}
	res.setRoles(initialRoles)
	res.setMembers(initialMembers)
	return res
}

func (m *MembersStore) HasRole(member string, role string) bool {
	rolesRaw, ok := m.Members.Get(member)
	if !ok {
		return false
	}
	roles, ok := rolesRaw.(*avl.Tree)
	if !ok {
		return false
	}
	return roles.Has(role)
}

func (m *MembersStore) IsMember(member string) bool {
	return m.Members.Has(member)
}

func (m *MembersStore) RoleInfo(role string) RoleInfo {
	roleDataRaw, ok := m.Roles.Get(role)
	if !ok {
		panic("role does not exist")
	}
	roleData, ok := roleDataRaw.(*Role)
	if !ok {
		panic("a value of memberstore.roles is not a Role, should not happen")
	}
	return RoleInfo{
		Name:        roleData.Name,
		Description: roleData.Description,
	}
}

func (m *MembersStore) MembersCount() uint64 {
	return uint64(m.Members.Size())
}

func (m *MembersStore) GetMembers() []string {
	members := make([]string, 0, m.Members.Size())
	m.Members.Iterate("", "", func(key string, value interface{}) bool {
		members = append(members, key)
		return false
	})
	return members
}

func (m *MembersStore) GetRoles() []string {
	i := 0
	res := make([]string, m.Roles.Size())
	m.Roles.Iterate("", "", func(key string, value interface{}) bool {
		res[i] = key
		i++
		return false
	})
	return res
}

func (m *MembersStore) GetMemberRoles(member string) []string {
	rolesRaw, ok := m.Members.Get(member)
	if !ok {
		return []string{}
	}
	roles, ok := rolesRaw.(*avl.Tree)
	if !ok {
		return []string{}
	}
	i := 0
	res := make([]string, roles.Size())
	roles.Iterate("", "", func(key string, value interface{}) bool {
		res[i] = key
		i++
		return false
	})
	return res
}

func (m *MembersStore) CountMemberRoles(member string) int {
	rolesRaw, ok := m.Members.Get(member)
	if !ok {
		return 0
	}
	roles, ok := rolesRaw.(*avl.Tree)
	if !ok {
		return 0
	}
	return roles.Size()
}

func (m *MembersStore) GetMembersWithRole(role string) []string {
	roleDataRaw, ok := m.Roles.Get(role)
	if !ok {
		return []string{}
	}
	roleData, ok := roleDataRaw.(*Role)
	if !ok {
		return []string{}
	}
	i := 0
	res := make([]string, roleData.Members.Size())
	roleData.Members.Iterate("", "", func(key string, value interface{}) bool {
		res[i] = key
		i++
		return false
	})
	return res
}

func (m *MembersStore) CountMembersWithRole(role string) uint32 {
	return uint32(len(m.GetMembersWithRole(role)))
}

func (m *MembersStore) setRoles(roles []RoleInfo) {
	for _, role := range roles {
		m.AddRole(role)
	}
}

func (m *MembersStore) setMembers(members []Member) {
	for _, member := range members {
		m.AddMember(member.Address, member.Roles)
	}
}

func (m *MembersStore) AddMember(member string, roles []string) {
	if m.IsMember(member) {
		panic("member already exists")
	}

	membersRoles := avl.NewTree()
	for _, role := range roles {
		if !m.Roles.Has(role) {
			panic("role: " + role + " does not exist")
		}
		membersRoles.Set(role, struct{}{})
	}
	m.Members.Set(member, membersRoles)

	chain.Emit(EventAddMember,
		"address", member,
	)
}

func (m *MembersStore) RemoveMember(member string) {
	if !m.IsMember(member) {
		panic("member does not exist")
	}

	memberRolesRaw, ok := m.Members.Get(member)
	if !ok {
		panic("should not happen")
	}
	memberRoles, ok := memberRolesRaw.(*avl.Tree)
	if !ok {
		panic("a value of memberstore.members is not an avl.Tree, should not happen")
	}

	memberRoles.Iterate("", "", func(key string, value interface{}) bool {
		roleRaw, ok := m.Roles.Get(key)
		if !ok {
			return false
		}
		role, ok := roleRaw.(*Role)
		if !ok {
			panic("a value of memberstore.roles is not a Role, should not happen")
		}
		role.Members.Remove(role.Name)
		return false
	})
	m.Members.Remove(member)

	chain.Emit(EventRemoveMember,
		"address", member,
	)
}

func (m *MembersStore) AddRole(role RoleInfo) {
	if m.Roles.Has(role.Name) {
		panic("role already exists")
	}

	roleData := &Role{
		Name:        role.Name,
		Description: role.Description,
		Members:     avl.NewTree(),
	}

	m.Roles.Set(role.Name, roleData)
}

func (m *MembersStore) RemoveRole(role string) {
	roleDataRaw, ok := m.Roles.Get(role)
	if !ok {
		panic("role does not exist")
	}
	roleData, ok := roleDataRaw.(*Role)
	if !ok {
		panic("a value of memberstore.roles is not a Role, should not happen")
	}

	roleData.Members.Iterate("", "", func(key string, value interface{}) bool {
		memberRaw, ok := m.Members.Get(key)
		if !ok {
			return false
		}
		member, ok := memberRaw.(*avl.Tree)
		if !ok {
			panic("a value of memberstore.members is not an avl.Tree, should not happen")
		}
		member.Remove(role)
		return false
	})
	m.Roles.Remove(role)
}

func (m *MembersStore) AddRoleToMember(member string, role string) {
	if !m.IsMember(member) {
		panic("member does not exist")
	}
	if !m.Roles.Has(role) {
		panic("role " + role + " does not exist")
	}
	if m.HasRole(member, role) {
		panic("member already has the role")
	}

	memberRolesRaw, ok := m.Members.Get(member)
	if !ok {
		panic("should not happen")
	}
	memberRoles, ok := memberRolesRaw.(*avl.Tree)
	if !ok {
		panic("a value of memberstore.members is not an avl.Tree, should not happen")
	}

	roleDataRaw, ok := m.Roles.Get(role)
	if !ok {
		panic("should not happen")
	}
	roleData, ok := roleDataRaw.(*Role)
	if !ok {
		panic("a value of memberstore.roles is not a Role, should not happen")
	}

	roleData.Members.Set(member, struct{}{})
	memberRoles.Set(role, struct{}{})
}

func (m *MembersStore) RemoveRoleFromMember(member string, role string) {
	if !m.IsMember(member) {
		panic("member does not exist")
	}
	if !m.Roles.Has(role) {
		panic("role " + role + " does not exist")
	}
	if !m.HasRole(member, role) {
		panic("member does not have the role")
	}

	memberRolesRaw, ok := m.Members.Get(member)
	if !ok {
		panic("should not happen")
	}
	memberRoles, ok := memberRolesRaw.(*avl.Tree)
	if !ok {
		panic("a value of memberstore.members is not an avl.Tree, should not happen")
	}

	roleDataRaw, ok := m.Roles.Get(role)
	if !ok {
		panic("should not happen")
	}
	roleData, ok := roleDataRaw.(*Role)
	if !ok {
		panic("a value of memberstore.roles is not a Role, should not happen")
	}

	memberRoles.Remove(role)
	roleData.Members.Remove(member)
}

func (m *MembersStore) GetMembersJSON() string {
	// XXX: replace with protoc-gen-gno
	members := []*json.Node{}
	for _, memberID := range m.GetMembers() {
		roles := []*json.Node{}
		for _, role := range m.GetMemberRoles(memberID) {
			roles = append(roles, json.StringNode("", role))
		}
		members = append(members, json.ObjectNode("", map[string]*json.Node{
			"address": json.StringNode("", memberID),
			"roles":   json.ArrayNode("", roles),
		}))
	}
	node := json.ArrayNode("", members)
	bz, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	return string(bz)
}
