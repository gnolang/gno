package collections

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/lou/curator"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"

	"gno.land/r/sys/users"
)

var (
	gridMode           bool = false
	collectionID       seqid.ID = seqid.ID(0)
	collections        *avl.Tree = avl.NewTree()
	collectionsByTag   *avl.Tree = avl.NewTree()
	selectedCollection *avl.Tree = avl.NewTree()
	selectedTag        string
	tag                string
)

func NewCollection(title, description string, tags string) string {
	caller := std.PreviousRealm().Address()
	splitTags := strings.Fields(tags)
	id := collectionID.Next()

	c, err := curator.CreateCollection(id, title, description, splitTags, caller)
	if err != nil {
		panic(err)
	}

	collections.Set(id.String(), c)
	return "Successfully created collection! " + id.String()
}

func UpdateCollection(id, title, description string, tags []string) string {
	caller := std.PreviousRealm().Address()
	val, removed := collections.Remove(id)
	old := val.(*curator.Collection)
	if !removed {
		panic("update failed; couldn't remove old collection")
	}

	newCollection, err := curator.CreateCollection(old.ID(), title, description, tags, caller)
	if err != nil {
		panic(err)
	}

	collections.Set(caller.String(), newCollection)
	return "Successfully updated collection."
}

func DeleteCollection() string {
	caller := std.PreviousRealm().Address()

	if _, removed := collections.Remove(caller.String()); !removed {
		panic("couldn't delete collection.")
	}
	return "Collection deleted."
}

func AddToCollection(id, title, path string, tags string) string {
	splitTags := strings.Fields(tags)
	if !strings.Contains(path, "gno.land") {
		panic("please format with \"gno.land/\"<...>")
	}

	val, exists := collections.Get(id)
	if !exists {
		panic("collection is unrecognisable.")
	}
	collection := val.(*curator.Collection)

	c, err := curator.AddToCollection(collection, title, path, splitTags)
	if err != nil {
		panic(err)
	}
	if _, removed := collections.Remove(c.ID().String()); !removed {
		panic("adding of collection failed; couldn't remove current collection")
	}
	collections.Set(c.ID().String(), c)
	return title + " successfully added to collection."
}

func ToggleGridMode() {
	gridMode = !gridMode
}

func ToggleGridModeByID(id string) {
	if collection, exists := collections.Get(id); exists {
		c := collection.(*curator.Collection)
		c.ToggleGridMode()
	} else {
		panic("collection is unrecognisable.")
	}
}

func FilterCollections(tag string) {
	selectedTag = tag
	if tag == "" {
		collectionsByTag = nil
		return
	}

	collectionsByTag = collections
	collectionsByTag.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*curator.Collection)
		if !c.FilterTag(tag) {
			collectionsByTag.Remove(key)
		}
		return true
	})

}

func renderCollection(id string) string {
	if collection, exists := collections.Get(id); exists {
		c := collection.(*curator.Collection)
		out := md.H1(c.Title()) + "\n\n"
		out += md.HorizontalRule() + "\n\n"

		out += "### " + md.Link("ADD TO COLLECTION ", txlink.Call("AddToCollection", "id", c.ID().String())) + "\n\n"
		if c.GridMode() {
			out += md.Link("Toggle to list", txlink.Call("ToggleGridModeByID", "id", id)) + "\n\n"
			out += "<gno-columns>\n"
		} else {
			out += md.Link("Toggle to grid", txlink.Call("ToggleGridModeByID", "id", id)) + "\n\n"
		}

		out += curator.RenderCollection(c)
		return out
	}
	return "Collection is unrecognisable."
}

func Render(path string) string {
	if path != "" {
		return renderCollection(path)
	}
	out := md.H1("Collections") + "\n\n"
	out += md.HorizontalRule()

	out += "### " + md.Link("CREATE A COLLECTION ", txlink.Call("NewCollection", "title", "Title", "description", "Description", "tags", "Tags")) + "\n\n"
	if collectionsByTag != nil {
		out += "Filtered by tag: " + md.InlineCode(selectedTag) + " | "
		out += md.Link("Clear filter", txlink.Call("FilterCollections", "tag", "")) + "\n\n"
		selectedCollection = collectionsByTag
	} else {
		out += "### " + md.Link("Filter by tag ", txlink.Call("FilterCollections", "tag", "Tag")) + "\n\n"
		selectedCollection = collections
	}
	if selectedCollection.Size() == 0 {
		out += "No collections found.\n\n"
	} else {
		colCount := 0
		if gridMode {
			out += md.Link("Toggle to list", txlink.Call("ToggleGridMode")) + "\n\n"
			out += "<gno-columns>\n"
		} else {
			out += md.Link("Toggle to grid", txlink.Call("ToggleGridMode")) + "\n\n"
		}

		selectedCollection.ReverseIterate("", "", func(key string, value interface{}) bool {
			c := value.(*curator.Collection)
			user := users.ResolveAddress(c.Owner())

			out += c.PreviewCollection(gridMode)
			out += "\nfrom "
			if user != nil {
				out += md.Bold(user.RenderLink("")) + "\n\n"
			} else {
				out += md.Bold(c.Owner().String()) + "\n\n"
			}

			out += " [ " + md.Link("✎ EDIT", txlink.Call("UpdateCollection", "title", c.Title(), "desc", c.Description()))
			out += " | " + md.Link("🗑 DELETE", txlink.Call("DeleteCollection"))
			out += " ]\n\n"
			out += md.HorizontalRule()

			colCount++
			if colCount%3 == 0 {
				out += "\n</gno-columns>\n\n"
				if colCount != selectedCollection.Size() {
					out += "<gno-columns>\n"
				}
			} else if gridMode {
				out += "\n|||\n"
			}

			return false
		})
	}

	return out
}
