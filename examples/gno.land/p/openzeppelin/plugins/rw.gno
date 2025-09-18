package plugins

import "gno.land/p/openzeppelin/storage"

func PluginsKey() string {
	return "plugins/list"
}

func ReadPlugins(state storage.State) map[string]Func {
	v, ok := state.Get(PluginsKey())
	if !ok {
		return nil
	}
	return v.(map[string]Func)
}

func WritePlugins(state storage.State, list []*Plugin) {
	m := make(map[string]Func)
	for _, p := range list {
		m[p.ID] = p.Func
	}
	state.Set(PluginsKey(), m)
}
