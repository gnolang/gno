// PKGPATH: gno.land/r/twitter_tests
package twitter_tests

import (
	"gno.land/r/twitter"
)

var gProfileID twitter.ProfileID

func init() {
	gProfileID = twitter.CreateProfile("Milerius")
	//boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	//pid := boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
	//boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println("-----------")
	println(twitter.Render("Milerius"))
	println("-----------")
}
