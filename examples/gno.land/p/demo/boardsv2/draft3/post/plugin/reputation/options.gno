package pluginreputation

type Option func(*Plugin)

func UsePolicy(v Policy) Option {
	return func(p *Plugin) {
		p.Policy = v
	}
}

func AllowedPostLevels(levels []int) Option {
	return func(p *Plugin) {
		p.AllowedPostLevels = levels
	}
}
