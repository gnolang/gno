package boardsv2

import "gno.land/p/demo/avl"

// TODO: This goes in the realm
// type Boards struct {
// 	// TODO: Define how do we want to display and sort boards and posts (upvotes, pinned, ...)
// 	boards avl.Tree
// 	Title       string
// 	Description string
// }

func Render(path string) string {
	// TODO: Implement render
	return ""
}

// TODO: Define public API

func CreateBoard() {} // Maybe
func EditBoard()   {} // Maybe
func ForkBoard()   {} // Maybe

func CreatePost() {}
func EditPost()   {}
func ForkPost()   {}
func DeletePost() {}
func Repost()     {}
func Pin()        {}
func Invite()     {} // Maybe: Could also rely on an allow list
func UpVote()     {}
func DownVote()   {}

func Comment()       {} // Maybe
func EditComment()   {} // Maybe
func DeleteComment() {} // Maybe

func ToggleCommentsSupport() {} // Maybe
func ToggleThreadsSupport()  {} // Maybe
func GetTags()               {} // Maybe: List of allowed tags (moderated)

func AddModerator()    {} // Maybe
func RemoveModerator() {} // Maybe
func GetModerators()   {} // Maybe
