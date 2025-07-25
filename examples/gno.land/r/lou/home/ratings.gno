package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/lou/comment"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"

	"gno.land/r/sys/users"
)

var (
	ratings  *avl.Tree
	ratingID seqid.ID
)

func init() {
	ratings = avl.NewTree()
	ratingID = seqid.ID(0)
}

func NewRating(_ realm, rating uint64, commentStr string) string {
	caller := std.PreviousRealm().Address()
	if _, exists := ratings.Get(caller.String()); exists {
		panic(caller.String() + " has already rated and update instead")
	}
	id := ratingID.Next()

	p, err := comment.NewComment(rating, commentStr, std.PreviousRealm().Address(), id)
	if err != nil {
		panic(err)
	}

	ratings.Set(caller.String(), p)
	return ufmt.Sprintf("Successfully added a #%d rating!", rating)
}

func UpdateRating(cur realm, rating uint64, commentStr string) string {
	checkCaller()
	caller := std.PreviousRealm().Address()

	val, removed := ratings.Remove(caller.String())
	oldComment := val.(*comment.Comment)
	if !removed {
		panic("removal of rating was unsuccessful")
	}
	if commentStr == "" && oldComment.Comment() != nil {
		commentStr = *oldComment.Comment()
	}
	return NewRating(cur, rating, commentStr) + "\nSuccessfully updated rating!\n"
}

func UpdateByID(cur realm, comment string, id seqid.ID, rating uint64) string {
	caller := std.PreviousRealm().Address()

	if key, found := findRatingByID(id); found {
		if key == caller.String() {
			return UpdateRating(cur, rating, comment)
		}
	}
	panic("not authorized to update this rating")
}

func DeleteRating(_ realm, id seqid.ID) string {
	caller := std.PreviousRealm().Address()

	if Ownable.OwnedByPrevious() {
		if key, found := findRatingByID(id); found {
			ratings.Remove(key)
		}
		return ufmt.Sprintf("Rating as admin was removed successfully!")
	}
	checkCaller()
	if _, removed := ratings.Remove(caller.String()); !removed {
		panic("invalid removal of comment")
	}
	return ufmt.Sprintf("Rating was removed successfully!")
}

func renderPostRating() string {
	out := "# "
	out += md.Link("☆", txlink.Call("NewRating", "rating", "1", "commentStr", ""))
	out += md.Link("☆", txlink.Call("NewRating", "rating", "2", "commentStr", ""))
	out += md.Link("☆", txlink.Call("NewRating", "rating", "3", "commentStr", ""))
	out += md.Link("☆", txlink.Call("NewRating", "rating", "4", "commentStr", ""))
	out += md.Link("☆", txlink.Call("NewRating", "rating", "5", "commentStr", "")) + "\n\n"
	out += md.Link("[ ADD A RATING ]", txlink.Call("NewRating", "rating", "0", "comment", "Comment")) + "\n\n"
	return out
}

func renderRatings() string {
	out := md.HorizontalRule()
	out += md.H1("Ratings\n\n")
	out += renderPostRating()

	if ratings.Size() == 0 {
		out += "No ratings found."
		return out
	}
	ratings.ReverseIterate("", "", func(key string, value interface{}) bool {
		c := value.(*comment.Comment)
		out += "# " + comment.RenderRating(c.Rating())
		out += ufmt.Sprintf("\n> (%d) ", c.Rating())
		if *c.Comment() != "" {
			out += ufmt.Sprintf("\"%s\" ", *c.Comment())
		}

		out += "from "
		user := users.ResolveAddress(c.Author())
		if user != nil {
			out += md.Bold(user.RenderLink(""))
		} else {
			out += md.Bold(key)
		}

		out += " [" + md.Link("✎ EDIT", txlink.Call("UpdateByID", "comment", "Comment", "id", c.ID().String(), "rating", "0"))
		out += " | " + md.Link("🗑 DELETE", txlink.Call("DeleteRating", "id", c.ID().String()))
		out += " ]\n\n"
		return false
	})

	return out
}

func checkCaller() {
	caller := std.PreviousRealm().Address()
	if !caller.IsValid() {
		panic("invalid address")
	}

	if _, exists := ratings.Get(caller.String()); !exists {
		panic("rating either unrecognisable/ inaccessible from your address")
	}
}

func findRatingByID(id seqid.ID) (string, bool) {
	keyP := ""
	found := false

	ratings.Iterate("", "", func(key string, value any) bool {
		c := value.(*comment.Comment)
		if c.ID() == id {
			keyP = key
			found = true
			return true
		}
		return false
	})

	return keyP, found
}
