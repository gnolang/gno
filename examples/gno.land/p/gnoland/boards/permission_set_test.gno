package boards

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestNewPermissionSet(t *testing.T) {
	cases := []struct {
		name  string
		perms []Permission
		check []Permission
		want  bool
	}{
		{
			name:  "empty",
			check: []Permission{0},
			want:  false,
		},
		{
			name:  "single permission",
			perms: []Permission{0},
			check: []Permission{0},
			want:  true,
		},
		{
			name:  "multiple permissions",
			perms: []Permission{1, 3, 5},
			check: []Permission{1, 3, 5},
			want:  true,
		},
		{
			name:  "high permission value",
			perms: []Permission{100},
			check: []Permission{100},
			want:  true,
		},
		{
			name:  "missing permission",
			perms: []Permission{100},
			check: []Permission{0},
			want:  false,
		},
		{
			name:  "multiple missing permissions",
			perms: []Permission{1, 3, 5},
			check: []Permission{0, 2, 4},
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewPermissionSet(tc.perms...)

			for _, p := range tc.check {
				uassert.Equal(t, tc.want, s.Has(p))
			}
		})
	}
}

func TestPermissionSetHas(t *testing.T) {
	cases := []struct {
		name  string
		set   PermissionSet
		check Permission
		want  bool
	}{
		{
			name:  "out of range",
			set:   NewPermissionSet(0),
			check: 100,
			want:  false,
		},
		{
			name:  "nil set",
			check: 0,
			want:  false,
		},
		{
			name:  "permission present",
			set:   NewPermissionSet(5),
			check: 5,
			want:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uassert.Equal(t, tc.want, tc.set.Has(tc.check))
		})
	}
}

func TestPermissionSetIsEmpty(t *testing.T) {
	cases := []struct {
		name string
		set  PermissionSet
		want bool
	}{
		{
			name: "nil set",
			want: true,
		},
		{
			name: "non-empty set",
			set:  NewPermissionSet(0),
			want: false,
		},
		{
			name: "empty allocated set",
			set:  make(PermissionSet, 1),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uassert.Equal(t, tc.want, tc.set.IsEmpty())
		})
	}
}
