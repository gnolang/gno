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
		members []std.Address
	}{
		{
			name:    "with parent",
			parent:  New(),
			members: []std.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name:    "without parent",
			members: []std.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name: "multiple members",
			members: []std.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
		},
		{
			name: "no members",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			options := []Option{WithParent(tc.parent)}
			for _, m := range tc.members {
				options = append(options, WithMember(m))
			}

			dao := New(options...)

			if tc.parent == nil {
				uassert.Equal(t, nil, dao.Parent())
			} else {
				uassert.NotEqual(t, nil, dao.Parent())
			}

			urequire.Equal(t, len(tc.members), len(dao.Members()), "dao members")
			for i, m := range dao.Members() {
				uassert.Equal(t, tc.members[i], m)
			}
		})
	}
}

func TestAdminDAOAddMember(t *testing.T) {
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))

	err := dao.AddMember(member)
	urequire.NoError(t, err)
	uassert.Equal(t, 2, len(dao.Members()))
	uassert.True(t, dao.IsMember(member))

	err = dao.AddMember(member)
	uassert.ErrorIs(t, err, ErrMemberExists)
}

func TestAdminDAOIsMember(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		dao    *AdminDAO
		want   bool
	}{
		{
			name:   "member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
			want:   true,
		},
		{
			name:   "not a dao member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
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
