package teams

type ActionFunc func(t *Team) Msg

type Action interface {
	call(t *Team) Msg
}

type action struct {
	actionFunc ActionFunc
}

func (a action) call(t *Team) Msg {
	return a.actionFunc(t)
}

func ActionableMsg(msg ...Msg) Action {
	switch len(msg) {
	case 0:
		return nil
	case 1:
		return Actionable(func(_ *Team) Msg {
			return msg[0]
		})
	default:
	}

	fns := make([]ActionFunc, len(msg))
	for i, m := range msg {
		fns[i] = func(_ *Team) Msg {
			return m
		}
	}
	return Actionable(fns...)
}

func Actionable(fn ...ActionFunc) Action {
	switch len(fn) {
	case 0:
		return nil
	case 1:
		return &action{actionFunc: fn[0]}
	default:
	}

	actions := make([]Action, len(fn))
	for i, f := range fn {
		actions[i] = &action{actionFunc: f}
	}
	return ChainActions(actions...)
}

func ChainActions(actions ...Action) Action {
	switch len(actions) {
	case 0:
		return nil
	case 1:
		return actions[0]
	default:
	}

	return Actionable(func(t *Team) Msg {
		msgs := make([]Msg, len(actions))
		for i, action := range actions {
			msgs[i] = action.call(t)
		}
		return msgs
	})
}
