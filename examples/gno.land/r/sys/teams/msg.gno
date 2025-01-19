package teams

import "std"

type Msg interface{}

type AddMemberMsg struct {
	Member std.Address
}

func AddMemberAction(member std.Address) Action {
	return Actionable(func(t *Team) Msg {
		if t.members.Has(member.String()) {
			panic(ErrAleadyExist)
		}
		t.members.Set(member.String(), struct{}{})
		return nil
	})
}

type RemoveMemberMsg struct {
	Member std.Address
}

func RemoveMemberAction(member std.Address) Action {
	return Actionable(func(t *Team) Msg {
		if t.members.Has(member.String()) {
			panic(ErrAleadyExist)
		}
		t.members.Set(member.String(), struct{}{})
		return nil
	})
}

type AddPackageMsg struct {
	Path string
}
