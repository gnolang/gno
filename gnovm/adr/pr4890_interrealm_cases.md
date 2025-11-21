# When calling func(..., fn func, ...):

 - option 1: do not storage-cross
 - option 2: storage-cross to declared as storage
 - option 3: storage-cross to receiver
 - option 4: storage-cross to virtual-realm (of /p/package)

"Storage-crossing" is like (regular) crossing for method calls except it
does not affect CurrentRealm()/PreviousRealm(). Since method calls can
be nested the storage-realms are represented as a stack which maybe
empty.

A "virtual-realm" is used when calling a method of a nil receiver
declared in a /p/package (/p/package is immutable). Two
virtual-realms of the same /p/package are identical and are immutable.

When calling a method of a nil receiver declared in an /r/realm we
storage-cross to the storage realm /r/realm.

TODO "implicit-cross" is a misnomer, replace all w/ "storage-cross".

----------------------------------------
## fn is declared /r/realm:

 * fn is (non-crossing) function declared in /r/realm:
   - CASE rA1: function declared at package level                   -- do not storage-cross
   - CASE rA2: function declared in func(...) (non-crossing)        -- do not storage-cross
   - CASE rA3: function declared in func(cur realm, ...) (crossing) -- do not storage-cross

 * fn is method declared in /r/realm (nil/unattached receiver):
   - CASE rB1: method has nil receiver                              -- storage-cross to declared
   - CASE rB2: method has nil receiver                              -- storage-cross to (wrong) declared
   - CASE rB3: method has unattached receiver (inline)              -- do not storage-cross
   - CASE rB4: method has unattached receiver (var)                 -- do not storage-cross

 * fn is method declared in /r/realm (attached receiver):
   - CASE rC1: method has attached receiver                         -- storage-cross to receiver
   - CASE rC2: method has (wrong) attached receiver                 -- storage-cross to (wrong) receiver
   - CASE rC3: method has attached receiver via closure             -- storage-cross to receiver
   - CASE rC4: method has (wrong) attached receiver via closure     -- storage-cross to (wrong) receiver
   - CASE rC5: method has attached receiver via closure late        -- do not storage-cross

 * fn is crossing function declared in /r/realm:
   - CASE rD1: closure attached and run crossing selector           -- XXX 
   - CASE rD1: closure attached and run crossing arg                -- XXX

----------------------------------------
## fn is declared in /p/package:
(similar to /r/realm except for CASE pA3 and CASE pB2)

 * fn is function declared in /p/package:
   - CASE pA1: function declared at package level                   -- do not storage-cross (same)
   - CASE pA2: function declared in func(...) (non-crossing)        -- do not storage-cross (same)
   - CASE pA3: function declared in func(cur realm, ...) (crossing) -- illegal crossing function

 * fn is method declared in /p/package:
   - CASE pB1: method has nil receiver                              -- storage-cross to declared (identical)
   - CASE pB2: method has nil receiver                              -- storage-cross to (wrong, virtual-realm) declared
   - CASE pB3: method has unattached receiver (inline)              -- do not storage-cross (same)
   - CASE pB4: method has unattached receiver (var)                 -- do not storage-cross (same)

 * fn method declared in /p/package:
   - CASE pC1: method has attached receiver                         -- storage-cross to receiver (same)
   - CASE pC2: method has (wrong) attached receiver                 -- storage-cross to (wrong) receiver (same)
   - CASE pC3: method has attached receiver via closure             -- storage-cross to receiver (same)
   - CASE pC4: method has (wrong) attached receiver via closure     -- storage-cross to (wrong) receiver (same)
   - CASE pC5: method has attached receiver via closure late        -- do not storage-cross (same)

