package boards

import (
	"std"
	"testing"

	"gno.land/p/demo/boards2/admindao"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestNewACL(t *testing.T) {
	roles := []string{"a", "b"}
	dao := admindao.New()

	acl := NewACL(dao, WithRole("a", "permission1"), WithRole("b", "permission2"))

	urequire.Equal(t, len(roles), len(acl.Roles()), "roles")
	for i, r := range acl.Roles() {
		uassert.Equal(t, roles[i], string(r))
	}
}

func TestACLWithPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		args       Args
		acl        *ACL
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
			called: true,
		},
		{
			name:       "ok with arguments",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			args:       Args{"a", "b"},
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
			called: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
				WithRole("foo", "bar"),
			),
			err: "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				called bool
				args   Args
			)

			callback := func(a Args) {
				args = a
				called = true
			}

			testCaseFn := func() {
				tc.acl.WithPermission(tc.user, tc.permission, tc.args, callback)
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, testCaseFn, "panic")
				return
			} else {
				urequire.NotPanics(t, testCaseFn, "no panic")
			}

			urequire.Equal(t, tc.called, called, "callback called")
			urequire.Equal(t, len(tc.args), len(args), "args count")
			for i, a := range args {
				uassert.Equal(t, tc.args[i].(string), a.(string))
			}
		})
	}
}

func TestACLGetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []string
		acl   *ACL
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			acl:   NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")),
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "foo", "bar"},
			acl:   NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo", "bar")),
		},
		{
			name: "without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			acl:  NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
		},
		{
			name: "not a user",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			acl:  NewACL(admindao.New()),
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin"),
				WithUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin"),
				WithUser("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roles := tc.acl.GetUserRoles(tc.user)

			urequire.Equal(t, len(tc.roles), len(roles), "user role count")
			for i, r := range roles {
				uassert.Equal(t, tc.roles[i], string(r))
			}
		})
	}
}

func TestACLHasRole(t *testing.T) {
	cases := []struct {
		name string
		user std.Address
		role Role
		acl  *ACL
		want bool
	}{
		{
			name: "ok",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "admin",
			acl:  NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")),
			want: true,
		},
		{
			name: "ok with multiple roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "foo",
			acl:  NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo")),
			want: true,
		},
		{
			name: "user without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			acl:  NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
		},
		{
			name: "has no role",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "bar",
			acl:  NewACL(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.acl.HasRole(tc.user, tc.role)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestACLHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		acl        *ACL
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
			want: true,
		},
		{
			name:       "ok with multiple users",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "foo"),
				WithRole("foo", "bar"),
			),
			want: true,
		},
		{
			name:       "ok with multiple roles",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo", "baz"),
				WithRole("foo", "bar"),
				WithRole("baz", "other"),
			),
			want: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			acl: NewACL(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.acl.HasPermission(tc.user, tc.permission)
			uassert.Equal(t, got, tc.want)
		})
	}
}
