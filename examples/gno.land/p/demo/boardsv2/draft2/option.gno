package boards

type Option struct{}

// LinearReputationPolicy allows upvoting or downvoting a post by one
// for each account.
func LinearReputationPolicy() Option {}

// TokenBasedReputationPolicy allows upvoting or downvoting a post propotional
// to the specified tokens that an account holds.
func TokenBasedReputationPolicy() Option {}

// MaxPostDepth configures the max depth for nested comments.
// 0 -> boards
// 1 -> threads
// 2 -> comments-1 (direct comments to the threads)
// The above are already reserved.
// Setting it to zero will disable comments.
func MaxCommentDepth(d int) Option {}
