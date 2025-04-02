package boards2

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/gnoland/commondao"
)

var _ Permissions = (*DefaultPermissions)(nil)

func TestDefaultPermissionsWithPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		args       Args
		setup      func() *DefaultPermissions
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "ok with arguments",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			args:       Args{"a", "b"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			err: "unauthorized",
		},
		{
			name:       "is not a DAO member",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(commondao.New())
			},
			err: "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				called bool
				args   Args
			)

			perms := tc.setup()
			callback := func(a Args) {
				args = a
				called = true
			}

			testCaseFn := func() {
				perms.WithPermission(tc.user, tc.permission, tc.args, callback)
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
		setup func() *DefaultPermissions
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "foo", "bar"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("foo", "x")
				perms.AddRole("bar", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo", "bar")
				return perms
			},
		},
		{
			name: "without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "not a user",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(commondao.New())
			},
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("bar", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				perms.AddUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin")
				perms.AddUser("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar")
				return perms
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			roles := perms.GetUserRoles(tc.user)

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
		setup func() *DefaultPermissions
		want  bool
	}{
		{
			name: "ok",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "admin",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
			want: true,
		},
		{
			name: "ok with multiple roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "foo",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("foo", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo")
				return perms
			},
			want: true,
		},
		{
			name: "user without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "has no role",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "bar",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "x")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			got := perms.HasRole(tc.user, tc.role)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestDefaultPermissionsHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		setup      func() *DefaultPermissions
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple users",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				perms.AddUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple roles",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddRole("baz", "other")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo", "baz")
				return perms
			},
			want: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			got := perms.HasPermission(tc.user, tc.permission)
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
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				return perms
			},
		},
		{
			name:  "multiple users",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddUser("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "a")
				perms.AddUser("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")
				return perms
			},
		},
		{
			name: "duplicated user",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			err: "user already exists",
		},
		{
			name:  "duplicated user",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "foo"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				return perms
			},
			err: "invalid role: foo",
		},
		{
			name: "already a DAO member",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(
					commondao.New(commondao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")),
				)
			},
			err: "user already exists",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()

			err := perms.AddUser(tc.user, tc.roles...)

			if tc.err != "" {
				urequire.Error(t, err, "expected an error")
				uassert.Equal(t, tc.err, err.Error())
				return
			} else {
				urequire.NoError(t, err)
			}

			roles := perms.GetUserRoles(tc.user)
			uassert.Equal(t, len(tc.roles), len(roles))
			for i, r := range roles {
				urequire.Equal(t, string(tc.roles[i]), string(r))
			}
		})
	}
}

func TestDefaultPermissionsSetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []Role
		setup func() *DefaultPermissions
		err   string
	}{
		{
			name:  "single role",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"b"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
		},
		{
			name:  "multiple roles",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"b", "c"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.AddRole("c", "permission2")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
		},
		{
			name:  "duplicated role",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "c"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.AddRole("c", "permission2")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a", "c")
				return perms
			},
		},
		{
			name: "remove roles",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a", "b")
				return perms
			},
		},
		{
			name:  "invalid role",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"x", "a"},
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
			err: "invalid role: x",
		},
		{
			name: "user not found",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(commondao.New())
			},
			err: "user not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()

			err := perms.SetUserRoles(tc.user, tc.roles...)

			if tc.err != "" {
				urequire.True(t, err != nil, "expected an error")
				uassert.Equal(t, tc.err, err.Error())
				return
			} else {
				urequire.NoError(t, err)
			}

			roles := perms.GetUserRoles(tc.user)
			uassert.Equal(t, len(tc.roles), len(roles))
			for i, r := range roles {
				urequire.Equal(t, string(tc.roles[i]), string(r))
			}
		})
	}
}

func TestDefaultPermissionsRemoveUser(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		setup func() *DefaultPermissions
		want  bool
	}{
		{
			name: "ok",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				perms := NewDefaultPermissions(commondao.New())
				perms.AddUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			want: true,
		},
		{
			name: "user not found",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *DefaultPermissions {
				return NewDefaultPermissions(commondao.New())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			got := perms.RemoveUser(tc.user)
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestDefaultPermissionsIterateUsers(t *testing.T) {
	users := []User{
		{
			Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			Roles:   []Role{"foo"},
		},
		{
			Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
			Roles:   []Role{"foo", "bar"},
		},
		{
			Address: "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5",
			Roles:   []Role{"bar"},
		},
	}

	perms := NewDefaultPermissions(commondao.New())
	perms.AddRole("foo", "perm1")
	perms.AddRole("bar", "perm2")
	for _, u := range users {
		perms.AddUser(u.Address, u.Roles...)
	}

	cases := []struct {
		name               string
		start, count, want int
	}{
		{
			name:  "exceed users count",
			count: 50,
			want:  3,
		},
		{
			name:  "exact users count",
			count: 3,
			want:  3,
		},
		{
			name:  "two users",
			start: 1,
			count: 2,
			want:  2,
		},
		{
			name:  "one user",
			start: 1,
			count: 1,
			want:  1,
		},
		{
			name:  "no iteration",
			start: 50,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var i int
			perms.IterateUsers(0, len(users), func(u User) bool {
				urequire.True(t, i < len(users), "expect iterator to respect number of users")
				uassert.Equal(t, users[i].Address, u.Address)

				urequire.Equal(t, len(users[i].Roles), len(u.Roles), "expect number of roles to match")
				for j, r := range u.Roles {
					uassert.Equal(t, string(users[i].Roles[j]), string(u.Roles[j]))
				}

				i++
			})

			uassert.Equal(t, i, len(users), "expect iterator to iterate all users")
		})
	}
}
