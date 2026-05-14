# Exercises in Interrealm Cases

These are an attempt to go over some cases systematically.  It may help to go
over these cases, but they are rather repetitive.  In the future as more cases
are considered, they should be added here systematically, and also duplicated
in gno.land/pkg/integration/testdata/interrealm_final.txtar.  That file should
ideally be transcribed as a test file that makes use of revive() and "testing".

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
   - CASE rA1: function declared at package level                   -- do not storage-cross
   - CASE rA2: function declared in func(...) (non-crossing)        -- do not storage-cross
   - CASE rA3: function declared in func(cur realm, ...) (crossing) -- do not storage-cross

 * fn is method declared in /r/realm (nil or unreal receiver):
   - CASE rB1: method has nil receiver                              -- do not storage-cross
   - CASE rB2: method has nil receiver (external)                   -- do not storage-cross
   - CASE rB3: method has unreal receiver (inline)                  -- do not storage-cross
   - CASE rB4: method has unreal receiver (var)                     -- do not storage-cross

 * fn is method declared in /r/realm (real receiver):
   - CASE rC1: method has real receiver                             -- storage-cross to receiver
   - CASE rC2: method has real receiver (external)                  -- storage-cross to receiver
   - CASE rC3: method has real receiver via closure                 -- storage-cross to receiver
   - CASE rC4: method has real receiver via closure (external)      -- storage-cross to receiver
   - CASE rC5: method has real receiver via closure too late        -- do not storage-cross

 * fn is crossing function declared in /r/realm:
   - CASE rD1: direct modification from external closure            -- illegal modification

----------------------------------------
## fn is declared in /p/package:

These cases are similar to /r/realm except for CASE rA3 and rD1 since
/p/package cannot contain crossing functions.

Note that the test cases in interrealm_final.txtar are similar but modified to
account for the difference between realm and pure package.

 * fn is function declared in /p/package:
   - CASE pA1: function declared at package level                   -- do not storage-cross
   - CASE pA2: function declared in func(...) (non-crossing)        -- do not storage-cross
   - CASE pA3: function declared in func(cur realm, ...) (crossing) -- illegal crossing function

 * fn is method declared in /p/package (nil or unreal receiver):
   - CASE pB1: method has nil receiver                              -- do not storage-cross
   - CASE pB2: method has nil receiver (external)                   -- do not storage-cross
   - CASE pB3: method has unreal receiver (inline)                  -- do not storage-cross
   - CASE pB4: method has unreal receiver (var)                     -- do not storage-cross

 * fn is method declared in /p/package:
   - CASE pC1: method has real receiver                             -- storage-cross to receiver
   - CASE pC2: method has real receiver (external)                  -- storage-cross to receiver
   - CASE pC3: method has real receiver via closure                 -- storage-cross to receiver
   - CASE pC4: method has real receiver via closure (external)      -- storage-cross to receiver
   - CASE pC5: method has real receiver via closure too late        -- do not storage-cross

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
		// ----------------------------------------
		// CASE rC5: method has real receiver via closure too late -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(alice.Object)
		cls := func() {
			print(obj)
			bob.Do(obj.Method)                      // FAIL: not yet persisted
			bob.SetClosure(cls)                     // obj persisted to bob too late
			// this would work otherwise:
			// bob.Do(obj.Method)                   
		}
		cls()

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
		// ----------------------------------------
		// CASE pC5: method has real receiver via closure too late -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(peter.Object)
		cls := func() {
			print(obj)
			bob.Do(obj.Method)                      // FAIL: not yet persisted
			alice.SetClosure(cls)                   // obj persisted to bob too late
			// this would work otherwise:
			// bob.Do(obj.Method)              
		}
		cls()

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
