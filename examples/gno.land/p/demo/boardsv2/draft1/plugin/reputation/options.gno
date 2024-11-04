package reputationplugin

type Option func(*Plugin)

func UseTokenBasePolicy() Option {
	return func(p *Plugin) {
		p.Policy = PolicyTokenBase
	}
}

func AllowedPostLevels(levels []int) Option {
	return func(p *Plugin) {
		p.AllowedPostLevels = levels
	}
}
