package boards2

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var _ boards.Permissions = (*BasicPermissions)(nil)

func TestBasicPermissionsWithPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       address
		permission boards.Permission
		args       boards.Args
		setup      func() *BasicPermissions
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
			args:       boards.Args{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
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
				return NewBasicPermissions()
			},
			err: "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var called bool

			perms := tc.setup()
			callback := func(realm) { called = true }

			testCaseFn := func() {
				perms.WithPermission(cross, tc.user, tc.permission, tc.args, callback)
			}

			if tc.err != "" {
				urequire.AbortsWithMessage(t, tc.err, testCaseFn, "expect panic with message")
				return
			} else {
				urequire.NotAborts(t, testCaseFn, "expect no panic")
			}

			urequire.Equal(t, tc.called, called, "expect callback to be called")
		})
	}
}

func TestBasicPermissionsSetPublicPermissions(t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	perms := NewBasicPermissions()

	// Add a new role with permissions
	perms.AddRole("adminRole", "fooPerm", "barPerm", "bazPerm")
	urequire.False(t, perms.HasPermission(user, "fooPerm"))
	urequire.False(t, perms.HasPermission(user, "barPerm"))
	urequire.False(t, perms.HasPermission(user, "bazPerm"))

	// Assign a couple of public permissions
	perms.SetPublicPermissions("fooPerm", "bazPerm")
	urequire.True(t, perms.HasPermission(user, "fooPerm"))
	urequire.False(t, perms.HasPermission(user, "barPerm"))
	urequire.True(t, perms.HasPermission(user, "bazPerm"))

	// Clear all public permissions
	perms.SetPublicPermissions()
	urequire.False(t, perms.HasPermission(user, "fooPerm"))
	urequire.False(t, perms.HasPermission(user, "barPerm"))
	urequire.False(t, perms.HasPermission(user, "bazPerm"))
}

func TestBasicPermissionsGetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  address
		roles []string
		setup func() *BasicPermissions
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("admin", "x")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "bar", "foo"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "not a user",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *BasicPermissions {
				return NewBasicPermissions()
			},
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
		user  address
		role  boards.Role
		setup func() *BasicPermissions
		want  bool
	}{
		{
			name: "ok",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "admin",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "has no role",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
		user       address
		permission boards.Permission
		setup      func() *BasicPermissions
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: "bar",
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
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
				perms := NewBasicPermissions()
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
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestBasicPermissionsSetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  address
		roles []boards.Role
		setup func() *BasicPermissions
		err   string
	}{
		{
			name:  "add user",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{"a"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				return perms
			},
		},
		{
			name:  "add user with multiple roles",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				return perms
			},
		},
		{
			name:  "add when other users exists",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{"a"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				perms.SetUserRoles(cross, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "a")
				perms.SetUserRoles(cross, "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")
				return perms
			},
		},
		{
			name:  "update user roles",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{"a", "b"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
		},
		{
			name:  "clear user roles",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				perms.AddRole("b", "permission2")
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a", "b")
				return perms
			},
		},
		{
			name:  "set invalid role",
			user:  address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			roles: []boards.Role{"a", "foo"},
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.AddRole("a", "permission1")
				return perms
			},
			err: "invalid role: foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()

			setUserRoles := func() {
				perms.SetUserRoles(cross, tc.user, tc.roles...)
			}

			if tc.err != "" {
				urequire.AbortsWithMessage(t, tc.err, setUserRoles, "expected an error")
				return
			} else {
				urequire.NotAborts(t, setUserRoles, "expected no error")
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
		user  address
		setup func() *BasicPermissions
		want  bool
	}{
		{
			name: "ok",
			user: address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *BasicPermissions {
				perms := NewBasicPermissions()
				perms.SetUserRoles(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			want: true,
		},
		{
			name: "user not found",
			user: address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *BasicPermissions {
				return NewBasicPermissions()
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
	users := []boards.User{
		{
			Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			Roles:   []boards.Role{"foo"},
		},
		{
			Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
			Roles:   []boards.Role{"bar", "foo"},
		},
		{
			Address: "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5",
			Roles:   []boards.Role{"bar"},
		},
	}

	perms := NewBasicPermissions()
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
			perms.IterateUsers(0, len(users), func(u boards.User) bool {
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
