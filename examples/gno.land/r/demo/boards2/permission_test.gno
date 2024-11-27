package boards2

import (
	"std"
	"testing"

	"gno.land/p/demo/boards2/admindao"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var _ Permissioner = (*DefaultPermissions)(nil)

func TestNewDefaultPermissions(t *testing.T) {
	roles := []Role{"a", "b"}
	dao := admindao.New()

	perms := NewDefaultPermissions(dao, WithRole("a", "permission1"), WithRole("b", "permission2"))

	urequire.Equal(t, len(roles), len(perms.Roles()), "roles")
	for i, r := range perms.Roles() {
		uassert.Equal(t, string(roles[i]), string(r))
	}

	for _, r := range roles {
		uassert.True(t, perms.RoleExists(r))
	}
}

func TestDefaultPermissionsWithPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		args       Args
		perms      *DefaultPermissions
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			perms: NewDefaultPermissions(
				admindao.New(admindao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
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
			perms: NewDefaultPermissions(
				admindao.New(admindao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
			called: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			perms: NewDefaultPermissions(
				admindao.New(admindao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
				WithRole("foo", "bar"),
			),
			err: "unauthorized",
		},
		{
			name:       "is not a DAO member",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			perms: NewDefaultPermissions(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
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
				tc.perms.WithPermission(tc.user, tc.permission, tc.args, callback)
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

func TestDefaultPermissionsGetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []string
		perms *DefaultPermissions
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")),
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "foo", "bar"},
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo", "bar")),
		},
		{
			name:  "without roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
		},
		{
			name:  "not a user",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			perms: NewDefaultPermissions(admindao.New()),
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			perms: NewDefaultPermissions(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin"),
				WithUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin"),
				WithUser("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roles := tc.perms.GetUserRoles(tc.user)

			urequire.Equal(t, len(tc.roles), len(roles), "user role count")
			for i, r := range roles {
				uassert.Equal(t, tc.roles[i], string(r))
			}
		})
	}
}

func TestDefaultPermissionsHasRole(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		role  Role
		perms *DefaultPermissions
		want  bool
	}{
		{
			name:  "ok",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role:  "admin",
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")),
			want:  true,
		},
		{
			name:  "ok with multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role:  "foo",
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo")),
			want:  true,
		},
		{
			name:  "user without roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
		},
		{
			name:  "has no role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role:  "bar",
			perms: NewDefaultPermissions(admindao.New(), WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.perms.HasRole(tc.user, tc.role)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestDefaultPermissionsHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		perms      *DefaultPermissions
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			perms: NewDefaultPermissions(
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
			perms: NewDefaultPermissions(
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
			perms: NewDefaultPermissions(
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
			perms: NewDefaultPermissions(
				admindao.New(),
				WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo"),
				WithRole("foo", "bar"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.perms.HasPermission(tc.user, tc.permission)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestDefaultPermissionsAddUser(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []Role
		setup func() *DefaultPermissions
		err   string
	}{
		{
			name:  "single user",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "b"},
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(
					admindao.New(),
					WithRole("a", "permission1"),
					WithRole("b", "permission2"),
				)
			},
		},
		{
			name:  "multiple users",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a"},
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(
					admindao.New(),
					WithRole("a", "permission1"),
					WithUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "a"),
					WithUser("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
				)
			},
		},
		{
			name: "duplicated user",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(
					admindao.New(admindao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
					WithUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
				)
			},
			err: "user already exists",
		},
		{
			name:  "duplicated user",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "foo"},
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(admindao.New(), WithRole("a", "permission1"))
			},
			err: "invalid role: foo",
		},
		{
			name: "already a DAO member",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(
					admindao.New(admindao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
				)
			},
			err: "member already exist",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perm := tc.setup()

			err := perm.AddUser(tc.user, tc.roles...)

			if tc.err != "" {
				urequire.True(t, err != nil, "expected an error")
				uassert.Equal(t, tc.err, err.Error())
				return
			} else {
				urequire.NoError(t, err)
			}

			roles := perm.GetUserRoles(tc.user)
			uassert.Equal(t, len(tc.roles), len(roles))
			for i, r := range roles {
				urequire.Equal(t, string(tc.roles[i]), string(r))
			}
		})
	}
}
