package boardsv2

type Option struct{}

// LinearReputationPolicy allows upvoting or downvoting a post by one
// for each account.
func LinearReputationPolicy() Option {}

// TokenBasedReputationPolicy allows upvoting or downvoting a post propotional
// to the specified tokens that an account holds.
func TokenBasedReputationPolicy() Option {}

// TODO: make it configurable how many levels allowed
