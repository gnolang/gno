package boardsv2

import "errors"

func AddBoard(s PostStore, slug string /* ... */) (path string, _ error) {
	// TODO: Finish implementation

	return slug, nil
}

// NOTE: Define a pattern to add functionality to posts by type (AddComment, AddThread, AddPoll, Repost, Upvote, ...)
// NOTE: Maybe though functions that assert the right arguments
func AddComment(s PostStore, parentPath string, creator std.Address, message string) (path string, _ error) {
	// Try to get parent as a post or a comment, otherwise parent doesn't support comments
	p, found := s.Get(LevelPost, parentPath)
	if !found {
		p, found = s.Get(LevelComment, parentPath)
		if !found {
			return "", errors.New("parent post or comment not found: " + parentPath)
		}
	}

	// TODO:
	// Call the IsLocked function from the plugin for both the board post and thread post
	// of this new comment. And confirm that both of them are false
	// if so, then proceed, otherwise can not add new comments because locked.
	// level 0 - boards
	// level 1 - thread
	// level 2 - comment
	// level 3 - comment under comment
	// level 4 - comment under comment under comment
	// ...

	// TODO:
	// Consider using reverse iteration while checking IsLocked in parent levels.
	// If the keys in the AVL tree has levels as the prefix it should be optimized. If
	// timestamp is used it may not be.

	comment := NewComment(p /* ... */)

	// TODO: Finish implementation
	s.Set( /* ... */ )

	path = parentPath + "/" + comment.ID
	return path, nil
}

// NOTE: Arguments could potentially be many, consider variadic + sane defaults (?)
func AddThread(s PostStore, parentPath, slug string, creator std.Address /* ... */) (path string, _ error) {
	p, found := b.Get(LevelPost, parentPath)
	if !found {
		return "", errors.New("parent post not found: " + parentPath)
	}

	post := NewPost(p, slug /* ... */)

	// TODO: Finish implementation

	path = parentPath + "/" + post.ID
	return path, nil
}

// ----- Other features -----
// type VotesStore interface {
// 	/*...*/
// }
//
// func Upvote(s VotesStore /* ... */)   {}
// func DownVote(s VotesStore /* ... */) {}
