package plugins

import (
	"gno.land/p/openzeppelin/storage"
)

type Func func(state storage.State, args ...any)

type Plugin struct {
	ID   string
	Func Func
}

func Run(id string, args ...any) storage.ExecFn {
	return func(state storage.State) any {
		ps := ReadPlugins(state)
		if len(ps) == 0 {
			return nil
		}
		fn, ok := ps[id]
		if ok {
			fn(state, args...)
		}
		return nil
	}
}

func SetPlugins(ps []*Plugin) storage.ExecFn {
	return func(state storage.State) any {
		WritePlugins(state, ps)
		return nil
	}
}
