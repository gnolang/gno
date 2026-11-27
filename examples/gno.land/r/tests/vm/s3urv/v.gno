package s3urv

// Probe victim for S3u review: fresh objects handed to a crossing hook.

type Entry struct{ Owner string }

var entries []*Entry

// Add stores a fresh entry, then crosses into the caller's hook with it.
func Add(cur realm, hook func(realm, *Entry)) {
	e := &Entry{Owner: "victim"}
	entries = append(entries, e)
	hook(cross(cur), e)
}

// AddNotify stores a fresh entry and only notifies the hook (no pointer given).
func AddNotify(cur realm, hook func(realm)) {
	entries = append(entries, &Entry{Owner: "victim"})
	hook(cross(cur))
}

// New returns a fresh entry not yet stored.
func New(cur realm) *Entry { return &Entry{Owner: "victim"} }

// Keep stores an entry previously minted by New.
func Keep(cur realm, e *Entry) { entries = append(entries, e) }

func Last() *Entry { return entries[len(entries)-1] }

func Owner(i int) string { return entries[i].Owner }

func Len() int { return len(entries) }

// --- round 2: the realm's own stored writer relayed into its gates.

// W is this realm's own stored writer, made at init.
var W = func(e *Entry) { e.Owner = "stamped" }

type Holder struct{ E *Entry }

var holders []*Holder

// AddRelay stores a fresh entry (in a fresh backing array), then runs f on it.
func AddRelay(cur realm, f func(*Entry)) {
	e := &Entry{Owner: "victim"}
	entries = append(entries, e)
	f(e)
}

// AddNested stores a fresh holder whose fresh entry is also passed to f.
func AddNested(cur realm, f func(*Entry)) {
	e := &Entry{Owner: "victim"}
	holders = append(holders, &Holder{E: e})
	f(e)
}

func HolderOwner(i int) string { return holders[i].E.Owner }

var byKey = map[string]*Entry{}

// AddMap stores a fresh entry in a stored map, then runs f on it.
func AddMap(cur realm, f func(*Entry)) {
	e := &Entry{Owner: "victim"}
	byKey["k"] = e
	f(e)
}

// AttachDetach attaches then detaches a fresh entry, runs f, then keeps it.
func AttachDetach(cur realm, f func(*Entry)) {
	e := &Entry{Owner: "victim"}
	byKey["t"] = e
	delete(byKey, "t")
	f(e)
	byKey["t"] = e
}

func KeyOwner(k string) string { return byKey[k].Owner }

// HolderW writes through a holder; stored at init.
var HolderW = func(h *Holder) { h.E.Owner = "stamped" }

// AddHolder stores a fresh holder and passes the holder itself.
func AddHolder(cur realm, f func(*Holder)) {
	h := &Holder{E: &Entry{Owner: "victim"}}
	holders = append(holders, h)
	f(h)
}

// Scratch runs f on a fresh, never-stored entry and reports its owner.
func Scratch(cur realm, f func(*Entry)) string {
	e := &Entry{Owner: "victim"}
	f(e)
	return e.Owner
}

// ScratchHolder runs f on a fresh, never-stored holder.
func ScratchHolder(cur realm, f func(*Holder)) string {
	h := &Holder{E: &Entry{Owner: "victim"}}
	f(h)
	return h.E.Owner
}

// Stamp2 is a declared writer, handed out as a value.
func Stamp2(e *Entry) { e.Owner = "stamped" }

// ScratchSelf runs this realm's own stored W from a later crossing call.
func ScratchSelf(cur realm) string { return Scratch(cur, W) }

// Self-relay gates: this realm's own code picks its stored W (R2a).
func SelfRelay(cur realm)  { AddRelay(cur, W) }
func SelfNested(cur realm) { AddNested(cur, W) }
func SelfMap(cur realm)    { AddMap(cur, W) }
func SelfDetach(cur realm) { AttachDetach(cur, W) }
func SelfHolder(cur realm) { AddHolder(cur, HolderW) }

// --- round 3: a fresh *string (unreal heap item holding a primitive).

// WS is this realm's own stored writer on a *string, made at init.
var WS = func(p *string) { *p = "stamped" }

var sptrs []*string

// AddPtr stores a fresh *string, then runs f on it.
func AddPtr(cur realm, f func(*string)) {
	s := "victim"
	p := &s
	sptrs = append(sptrs, p)
	f(p)
}

func SelfPtr(cur realm) { AddPtr(cur, WS) }

func PtrOwner(i int) string { return *sptrs[i] }

// --- round 4: non-crossing helpers another realm calls (read-only entry).

func FillStr(p *string)        { *p = "filled" }
func FillEntry(e *Entry)       { e.Owner = "filled" }
func FillArr(p *[2]string)     { p[0] = "filled" }
func IncInt(p *int)            { *p++ }
func AddInt(p *int)            { *p += 5 }
func AppendTo(p *[]string)     { *p = append(*p, "filled") }
func FillElem(s []string)      { s[0] = "filled" }
func FillMap(m map[string]int) { m["k"] = 7 }
func Call(f func())            { f() }
func RangeInto(p *string, xs []string) {
	for _, *p = range xs {
	}
}
