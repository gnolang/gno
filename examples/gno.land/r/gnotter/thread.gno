package gnotter

import "std"

type ThreadID uint64

// Thread A thread is a Post or a Retweet/Comment on a Profile.
type Thread struct {
	profile *Profile
	id      ThreadID
	creator std.Address
	body    string
	likes   uint64
	// TODO: comments/hashtag
	threadID  ThreadID // original Thread.id
	parentID  ThreadID // parent thread.id (if comment or retweet)
	createdAt std.Time
}
