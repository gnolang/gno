package teams

import (
	"errors"
	"std"
)

var ErrAlreadyExist = errors.New("already exist")

type Cmd interface{}

type AddMemberCmd struct {
	Member std.Address
}

func AddMemberTask(member std.Address) Task {
	return CreateTask(func(t *Team) Cmd {
		if t.members.Has(member.String()) {
			panic(ErrAlreadyExist)
		}
		t.members.Set(member.String(), struct{}{})
		return nil
	})
}

type RemoveMemberCmd struct {
	Member std.Address
}

func RemoveMemberTask(member std.Address) Task {
	return CreateTask(func(t *Team) Cmd {
		if t.members.Has(member.String()) {
			panic(ErrAlreadyExist)
		}
		t.members.Set(member.String(), struct{}{})
		return nil
	})
}

type AddPackageCmd struct {
	Path string
}
