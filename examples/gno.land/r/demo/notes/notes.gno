package notes

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/btree"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

var (
	users avl.Tree // std.Address -> BTree
)

// TODO : add private notes with a check for the caller

type Note struct {
	id        int
	content   string
	created   time.Time
	charCount int
	private   bool
}

func (n Note) Less(than btree.Record) bool {
	other, ok := than.(Note)
	if !ok {
		panic("invalid comparison with non-Note record")
	}
	return n.created.Before(other.created)
}

func AddNote(noteContent string, private bool) {
	callerString := std.GetOrigCaller().String()

	userProfile, exists := users.Get(callerString)
	if !exists {
		newNote := Note{
			id:        0,
			content:   noteContent,
			created:   time.Now(),
			charCount: len(noteContent),
			private:   private,
		}

		userNotes := btree.New()
		userNotes.Insert(newNote)
		users.Set(callerString, userNotes)
	} else {
		userNotes := userProfile.(*btree.BTree)
		newNote := Note{
			id:        userNotes.Len(),
			content:   noteContent,
			created:   time.Now(),
			charCount: len(noteContent),
			private:   private,
		}
		userNotes.Insert(newNote)
	}
}

func GetNotes(caller std.Address) []Note {
	userNotes, exists := users.Get(caller.String())
	if !exists {
		panic("user notes not found")
	}
	notes := make([]Note, 0, userNotes.(*btree.BTree).Len())
	userNotes.(*btree.BTree).Ascend(func(r btree.Record) bool {
		notes = append(notes, r.(Note))
		return true
	})
	return notes
}

func GetNoteCount(caller std.Address) int {
	userNotes, exists := users.Get(caller.String())
	if !exists {
		panic("user notes not found")
	}
	return userNotes.(*btree.BTree).Len()
}

func GetNote(caller std.Address, id int) Note {
	userNotes, exists := users.Get(caller.String())
	if !exists {
		panic("user notes not found")
	}

	var foundNote Note
	userNotes.(*btree.BTree).Ascend(func(r btree.Record) bool {
		note := r.(Note)
		if note.id == id {
			foundNote = note
			return false
		}
		return true
	})

	return foundNote
}

func DeleteNote(caller std.Address, id int) {
	userNotes, exists := users.Get(caller.String())
	if !exists {
		panic("user notes not found")
	}

	btreeNotes := userNotes.(*btree.BTree)
	var noteToDelete Note
	btreeNotes.Ascend(func(r btree.Record) bool {
		note := r.(Note)
		if note.id == id {
			noteToDelete = note
			return false
		}
		return true
	})

	if noteToDelete.id == id {
		btreeNotes.Delete(noteToDelete)
	}
}

func Render(path string) string {
	userNotes := GetNotes(std.Address(path))
	if len(userNotes) == 0 {
		return ufmt.Sprintf("No notes found.")
	}

	var sb strings.Builder
	sb.WriteString(md.H1("User Notes"))

	caller := std.GetOrigCaller()
	isOwner := caller.String() == path

	for _, note := range userNotes {
		if note.private && !isOwner {
			continue
		}

		header := ufmt.Sprintf("Note #%d", note.id)
		if note.private {
			header += " (Private)"
		}
		sb.WriteString(md.H3(header))
		sb.WriteString(note.content)
		sb.WriteString("\n")
		sb.WriteString(md.Blockquote(ufmt.Sprintf("%d chars,", note.charCount)))
		sb.WriteString("\t")
		sb.WriteString(md.Italic(note.created.Format("01.02.2006 15:04")))
		sb.WriteString("\n")
	}

	return sb.String()
}
