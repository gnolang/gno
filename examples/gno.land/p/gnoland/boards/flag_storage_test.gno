package boards_test

import (
	"testing"

	"gno.land/p/nt/urequire"

	"gno.land/p/gnoland/boards"
)

func TestFlagStorageExists(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() boards.FlagStorage
		user   address
		exists bool
	}{
		{
			name: "found",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"})
				return s
			},
			user:   "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			exists: true,
		},
		{
			name: "not found",
			setup: func() boards.FlagStorage {
				return boards.NewFlagStorage()
			},
			user:   "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			urequire.Equal(t, tt.exists, s.Exists(tt.user))
		})
	}
}

func TestFlagStorageAdd(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() boards.FlagStorage
		flag   boards.Flag
		errMsg string
	}{
		{
			name: "ok",
			setup: func() boards.FlagStorage {
				return boards.NewFlagStorage()
			},
			flag: boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name: "flag exists",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"})
				return s
			},
			flag:   boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
			errMsg: "flag from user already exists: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name: "invalid user address",
			setup: func() boards.FlagStorage {
				return boards.NewFlagStorage()
			},
			flag:   boards.Flag{User: "foo"},
			errMsg: "post flagging error, invalid user address: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			err := s.Add(tt.flag)

			if tt.errMsg != "" {
				urequire.Error(t, err, "expect an error")
				urequire.ErrorContains(t, err, tt.errMsg, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.True(t, s.Exists(tt.flag.User), "expect flag to be added")
		})
	}
}

func TestFlagStorageRemove(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() boards.FlagStorage
		address address
		removed bool
	}{
		{
			name: "ok",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"})
				return s
			},
			address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			removed: true,
		},
		{
			name: "not found",
			setup: func() boards.FlagStorage {
				return boards.NewFlagStorage()
			},
			address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			urequire.Equal(t, tt.removed, s.Remove(tt.address))
		})
	}
}

func TestFlagStorageSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() boards.FlagStorage
		size  int
	}{
		{
			name: "empty",
			setup: func() boards.FlagStorage {
				return boards.NewFlagStorage()
			},
			size: 0,
		},
		{
			name: "one flag",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"})
				return s
			},
			size: 1,
		},
		{
			name: "multiple flags",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(boards.Flag{User: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"})
				s.Add(boards.Flag{User: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"})
				return s
			},
			size: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			urequire.Equal(t, tt.size, s.Size())
		})
	}
}

func TestFlagStorageIterate(t *testing.T) {
	flags := []boards.Flag{
		{
			User:   "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			Reason: "a",
		},
		{
			User:   "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			Reason: "b",
		},
	}
	tests := []struct {
		name    string
		setup   func() boards.FlagStorage
		reverse bool
		flags   []boards.Flag
	}{
		{
			name: "default",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(flags[0])
				s.Add(flags[1])
				return s
			},
			flags: flags,
		},
		{
			name: "reverse",
			setup: func() boards.FlagStorage {
				s := boards.NewFlagStorage()
				s.Add(flags[0])
				s.Add(flags[1])
				return s
			},
			reverse: true,
			flags:   []boards.Flag{flags[1], flags[0]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			count := s.Size()
			if tt.reverse {
				count = -count
			}

			var i int
			s.Iterate(0, count, func(f boards.Flag) bool {
				urequire.Equal(t, tt.flags[i].User, f.User, "expect user to match")
				urequire.Equal(t, tt.flags[i].Reason, f.Reason, "expect reason to match")

				i++
				return false
			})
		})
	}
}
