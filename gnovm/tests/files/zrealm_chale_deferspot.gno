// PKGPATH: gno.land/r/tests/vm/chaledefer
package chaledefer

var count int
var kept func()
var keptPanic func()
var self = func() { count++ }

func init() {
	kept = func() { defer func() { count++ }() }
	keptPanic = func() {
		defer func() { count++ }()
		panic("boom")
	}
}

func show(s string, err any) { println(s, count, err) }

func main(cur realm) {
	show("d01 on-the-spot defer inside a stored value", try(func() { kept() }))
	show("d02 stored value panics; its own defer writes", try(func() { keptPanic() }))
	// The realm's own frame recovers a read-only callee's panic, then writes.
	show("d03 writable frame writes after recovering", try(func() {
		func() {
			defer func() { recover(); count++ }()
			keptPanic()
		}()
	}))
	show("d05 package var self called by value", try(func() { self() }))
	self = func() { count += 100 }
	show("d06 reassigned package var", try(func() { self() }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// d01 on-the-spot defer inside a stored value 0 gno.land/r/tests/vm/chaledefer's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaledefer) package{})).count
// d02 stored value panics; its own defer writes 0 gno.land/r/tests/vm/chaledefer's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaledefer) package{})).count
// d03 writable frame writes after recovering 1 undefined
// d05 package var self called by value 1 gno.land/r/tests/vm/chaledefer's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaledefer) package{})).count
// d06 reassigned package var 1 gno.land/r/tests/vm/chaledefer's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaledefer) package{})).count
