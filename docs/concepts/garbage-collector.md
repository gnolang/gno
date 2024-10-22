# Garbage Collector Specification for GNO


### Overview
The garbage collector follows a mark-and-sweep approach to manage memory in Gno.

It tracks and collects objects on the heap that are no longer reachable from any root reference, thereby freeing memory used by unreachable objects. The GC operates in two phases: marking and sweeping.

The garbage collector is only running if we are about to go out of memory.

All root objects are tied to the scope they are defined in. When the scope ends, the associated root objects are dropped.

When all roots for a particular heap allocated object are dropped, it effectively becomes garbage which would be collected during a collection cycle that would be triggered in the above described way.

The garbage collector works in a deterministic manner.

Below is the API associated with the garbage collector.

###  Heap and Objects
####  Heap

   The Heap structure is responsible for managing all the objects in the system. It maintains:

   •	objects: A list of all objects currently allocated on the heap.

   •	roots: A list of root objects, which are entry points for reachability analysis during the mark phase.

   GcObj (Heap Object)
   Each GcObj represents an object on the heap, and it contains:

   •	marked: A boolean flag indicating whether the object is reachable (used during the mark phase).

   •	tv: The actual object data encapsulated in a TypedValue.

   •	refs: A slice of pointers to other GcObjs that this object references, which are traversed during garbage collection.

### Object Creation
   •	NewObject: Initializes a new object (GcObj) on the heap, wrapping the provided TypedValue.
   •	MakeHeapObj: Creates a heap object from a TypedValue if it is a type that requires heap allocation (e.g., slices, structs, arrays, strings).
   •	AddObject: Adds a newly created GcObj to the heap's objects list.
### Root Management
   •	AddRoot: Adds a GcObj to the roots list, ensuring it will be scanned during the mark phase.
   •	RemoveRoot: Removes a GcObj from the roots list if certain conditions are met (e.g., if it references the TypedValue being removed).
### Reference Management
   •	AddRef: Establishes a reference from one object to another, ensuring that a root object or any heap object can track its dependencies.
   •	FindObjectByTV: Finds a heap object by its TypedValue. This is essential for managing object relationships and references.
### Garbage Collection Algorithm
   Mark Phase
   The mark phase identifies all reachable objects by starting from the root objects and recursively marking all objects reachable from them.
   •	Mark Process (mark):
   ◦	For each root in roots, the mark function checks whether the object has been visited (using the marked flag).
   ◦	If not marked, the object is marked as reachable and the function recursively follows the references (refs) to mark all reachable objects.
###   Sweep Phase
   The sweep phase frees the memory for unmarked objects (those that were not reachable in the mark phase).
   •	Sweep Process (sweep):
   ◦	Traverses the objects list and separates marked (reachable) and unmarked (unreachable) objects.
   ◦	Unmarked objects are removed from the heap, and their memory is freed.
   ◦	Marked objects are unmarked in preparation for the next GC cycle.
### GC Trigger
   The garbage collection process is initiated by calling the MarkAndSweep method. This method runs both the mark and sweep phases to clean up unused heap memory:
   1	Mark: Calls the mark function on all root objects to mark reachable objects.
   2	Sweep: Calls the sweep function to remove and return all unmarked objects (garbage).

	


