package admindao

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		parent  *AdminDAO
		roles   []string
		members []Member
	}{
		{
			name:   "with parent",
			parent: New(),
			roles:  []string{"admin", "moderator"},
			members: []Member{
				{
					Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
					Roles:   []Role{"admin"},
				},
				{
					Address: std.Address("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
					Roles:   []Role{"moderator"},
				},
			},
		},
		{
			name:  "without parent",
			roles: []string{"admin", "moderator"},
			members: []Member{
				{
					Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
					Roles:   []Role{"admin"},
				},
				{
					Address: std.Address("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
					Roles:   []Role{"moderator"},
				},
			},
		},
		{
			name:  "single role",
			roles: []string{"admin"},
			members: []Member{
				{
					Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
					Roles:   []Role{"admin"},
				},
				{
					Address: std.Address("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
					Roles:   []Role{"admin"},
				},
			},
		},
		{
			name:  "many roles",
			roles: []string{"admin", "bar", "foo", "moderator"},
			members: []Member{
				{
					Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
					Roles:   []Role{"admin", "foo"},
				},
				{
					Address: std.Address("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
					Roles:   []Role{"moderator", "bar"},
				},
			},
		},
		{
			name:  "duplicated roles",
			roles: []string{"admin", "moderator"},
			members: []Member{
				{
					Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
					Roles:   []Role{"admin"},
				},
				{
					Address: std.Address("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
					Roles:   []Role{"moderator", "admin"},
				},
			},
		},
		{
			name: "no roles",
			members: []Member{
				{Address: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var options []Option
			if tc.parent != nil {
				options = append(options, WithParent(tc.parent))
			}
			for _, m := range tc.members {
				options = append(options, WithMember(m.Address, m.Roles...))
			}

			dao := New(options...)

			if tc.parent == nil {
				uassert.Equal(t, nil, dao.Parent())
			} else {
				uassert.NotEqual(t, nil, dao.Parent())
			}

			urequire.Equal(t, len(tc.roles), len(dao.Roles()), "dao roles")
			for i, r := range dao.Roles() {
				uassert.Equal(t, tc.roles[i], string(r))
			}

			urequire.Equal(t, len(tc.members), len(dao.Members()), "dao members")
			for i, m := range dao.Members() {
				uassert.Equal(t, tc.members[i].Address, m.Address)
				urequire.Equal(t, len(tc.members[i].Roles), len(m.Roles), "member roles")
				for j, r := range m.Roles {
					uassert.Equal(t, string(tc.members[i].Roles[j]), string(r))
				}
			}
		})
	}
}

func TestAdminDAOGetMemberRoles(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		roles  []string
		dao    *AdminDAO
	}{
		{
			name:   "single role",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			roles:  []string{"admin"},
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin")),
		},
		{
			name:   "multiple roles",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			roles:  []string{"admin", "foo", "bar"},
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin", "foo", "bar")),
		},
		{
			name:   "member without roles",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
		},
		{
			name:   "not a dao member",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			dao:    New(),
		},
		{
			name:   "multiple members",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			roles:  []string{"admin"},
			dao: New(
				WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin"),
				WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roles := tc.dao.GetMemberRoles(tc.member)

			urequire.Equal(t, len(tc.roles), len(roles), "member role count")
			for i, r := range roles {
				uassert.Equal(t, tc.roles[i], string(r))
			}
		})
	}
}

func TestAdminIsMember(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		dao    *AdminDAO
		want   bool
	}{
		{
			name:   "member",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
			want:   true,
		},
		{
			name:   "not a dao member",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			dao:    New(WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.dao.IsMember(tc.member)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestAdminDAOHasRole(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		role   Role
		dao    *AdminDAO
		want   bool
	}{
		{
			name:   "ok",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			role:   "admin",
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin")),
			want:   true,
		},
		{
			name:   "ok with multiple roles",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			role:   "foo",
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin", "foo")),
			want:   true,
		},
		{
			name:   "member without roles",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
		},
		{
			name:   "has no role",
			member: std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
			role:   "bar",
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "foo")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.dao.HasRole(tc.member, tc.role)
			uassert.Equal(t, got, tc.want)
		})
	}
}
