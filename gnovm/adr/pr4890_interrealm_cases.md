# Exercises in Interrealm Cases

These are an attempt to go over some cases systematically.  It may help to go
over these cases, but they are rather repetitive.  In the future as more cases
are considered, they should be added here systematically, and also duplicated
in gno.land/pkg/integration/testdata/interrealm_v2.txtar.  That file should
ideally be transcribed as a test file that makes use of revive() and "testing".

This document has been updated for **interrealm spec v2** (see
`gnovm/adr/interrealm_v2.md`).  The v1 case outcomes ("do not storage-cross"
etc.) are preserved in the three sibling test files
(`interrealm_final.txtar`, `interrealm_mix_call.txtar`, `interrealm_mix_run.txtar`)
for historical comparison; the affirmative-coverage v2 test suite lives in
`interrealm_v2.txtar`.

**Closure-capability reclassification.** PR #4890 originally used `ExecGood`
in `interrealm_final_omarsy.txtar` as the positive control: `peter.F` (called
from `/r/main`) builds a closure that captures an `/r/alice` slice; alice
then stores and executes it.

Under Rule 3 of `PushFrameCall`, a closure carries the authority of its
minter. The closure was minted while `m.Realm = /r/main`, so it cannot
write `/r/alice` — alice's execution does not raise the closure's authority.
Both `ExecGood` (renamed `ExecPreviouslyGood`) and `ExecBad` therefore fail
with the readonly-taint error. To get a closure that writes `/r/A`'s state,
`/r/A` itself must mint it.

The two key v2 semantics changes relative to v1:

1. **Layer 1 (declaring-realm borrow) fires uniformly** for every /r/-declared
   callable — function, method on any receiver shape, closure whose
   construction site was /r/X, or interface dispatch to an /r/X-declared
   method.  m.Realm shifts to /r/X for the call duration.  v1 fired this only
   for some method shapes; v2 makes it uniform, so the "do not storage-cross"
   cases of v1 now correctly route writes back to the declaring realm at each
   hop.

2. **Construction-time enforcement.**  Foreign /r/-declared types cannot be
   constructed via composite literal, `new(...)`, or `make(...)` from outside
   the declaring realm.  Several v1 cases that used `new(alice.Object)` from a
   caller realm now fail at construction time rather than at the later
   storage-cross check.  Use a constructor function (e.g. `alice.NewObject()`)
   to mint foreign-realm objects through the declaring realm.

"Storage-crossing" is like (regular) crossing for method calls except it
does not affect CurrentRealm()/PreviousRealm(). Since method calls can
be nested the storage-realms are represented as a stack which maybe
empty.

"Present realm" is the present storage-realm if any storage-crossing
has occurred by method call since the last explicit cross. Otherwise
it is the realm last explicitly crossed into.

When calling a method of a nil receiver declared in an /r/realm we
storage-cross to the storage realm /r/realm.

TODO "implicit-cross" is a misnomer, replace all w/ "storage-cross".

----------------------------------------
## fn is declared /r/realm:

 * fn is (non-crossing) function declared in /r/realm:
   - CASE rA1: function declared at package level                   -- Layer 1 borrows (succeeds)
   - CASE rA2: function declared in func(...) (non-crossing)        -- Layer 1 borrows (succeeds)
   - CASE rA3: function declared in func(cur realm, ...) (crossing) -- Layer 1 borrows (succeeds)

 * fn is method declared in /r/realm (nil or unreal receiver):
   - CASE rB1: method has nil receiver                              -- Layer 1 borrows (succeeds)
   - CASE rB2: method has nil receiver (external)                   -- Layer 1 borrows (succeeds)
   - CASE rB3: method has unreal receiver (inline)                  -- construction-time panic
   - CASE rB4: method has unreal receiver (var)                     -- construction-time panic

 * fn is method declared in /r/realm (real receiver):
   - CASE rC1: method has real receiver                             -- construction-time panic
   - CASE rC2: method has real receiver (external)                  -- construction-time panic
   - CASE rC3: method has real receiver via closure                 -- construction-time panic
   - CASE rC4: method has real receiver via closure (external)      -- construction-time panic

 * fn is crossing function declared in /r/realm:
   - CASE rD1: direct modification from external closure            -- illegal modification (static)

----------------------------------------
## fn is declared in /p/package:

These cases are similar to /r/realm except for CASE rA3 and rD1 since
/p/package cannot contain crossing functions.

