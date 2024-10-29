package pluginfork

type Option func(*Plugin)

func AllowedPostLevels(levels []int) Option {
	return func(p *Plugin) {
		p.AllowedPostLevels = levels
	}
}
