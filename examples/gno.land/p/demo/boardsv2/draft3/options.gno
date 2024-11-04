package boards

import (
	pluginreputation "gno.land/p/demo/boards/post/plugin/reputation"
)

// Option configures board applications.
type Options func(*App)

// LinearReputationPolicy allows upvoting or downvoting a post by one for each account.
func LinearReputationPolicy() Option {
	return func(a *App) {
		a.reputationPolicy = pluginreputation.PolicyLinear
	}
}

// TokenBasedReputationPolicy allows upvoting or downvoting
// a post propotional to the specified tokens that an account holds.
func TokenBasedReputationPolicy() Option {
	return func(a *App) {
		a.reputationPolicy = pluginreputation.PolicyTokenBased
	}
}

// MaxCommentsDepth configures the max depth for nested comments.
// Setting it to -1 allows an infinite number of nested comments (default).
func MaxCommentsDepth(d int) Option {
	return func(a *App) {
		a.maxCommentsDepth = d
	}
}