Note that the test cases in interrealm_v2.txtar are similar but modified to
account for the difference between realm and pure package.

 * fn is function declared in /p/package:
   - CASE pA1: function declared at package level                   -- Layer 1 borrows at write site
   - CASE pA2: function declared in func(...) (non-crossing)        -- Layer 1 borrows at write site
   - CASE pA3: function declared in func(cur realm, ...) (crossing) -- illegal crossing function (static)

 * fn is method declared in /p/package (nil or unreal receiver):
   - CASE pB1: method has nil receiver                              -- Layer 1 borrows at write site
   - CASE pB2: method has nil receiver (external)                   -- Layer 1 borrows at write site
   - CASE pB3: method has unreal receiver (inline)                  -- Layer 1 borrows at write site
   - CASE pB4: method has unreal receiver (var)                     -- Layer 1 borrows at write site

 * fn is method declared in /p/package:
   - CASE pC1: method has real receiver                             -- Layer 2 + Layer 1 borrow (succeeds)
   - CASE pC2: method has real receiver (external)                  -- Layer 2 + Layer 1 borrow (succeeds)
   - CASE pC3: method has real receiver via closure                 -- Layer 2 + Layer 1 borrow (succeeds)
   - CASE pC4: method has real receiver via closure (external)      -- Layer 2 + Layer 1 borrow (succeeds)

----------------------------------------
## NOTEs

 * Assigning to a field of a yet unreal object does NOT result in immediate
   recursive attachement to realm; it may happen during finalization; in the
   future "attach()" will allow for realm attachment before finalization.
 * After finalization all objects are persisted in one realm or another.
 * A variable declaration via `:=` or `var` in a block is initially unreal and
   unattached.
 * Function values are never parents except for the function's captured names.
 * TODO verify all of the above.

 * Under v2, "Layer 1 borrows (succeeds)" means: m.Realm shifts to the
   declaring realm for the call duration; any write to the declaring realm's
   own state from inside the body lands cleanly.  For the /r/ cases, the
   write site is the called fn's body directly; for the /p/ cases, the called
   fn invokes `bob.PrivateFunc` (which is /r/bob-declared), and Layer 1 fires
   at THAT inner call — hence "at write site."
 * "Layer 2 + Layer 1 borrow (succeeds)" describes /p/-methods invoked on a
   defined foreign receiver: Layer 2 first shifts m.Realm to the receiver's
   authoring realm, then the inner `bob.PrivateFunc` call fires Layer 1 to
   /r/bob where the write lands.
 * "construction-time panic" replaces the v1 "FAIL: caller != bob" outcome for
   cases that use `new(alice.Object)`.  The construction is rejected before
   any of the storage-cross logic runs.  Under v2, `alice.NewObject()` (a
   constructor function declared in /r/alice) is the only way for foreign
   realms to obtain an `alice.Object`.
	
The pseudocode walkthrough below is the **v1** trace.  Under v2 the
storage-realm annotations would be different — every /r/-declared callable
pushes its own realm onto the storage stack (Layer 1), so e.g. in `bob.Do`
the storage would be `[bob]` rather than `[]`, and inside `alice.TopFunc` it
would be `[bob,alice]`.  The "FAIL: caller != bob" labels on rA1/rA2/rA3/rB1/rB2
become SUCCESS because Layer 1 routes the inner `bob.PrivateFunc` write back
to `/r/bob`; the rB3/rB4/rC1-rC4 labels become "PANIC: cannot allocate
gno.land/r/alice.Object in realm gno.land/r/caller" at the `new(alice.Object)`
line (construction-time enforcement).  The walkthrough is preserved as-is
because it documents the *case structure* and v1 expectations.  For v2
outcomes see the case-summary table above and `interrealm_v2.txtar`.

