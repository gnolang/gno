package permissions

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// Test permission constants
const (
	testPermA boards.Permission = iota
	testPermB
	testPermC
)

var _ boards.Permissions = (*Permissions)(nil)

func TestBasicPermissionsWithPermission(cur realm, t *testing.T) {
	cases := []struct {
		name       string
		user       address
		permission boards.Permission
		args       boards.Args
		setup      func() *Permissions
		err        string
		called     bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "ok with arguments",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			args:       boards.Args{"a", "b"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			called: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			err: "unauthorized, user g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 doesn't have the required permission",
		},
		{
			name:       "is not a DAO member",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			setup: func() *Permissions {
				return New()
			},
			err: "unauthorized, user g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 doesn't have the required permission",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var called bool

			perms := tc.setup()
			testCaseFn := func() {
				perms.WithPermission(tc.user, tc.permission, tc.args, func() {
					called = true
				})
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, cur, tc.err, testCaseFn, "expect panic with message")
				return
			} else {
				urequire.NotPanics(t, cur, testCaseFn, "expect no panic")
			}

			urequire.Equal(t, tc.called, called, "expect callback to be called")
		})
	}
}

func TestBasicPermissionsSetPublicPermissions(t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	perms := New()

	// Add a new role with permissions
	perms.AddRole("adminRole", testPermA, testPermB, testPermC)
	urequire.False(t, perms.HasPermission(user, testPermA))
	urequire.False(t, perms.HasPermission(user, testPermB))
	urequire.False(t, perms.HasPermission(user, testPermC))

	// Assign a couple of public permissions
	perms.SetPublicPermissions(testPermA, testPermC)
	urequire.True(t, perms.HasPermission(user, testPermA))
	urequire.False(t, perms.HasPermission(user, testPermB))
	urequire.True(t, perms.HasPermission(user, testPermC))

	// Clear all public permissions
	perms.SetPublicPermissions()
	urequire.False(t, perms.HasPermission(user, testPermA))
	urequire.False(t, perms.HasPermission(user, testPermB))
	urequire.False(t, perms.HasPermission(user, testPermC))
}

func TestBasicPermissionsGetUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		user  address
		roles []string
		setup func() *Permissions
	}{
		{
			name:  "single role",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("admin", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
		},
		{
			name:  "multiple roles",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin", "bar", "foo"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("admin", testPermA)
				perms.AddRole("foo", testPermA)
				perms.AddRole("bar", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo", "bar")
				return perms
			},
		},
		{
			name: "without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *Permissions {
				perms := New()
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "not a user",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *Permissions {
				return New()
			},
		},
		{
			name:  "multiple users",
			user:  "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			roles: []string{"admin"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("admin", testPermA)
				perms.AddRole("bar", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				perms.SetUserRoles("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "admin")
				perms.SetUserRoles("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", "admin", "bar")
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
		setup func() *Permissions
		want  bool
	}{
		{
			name: "ok",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "admin",
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("admin", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin")
				return perms
			},
			want: true,
		},
		{
			name: "ok with multiple roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "foo",
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("admin", testPermA)
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "admin", "foo")
				return perms
			},
			want: true,
		},
		{
			name: "user without roles",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() *Permissions {
				perms := New()
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
		},
		{
			name: "has no role",
			user: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			role: "bar",
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()
			got := perms.HasRole(tc.user, tc.role)
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestBasicPermissionsHasPermission(t *testing.T) {
	cases := []struct {
		name       string
		user       address
		permission boards.Permission
		setup      func() *Permissions
		want       bool
	}{
		{
			name:       "ok",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple users",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermA,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
				perms.SetUserRoles("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "foo")
				return perms
			},
			want: true,
		},
		{
			name:       "ok with multiple roles",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermB,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.AddRole("baz", testPermB)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo", "baz")
				return perms
			},
			want: true,
		},
		{
			name:       "no permission",
			user:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			permission: testPermB,
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("foo", testPermA)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "foo")
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

func TestBasicPermissionsSetUserRoles(cur realm, t *testing.T) {
	cases := []struct {
		name          string
		user          address
		expectedRoles []boards.Role
		setup         func() *Permissions
		err           string
	}{
		{
			name:          "add user",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				return perms
			},
		},
		{
			name:          "add user with multiple roles",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a", "b"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				perms.AddRole("b", testPermB)
				return perms
			},
		},
		{
			name:          "add when other users exists",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				perms.SetUserRoles("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", "a")
				perms.SetUserRoles("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")
				return perms
			},
		},
		{
			name:          "add user using single role",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a"},
			setup: func() *Permissions {
				perms := New(UseSingleUserRole())
				perms.AddRole("a", testPermA)
				return perms
			},
		},
		{
			name:          "update user roles",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a", "b"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				perms.AddRole("b", testPermB)
				perms.AddRole("c", testPermB)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "c")
				return perms
			},
		},
		{
			name:          "update user roles using single role",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"b"},
			setup: func() *Permissions {
				perms := New(UseSingleUserRole())
				perms.AddRole("a", testPermA)
				perms.AddRole("b", testPermB)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a")
				return perms
			},
		},
		{
			name:          "clear user roles",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				perms.AddRole("b", testPermB)
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "a", "b")
				return perms
			},
		},
		{
			name:          "set invalid role",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a", "foo"},
			setup: func() *Permissions {
				perms := New()
				perms.AddRole("a", testPermA)
				return perms
			},
			err: "invalid role: foo",
		},
		{
			name:          "use single role error",
			user:          address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			expectedRoles: []boards.Role{"a", "b"},
			setup: func() *Permissions {
				perms := New(UseSingleUserRole())
				perms.AddRole("a", testPermA)
				perms.AddRole("b", testPermB)
				return perms
			},
			err: "user can only have one role",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perms := tc.setup()

			setUserRoles := func() {
				perms.SetUserRoles(tc.user, tc.expectedRoles...)
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, cur, tc.err, setUserRoles, "expected an error")
				return
			} else {
				urequire.NotPanics(t, cur, setUserRoles, "expected no error")
			}

			roles := perms.GetUserRoles(tc.user)
			uassert.Equal(t, len(tc.expectedRoles), len(roles))
			for i, r := range roles {
				urequire.Equal(t, string(tc.expectedRoles[i]), string(r))
			}
		})
	}
}

