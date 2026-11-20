package s3uprobe

type Entry struct{ Owner string }

var entries []*Entry

// Add stores a fresh entry, then runs the caller's hook in the same call.
func Add(cur realm, hook func()) {
	e := &Entry{Owner: "victim"}
	entries = append(entries, e)
	hook()
}

// Last hands out the newest entry (a read).
func Last() *Entry { return entries[len(entries)-1] }

func Owner0() string { return entries[0].Owner }

// AddPass stores a fresh entry, then hands it to the caller's hook.
func AddPass(cur realm, hook func(*Entry)) {
	e := &Entry{Owner: "victim"}
	entries = append(entries, e)
	hook(e)
}

// AddSlot stores a fresh entry; the hook may only see the slice.
func AddSlot(cur realm, hook func([]*Entry)) {
	entries = append(entries, &Entry{Owner: "victim"})
	hook(entries)
}

type Inner struct{ Owner string }

type Box struct{ In *Inner }

var boxes []*Box

// Stamp is this realm's own stored function, made at init.
var Stamp = func(b *Box) { b.In.Owner = "stamped" }

// AddBox stores a fresh box, then runs the caller-chosen function on it.
func AddBox(cur realm, f func(*Box)) {
	b := &Box{In: &Inner{Owner: "victim"}}
	boxes = append(boxes, b)
	f(b)
}

func BoxOwner0() string { return boxes[0].In.Owner }
