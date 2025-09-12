package boards2

import (
	"std"
	"testing"

	"gno.land/p/nt/commondao"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var _ Permissions = (*BasicPermissions)(nil)

func TestBasicPermissionsWithPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		args       Args
		setup      func() *BasicPermissions
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "ok with arguments",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			args:       Args{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			err: "unauthorized",
		},
		{
			name:       "is not a DAO member",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				return NewBasicPermissions(commondao.New())
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
			callback := func(_ realm, a Args) {
				args = a
				called = true
			}

			testCaseFn := func() {
				perms.WithPermission(cross, tc.user, tc.permission, tc.args, callback)
			}

			if tc.err != "" {
				urequire.AbortsWithMessage(t, tc.err, testCaseFn, "panic")
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

func TestBasicPermissionsGetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []string
		setup func() *BasicPermissions
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "foo", "bar"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("foo", "x")
				perms.AddRole("bar", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo", "bar")
				return perms
			},
		},
		{
			name: "without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "not a user",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *BasicPermissions {
				return NewBasicPermissions(commondao.New())
			},
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("bar", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				perms.SetUserRoles(cross, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin")
				perms.SetUserRoles(cross, "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar")
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

func TestBasicPermissionsHasRole(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		role  Role
		setup func() *BasicPermissions
		want  bool
	}{
		{
			name: "ok",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "admin",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
			want: true,
		},
		{
			name: "ok with multiple roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "foo",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("admin", "x")
				perms.AddRole("foo", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo")
				return perms
			},
			want: true,
		},
		{
			name: "user without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "has no role",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
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

func TestBasicPermissionsHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       std.Address
		permission Permission
		setup      func() *BasicPermissions
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple users",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				perms.SetUserRoles(cross, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple roles",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.AddRole("baz", "other")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo", "baz")
				return perms
			},
			want: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "other",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("foo", "bar")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
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

func TestBasicPermissionsSetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		roles []Role
		setup func() *BasicPermissions
		err   string
	}{
		{
			name:  "add user",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				return perms
			},
		},
		{
			name:  "add user with multiple roles",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				return perms
			},
		},
		{
			name:  "add when other users exists",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.SetUserRoles(cross, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "a")
				perms.SetUserRoles(cross, "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")
				return perms
			},
		},
		{
			name:  "update user roles",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
		},
		{
			name:  "clear user roles",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a", "b")
				return perms
			},
		},
		{
			name:  "set invalid role",
			user:  std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []Role{"a", "foo"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.AddRole("a", "permission1")
				return perms
			},
			err: "invalid role: foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()

			err := perms.SetUserRoles(cross, tc.user, tc.roles...)

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

func TestBasicPermissionsRemoveUser(t *testing.T) {
	cases := []struct {
		name  string
		user  std.Address
		setup func() *BasicPermissions
		want  bool
	}{
		{
			name: "ok",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions(commondao.New())
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			want: true,
		},
		{
			name: "user not found",
			user: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *BasicPermissions {
				return NewBasicPermissions(commondao.New())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			got := perms.RemoveUser(cross, tc.user)
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestBasicPermissionsIterateUsers(t *testing.T) {
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

	perms := NewBasicPermissions(commondao.New())
	perms.AddRole("foo", "perm1")
	perms.AddRole("bar", "perm2")
	for _, u := range users {
		perms.SetUserRoles(cross, u.Address, u.Roles...)
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
				for j := range u.Roles {
					uassert.Equal(t, string(users[i].Roles[j]), string(u.Roles[j]))
				}

				i++
				return false
			})

			uassert.Equal(t, i, len(users), "expect iterator to iterate all users")
		})
	}
}