func TestBasicPermissionsAddRoleOverwrite(t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	perms := New()
	perms.AddRole("foo", testPermA)
	perms.SetUserRoles(user, "foo")

	// Re-adding an existing role replaces its permissions and keeps members
	perms.AddRole("foo", testPermB)
	uassert.True(t, perms.HasRole(user, "foo"), "expect role members to be kept")
	uassert.False(t, perms.HasPermission(user, testPermA), "expect old permissions to be revoked")
	uassert.True(t, perms.HasPermission(user, testPermB), "expect new permissions to be granted")
}

func TestBasicPermissionsAddRoleEmptyName(cur realm, t *testing.T) {
	urequire.PanicsWithMessage(t, cur, "role name is required", func() {
		New().AddRole("  ", testPermA)
	}, "expect whitespace only role names to be rejected")

	urequire.PanicsWithMessage(t, cur, "permissions super role name is required", func() {
		New(WithSuperRole("  "))
	}, "expect whitespace only super role names to be rejected")
}

func TestBasicPermissionsSuperRole(t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	super := boards.Role("owner")

	perms := New(WithSuperRole(super))
	urequire.True(t, perms.RoleExists(super), "expect super role to exist")
	urequire.False(t, perms.RoleExists("unknown"))

	// Super role members get every permission without explicit mapping
	perms.SetUserRoles(user, super)
	urequire.True(t, perms.HasRole(user, super))
	uassert.True(t, perms.HasUser(user))
	uassert.True(t, perms.HasPermission(user, testPermA))
	uassert.True(t, perms.HasPermission(user, testPermC))

	// AddRole on the super role is a no-op; it keeps all permissions
	perms.AddRole(super, testPermB)
	uassert.True(t, perms.HasPermission(user, testPermA), "expect super role to keep all permissions")
}