----------------------------------------
## NOTEs

 * (Storage) attachment is like tainting w/ colors.
 * Assigning to a field of an already-attached object results in
   immeediate recursive attachment to the storage realm of the parent.
 * Assigning to a field of a yet unattached object does NOT result
   in immediate recursive attachement; it may happen during finalization.
 * After finalization all objects are attached to one realm or another.
 * A variable declaration via `:=` or `var` does not result in
   immediate attachment. That is, block values are not by default attached.
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
		//
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
		//
		// --------------------------------------------------------------------------------
		// CASE rB1: method has nil receiver -- storage-cross to declared
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do((*bob.Object)(nil).Method)          // SUCCESS: storage-crossed to bob
		// ----------------------------------------
		// CASE rB2: method has nil receiver -- storage-cross to (wrong) declared
		//   - in bob.Do:                           (storage:[],      current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[alice], current:caller, previous:nil)
		bob.Do((*alice.Object)(nil).Method)        // FAIL: alice != bob
		// ----------------------------------------
		// CASE rB3: method has unattached receiver (inline) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do(new(alice.Object).Method)           // FAIL: caller != bob
		// ----------------------------------------
		// CASE rB4: method has unattached receiver (var) -- do not storage-cross (same as rB3 above)
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(alice.Object)
		bob.Do(obj.Method)                          // FAIL: caller != bob
		//
		// --------------------------------------------------------------------------------
		// CASE rC1: method has attached receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[bob], current:caller, previous:nil)
		obj := new(alice.Object)
		bob.SetObject(obj)                          // obj attached to bob
		bob.Do(obj.Method)                          // SUCCESS: attached to bob
		// ----------------------------------------
		// CASE rC2: method has (wrong) attached receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[alice], current:caller, previous:nil)
		obj := new(alice.Object)
		alice.SetObject(obj)                        // obj attached to alice
		bob.Do(obj.Method)                          // FAIL: attached to alice (but need bob)
		// ----------------------------------------
		// CASE rC3: method has attached receiver via cls -- storage-cross to receiver
		obj := new(alice.Object)
		cls := func() {
			print(obj)
		}
		bob.SetClosure(cls)                         // obj recursively attached via cls to bob
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do(obj.Method)                          // SUCCESS: attached to bob
		// ----------------------------------------
		// CASE rC4: method has (wrong) attached receiver via cls -- stroage-cross to (wrong) receiver
		obj := new(alice.Object)
		cls := func() {
			print(obj)
		}
		alice.SetClosure(cls)                       // obj recursively attached via cls to alice
		//   - in bob.Do:                           (storage:[],      current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[alice], current:caller, previous:nil)
		bob.Do(obj.Method)                          // FAIL: alice != bob
		// ----------------------------------------
		// CASE rC5: method has attached receiver via cls late -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(alice.Object)
		cls := func() {
			bob.Do(obj.Method)                      // FAIL: not yet unattached
			bob.SetClosure(cls)                     // obj recursively attached to bob late
			bob.Do(obj.Method)                      // SUCCESS: attached to bob
		}
		cls()
		//
		// --------------------------------------------------------------------------------
		// CASE rD1: XXX closure attached to bob
		//   - in bob.ExecuteClosure:               (storage:[], current:bob, previous:caller)
		f := func() {
			bob.AllowedList = append(bob.AllowedList, 3)
			println("all right")
		}
		
		bob.SetClosure(cross, f)
		bob.ExecuteClosure(cross)
		//
		// ================================================================================
		// ## fn is declared /p/package: (similar to /r/realm except for CASE pA3 and CASE pB2)
		//
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
		// CASE pA3: function declared in func(cur realm, ...) (crossing) -- omitted (illegal crossing function)
		// bob.Do(peter.TopFuncFnCrossing(cross))   // INVALID (illegal crossing function)
		//
		// --------------------------------------------------------------------------------
		// CASE pB1: method has nil receiver -- storage-cross to declared as storage
		// (identical to rB1 for consistency)
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do((*bob.Object)(nil).Method)           // SUCCESS: storage-cross to bob
		// ----------------------------------------
		// CASE pB2: method has nil receiver -- storage-cross to (wrong, virtual-realm) declared as storage NOTE (differs)
		//   - in bob.Do:                           (storage:[],              current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[peter.virtual], current:caller, previous:nil)
		bob.Do((*peter.Object)(nil).Method)         // FAIL: peter.virtual != bob
		// ----------------------------------------
		// CASE pB3: method has unattached receiver (inline) -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		bob.Do(new(peter.Object).Method)            // FAIL: caller != bob
		// ----------------------------------------
		// CASE pB4: method has unattached receiver (var) -- do not storage-cross (same as pB3 above)
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(peter.Object)
		bob.Do(obj.Method)                          // FAIL: caller != bob
		//
		// --------------------------------------------------------------------------------
		// CASE pC1: method has attached receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],    current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[bob], current:caller, previous:nil)
		obj := new(peter.Object)
		bob.SetObject(obj)                          // obj attached to bob
		bob.Do(obj.Method)                          // SUCCESS: attached to bob
		// ----------------------------------------
		// CASE pC2: method has (wrong) attached receiver -- storage-cross to receiver
		//  - in bob.Do:                            (storage:[],      current:caller, previous:nil)
		//  - in obj.Method:                        (storage:[alice], current:caller, previous:nil)
		obj := new(peter.Object)
		alice.SetObject(obj)                        // obj attached to alice
		bob.Do(obj.Method)                          // FAIL: attached to alice (but need bob)
		// ----------------------------------------
		// CASE pC3: method has attached receiver via cls -- storage-cross to receiver
		obj := new(peter.Object)
		cls := func() {
			print(obj)
		}
		bob.SetClosure(cls)                         // obj recursively attached via cls to bob
		//   - in bob.Do:                           (storage:[],    current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[bob], current:caller, previous:nil)
		bob.Do(obj.Method)                          // SUCCESS: attached to bob
		// ----------------------------------------
		// CASE pC4: method has (wrong) attached receiver via cls -- storage-cross to (wrong) receiver
		obj := new(peter.Object)
		cls := func() {
			print(obj)
		}
		alice.SetClosure(cls)                       // obj recursively attached via cls to alice
		//   - in bob.Do:                           (storage:[],      current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[alice], current:caller, previous:nil)
		bob.Do(obj.Method)                          // FAIL: alice != bob
		// ----------------------------------------
		// CASE pC5: method has attached receiver via cls late -- do not storage-cross
		//   - in bob.Do:                           (storage:[], current:caller, previous:nil)
		//   - in obj.Method:                       (storage:[], current:caller, previous:nil)
		obj := new(peter.Object)
		cls := func() {
			bob.Do(obj.Method)                  // FAIL: not yet unattached
			alice.SetClosure(cls)               // obj recursively attached via cls to alice late
			bob.Do(obj.Method)                  // SUCCESS: attached to bob
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
		bob.PrivateFunc()
	}
	var object *Object 
	func SetObject(obj any) { object = obj }  // CASE rC1, CASE pC1
	var closure func()
	func SetClosure(cls func()) { closure = cls } // CASE rC3, CASE pC3
```