```go
	//========================================
	// main.go
	package caller // as /r/caller/run (msg call)
	import `alice`
	import `peter`
	import `bob`
	func main() {

		// The storage realm(s) is initialy an empty stack (no methods have been called).
		//   - now: (storage:[], current:caller, previous:nil)

		// ================================================================================
		// ## fn is declared /r/realm:

		// --------------------------------------------------------------------------------
		// CASE rA1: function declared at package level -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in alice.TopFunc:                    (storage:[], current:caller, previous:nil)
		bob.Do(alice.TopFunc)                       // FAIL: caller != bob
		// ----------------------------------------
		// CASE rA2: function declared in func(...) (non-crossing) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in alice.TopFuncFnNoncrossing:       (storage:[], current:caller, previous:nil)
		//   - in alice.TopFuncFnNoncrossing.inner: (storage:[], current:caller, previous:nil)
		bob.Do(alice.TopFuncFnNoncrossing())        // FAIL: caller != bob
		// ----------------------------------------
		// CASE rA3: function declared in func(cur realm, ...) (crossing) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - alice.TopFuncFnCrossing:             (storage:[], current:alice,  previous:caller)
		//   - alice.TopFuncFnCrossing.inner:       (storage:[], current:caller, previous:nil)
		//     (cross doesn't matter in inner)
		bob.Do(alice.TopFuncFnCrossing(cross))      // FAIL: caller != bob

		// --------------------------------------------------------------------------------
		// CASE rB1: method has nil receiver -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do((*bob.Object)(nil).Method)           // FAIL: caller != bob
		// ----------------------------------------
		// CASE rB2: method has nil receiver (external) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do((*alice.Object)(nil).Method)         // FAIL: caller != bob
		// ----------------------------------------
		// CASE rB3: method has unreal receiver (inline) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do(new(alice.Object).Method)            // FAIL: caller != bob
		// ----------------------------------------
		// CASE rB4: method has unreal receiver (var) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(alice.Object)
		bob.Do(obj.Method)                          // FAIL: caller != bob

		// --------------------------------------------------------------------------------
		// CASE rC1: method has real receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[bob], current:caller, previous:nil)
		obj := new(alice.Object)
		bob.SetObject(obj)                          // obj persisted in bob
		bob.Do(obj.Method)                          // SUCCESS: resides in bob
		// ----------------------------------------
		// CASE rC2: method has real receiver (external) -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[alice], current:caller, previous:nil)
		obj := new(alice.Object)
		alice.SetObject(obj)                        // obj persisted in alice
		bob.Do(obj.Method)                          // FAIL: resides in alice (needs bob)
		// ----------------------------------------
		// CASE rC3: method has real receiver via closure -- storage-cross to receiver
		obj := new(alice.Object)
		cls := func() {
			print(obj)
		}
		bob.SetClosure(cls)                         // obj persisted in bob via cls
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do(obj.Method)                          // SUCCESS: resides in bob
		// ----------------------------------------
		// CASE rC4: method has real receiver via closure (external) -- stroage-cross to receiver
		obj := new(alice.Object)
		cls := func() {
			print(obj)
		}
		alice.SetClosure(cls)                       // obj persisted in alice via cls
		//   - in bob.Do:                           (storage:[],      current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[alice], current:caller, previous:nil)
		bob.Do(obj.Method)                          // FAIL: alice != bob

		// --------------------------------------------------------------------------------
		// CASE rD1: direct modification from external closure -- illegal modification
		//   - in bob.ExecuteClosure:               (storage:[], current:bob, previous:caller)
		f := func() {
			bob.AllowedList = append(bob.AllowedList, 3) // INVALID (illegal modification)
		    panic("bob.AllowedList should not be mutable via dot selector from an external realm; static error expected")
		}


		// ================================================================================
		// ## fn is declared /p/package: (similar to /r/realm except for CASE pA3 and CASE pB2)

		// --------------------------------------------------------------------------------
		// CASE pA1: function declared at package level -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in peter.TopFunc:                    (storage:[], current:caller, previous:nil)
		bob.Do(peter.TopFunc)                       // FAIL: caller != bob
		// ----------------------------------------
		// CASE pA2: function declared in func(...) (non-crossing) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in peter.TopFuncFnNoncrossing:       (storage:[], current:caller, previous:nil)
		//   - in peter.TopFuncFnNoncrossing.inner: (storage:[], current:caller, previous:nil)
		bob.Do(peter.TopFuncFnNoncrossing())        // FAIL: caller != bob
		// ----------------------------------------
		// CASE pA3: function declared in func(cur realm, ...) (crossing) -- illegal crossing function
		bob.Do(peter.TopFuncFnCrossing(cross))      // INVALID (illegal crossing function)

		// --------------------------------------------------------------------------------
		// CASE pB1: method has nil receiver -- do not storage-cross
		// (identical to rB1 for consistency)
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do((*bob.Object)(nil).Method)           // FAIL: caller != bob
		// ----------------------------------------
		// CASE pB2: method has nil receiver (external) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do((*peter.Object)(nil).Method)         // FAIL: caller != bob
		// ----------------------------------------
		// CASE pB3: method has unreal receiver (inline) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do(new(peter.Object).Method)            // FAIL: caller != bob
		// ----------------------------------------
		// CASE pB4: method has unreal receiver (var) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(peter.Object)
		bob.Do(obj.Method)                          // FAIL: caller != bob

		// --------------------------------------------------------------------------------
		// CASE pC1: method has real receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[bob], current:caller, previous:nil)
		obj := new(peter.Object)
		bob.SetObject(obj)                          // obj persisted in bob
		bob.Do(obj.Method)                          // SUCCESS: persisted in bob
		// ----------------------------------------
		// CASE pC2: method has real receiver (external) -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],      current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[alice], current:caller, previous:nil)
		obj := new(peter.Object)
		alice.SetObject(obj)                        // obj persisted in alice
		bob.Do(obj.Method)                          // FAIL: resides in alice (needs bob)
		// ----------------------------------------
		// CASE pC3: method has real receiver via closure -- storage-cross to receiver
		obj := new(peter.Object)
		cls := func() {
			print(obj)
		}
		bob.SetClosure(cls)                         // obj persisted in bob via cls
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do(obj.Method)                          // SUCCESS: resides in bob
		// ----------------------------------------
		// CASE pC4: method has real receiver via closure (external) -- storage-cross to receiver
		obj := new(peter.Object)
		cls := func() {
			print(obj)
		}
		alice.SetClosure(cls)                       // obj persisted in alice via cls
		//   - in bob.Do:                           (storage:[],      current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[alice], current:caller, previous:nil)
		bob.Do(obj.Method)                          // FAIL: alice != bob

	}

	//----------------------------------------
	// in /r/alice:
	package alice
	import `bob`
	func TopFunc() { // CASE rA1
		bob.PrivateFunc()
	}
	func TopFuncFnNoncrossing() {
		return func() { // CASE rA2
			bob.PrivateFunc()
		}
	}
	func TopFuncFnCrossing(cur realm) {
		return func() { // CASE rA3
			bob.PrivateFunc()
		}
	}
	type Object struct {}
	func (obj *Object) Method() { // CASE rB2~4, CASE rC1~5
		bob.PrivateFunc()
	}
	var object *Object 
	func SetObject(obj any) { object = obj } // CASE rC2, CASE pC2
	var closure func() 
	func SetClosure(cls func()) { closure = cls } // case rC4~5, CASE pC4~5

	//----------------------------------------
	// in /p/peter:
	package peter
	import `bob`
	func TopFunc() { // CASE pA1
		bob.PrivateFunc()
	}
	func TopFuncFnNoncrossing() {
		return func() { // CASE pA2
			bob.PrivateFunc()
		}
	}
	/* ILLEGAL CROSSING FUNCTION
	func TopFuncFnCrossing(cur realm) {
		return func() { // CASE pA3
			bob.PrivateFunc()
		}
	}*/
	type Object struct {}
	func (obj *Object) Method() { // CASE pB2~4, CASE pC1~5
		bob.PrivateFunc()
	}

	//----------------------------------------
	// in /r/bob:
	package bob
	import `alice`
	func Do(fn func()) {
		// Normal non-crossing function doesn't change the current or
		// previous realm.
		// Should behave the same as if /p/bob.Do(), or the same as if
		// fn() directly.
		fn()
	}
	func DoCrossing(cur realm, fn func()) {
		fn()
	}
	var private int
	func PrivateFunc() {
		// This function is exposed but is not crossing, so fails
		// unless current realm is bob.
		// Generally not useful to expose such a "private function",
		// but can be useful in some circumstances.
		// TODO document when it is useful.
		private = -1
	}
	type Object struct {}
	func (obj *Object) Method() { // CASE rB1, CASE pB1
		// The effect of the three lines below should be the same.
		// They can only succeed if present realm is already 
		// They can only suc// bob and obj is nil or unreal; or if obj is attached
		// to bob (thus making the present realm bob by storage-realm).
		private = -3
		obj.method2()
		PrivateFunc()
	}
	func (obj *Object) method2() {
		private = -2
	}
	var object *Object 
	func SetObject(obj any) { object = obj }  // CASE rC1, CASE pC1
	var closure func()
	func SetClosure(cls func()) { closure = cls } // CASE rC3, CASE pC3
```