func TestBasicPermissionsHasUserUsersCount(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	guest := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")

	perms := New()
	perms.AddRole("foo", testPermA)
	urequire.Equal(t, 0, perms.UsersCount())
	urequire.False(t, perms.HasUser(member))

	// A user added only through a role counts as a user
	perms.SetUserRoles(member, "foo")
	urequire.True(t, perms.HasUser(member), "expect role holder to be a user")
	urequire.Equal(t, 1, perms.UsersCount())

	// A guest without roles counts as a user
	perms.SetUserRoles(guest)
	urequire.True(t, perms.HasUser(guest))
	urequire.Equal(t, 2, perms.UsersCount())

	// Clearing roles keeps the user and doesn't change the count
	perms.SetUserRoles(member)
	urequire.True(t, perms.HasUser(member), "expect user to remain after roles are cleared")
	urequire.Equal(t, 2, perms.UsersCount())

	// Re-assigning roles must not double count
	perms.SetUserRoles(member, "foo")
	urequire.Equal(t, 2, perms.UsersCount())
}

func TestBasicPermissionsSetUserRolesPanicKeepsState(cur realm, t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	perms := New()
	perms.AddRole("a", testPermA)
	perms.SetUserRoles(user, "a")

	urequire.PanicsWithMessage(t, cur, "invalid role: nope", func() {
		perms.SetUserRoles(user, "nope")
	}, "expect invalid role to panic")

	// The failed call must not have mutated state
	uassert.True(t, perms.HasUser(user))
	uassert.True(t, perms.HasRole(user, "a"))
	uassert.Equal(t, 1, perms.UsersCount())
}

func TestBasicPermissionsRemoveUser(t *testing.T) {
	cases := []struct {
		name  string
		user  address
		setup func() *Permissions
		want  bool
	}{
		{
			name: "ok",
			user: address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *Permissions {
				perms := New()
				perms.SetUserRoles("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return perms
			},
			want: true,
		},
		{
			name: "user not found",
			user: address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			setup: func() *Permissions {
				return New()
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

func TestBasicPermissionsRemoveUserWithRoles(t *testing.T) {
	user := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	perms := New()
	perms.AddRole("foo", testPermA)
	perms.AddRole("bar", testPermB)
	perms.SetUserRoles(user, "foo", "bar")

	urequire.True(t, perms.RemoveUser(user), "expect role holder to be removed")

	uassert.False(t, perms.HasUser(user), "expect removed user to not exist")
	uassert.False(t, perms.HasRole(user, "foo"), "expect removed user to keep no roles")
	uassert.False(t, perms.HasPermission(user, testPermA), "expect removed user to keep no permissions")
	uassert.Equal(t, 0, len(perms.GetUserRoles(user)))
	uassert.Equal(t, 0, perms.UsersCount())
	uassert.False(t, perms.RemoveUser(user), "expect second removal to report user not found")
}

func TestBasicPermissionsIterateUsers(t *testing.T) {
	// Users are listed sorted by address, including a roleless guest.
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
		{
			Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
		},
	}

	perms := New()
	perms.AddRole("foo", testPermA)
	perms.AddRole("bar", testPermB)
	// Add users in reverse order to pin that iteration is address-sorted,
	// not insertion-ordered.
	for i := len(users) - 1; i >= 0; i-- {
		perms.SetUserRoles(users[i].Address, users[i].Roles...)
	}

	cases := []struct {
		name               string
		start, count, want int
	}{
		{
			name:  "exceed users count",
			count: 50,
			want:  4,
		},
		{
			name:  "exact users count",
			count: 4,
			want:  4,
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
			count: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var visited int
			stopped := perms.IterateUsers(tc.start, tc.count, func(u boards.User) bool {
				i := tc.start + visited
				urequire.True(t, i < len(users), "expect iterator to respect number of users")
				uassert.Equal(t, users[i].Address, u.Address)

				urequire.Equal(t, len(users[i].Roles), len(u.Roles), "expect number of roles to match")
				for j := range u.Roles {
					uassert.Equal(t, string(users[i].Roles[j]), string(u.Roles[j]))
				}

				visited++
				return false
			})

			uassert.False(t, stopped, "expect full iteration not to report an early stop")
			uassert.Equal(t, tc.want, visited, "expect iterator to visit the windowed users")
		})
	}

	// Early stop is reported, and iteration actually halts.
	var visited int
	stopped := perms.IterateUsers(0, 10, func(boards.User) bool {
		visited++
		return true
	})
	uassert.True(t, stopped, "expect early stop to be reported")
	uassert.Equal(t, 1, visited, "expect iteration to halt on early stop")
}
