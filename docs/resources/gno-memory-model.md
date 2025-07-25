# Gno Memory Model

## The Typed Value

```go
type TypedValue struct {
        T Type    
        V Value   
        N [8]byte 
}
```

Both `Type` and `Value` are Go interface values.  Go does not support union
types, so primitive values like bools and ints are stored in the N field for
performance.

All values in Gno are stored represented as (type, value) tuples.  Vars in
scope blocks, fields in structs, elements in arrays, keys and values of maps
are all respresented by the same `TypedValue` struct.

This tuple representation lets the Gno VM implementation logic be simpler with
less code.  Reading and writing values are the same whether the static type of
the value is an interface or something concrete with no special logic for
memory optimizations which is less relevant in a massive multi-user
transactional operating system where most of the data resides in disk anyways.

Another benefit is that it promotes the development of new types of client
interfaces that can make use of the type information for object display and
interaction. The original vision of Tim Burners Lee's HTML DOM based World Wide
Web with restful HTTP requests like GET and POST has been steamrolled over by
continuous developments in HTML, CSS, Javascript, and the browser. The internet
is alive but the World Wide Web is dead. Gno is a reboot of the original vision
of the Web but better because everything is integrated based on a singular
well designed object-oriented language. Instead of POST requests Gno has 
typed method calls. All values are annotated with their types making the
environment REST-ful to the core.

> REST (Representational State Transfer) is a software architectural style that
> was created to describe the design and guide the development of the
> architecture for the World Wide Web. REST defines a set of constraints for
> how the architecture of a distributed, Internet-scale hypermedia system, such
> as the Web, should behave. - wikipedia

While the Gno VM implements the memory model described here in the most
straightforward way, future alternative implementations may represent values
differently in machine memory even while conforming to the spec as implemented
by the Gno VM.


## Objects and Values

There are many types of Values, and some of these value types are also Objects.
The types below are all values and those that are bolded are also objects.

 * Primitive            // bool, uint, int, uint8, ... int64
 * StringValue
 * BigintValue          // only used for constant expressions
 * BigdecValue          // only used for constant expressions
 * DataByteValue        // invisible type for byte array optimization
 * PointerValue         // base is always an object
 * **ArrayValue**
 * SliceValue
 * **StructValue**
 * **FuncValue**
 * **MapValue**
 * **BoundMethodValue** // func & receiver
 * TypeValue
 * PackageValue
 * **BlockValue**       // for package, file, if, range, switch, func
 * RefValue             // reference to an object stored in disk
 * **HeapItemValue**    // invisible type for loopvars and closure captures


## Pointers

```go
type PointerValue struct {
        TV    *TypedValue // escape val if pointer to var.
        Base  Value       // array/struct/block, or heapitem.
        Index int         // list/fields/values index, or -1 or -2 (see below).
}
```

The pointer is a reference to a typed value slot in an array, struct, block,
or heap item. It is also used internally in the VM for assigning to slots.
Even internally the Base is used to tell the realm finalizer when the base
has been updated.

## Blocks and Heap Items

All statements enclosed in {} parentheses will allocate a new block
and push it onto the block stack of the VM. The size of the block
is determined by the number of variables declared in the block statement.

```
type BlockValue struct {
	ObjectInfo            
	Source   BlockNode
	Values   []TypedValue
	Parent   Value         // Parent block if any, or RefValue{} to one.
	Blank    TypedValue    // Captures "_" underscore names.
	bodyStmt bodyStmt      // Holds a pointer to the current statement.
}
```

The following Gno AST nodes when executed will create a new block:

 * FuncLitStmt
 * BlockStmt    // a list of statements wrapped in {} 
 * ForStmt
 * IfCaseStmt
 * RangeStmt
 * SwitchCaseStmt
 * FuncDecl
 * FileNode
 * PackageNode

`IfStmt`s and `SwitchStmt`s also produce faux blocks that get merged onto the
following `IfCaseStmt` and `SwitchCaseStmt` respectively, but this is an
invisible implementation detail and the behavior may change.

Heap items are only used in blocks. Conceptually they are an object container
around a singleton typed value slot. It is not visible to the gno developer
but it is important to understand how they work when inspecting the block
space. 

```go
func Example(arg int) (res *int) {
	var x int = arg + 1
	return &x
}
```

The above code when executed will first produce the following block:

```
BlockValue{
    ...
    Source: <*FuncDecl node>,
    Values: [
        {T: nil, V: nil},                      // 'arg' parameter
        {T: nil, V: nil},                      // 'res' result
        {T: HeapItemType{},
	 V: &HeapItemValue{{T: nil, V: nil}},  // 'x' variable
    ],
    ...
}
```

In the above example the third slot for `x` is not initialized to the zero
value of a typed value slot, but rather it is prefilled with a heap item. 

Variables declared in a closure or passed by reference are first discovered and
marked as such from the preprocessor, and NewBlock() will prepopulate these
slots with `*HeapItemValues`.  When a `*HeapItemValue` is present in a block
slot it is not written over but instead the value is written into the heap
item's slot. 

When the example code executes `return &x` instead of returning a
`PointerValue` with `.Base` set to the `BlockValue` and `.Index` of 2, it sets
`.Base` to the `*HeapItemValue` with `.Index` of 0 since a heap item only
contains one slot. The pointer's `.TV` is set to the single slot of of the heap
item. This way the when the pointer is used later in another transaction there
is no need to load the whole original block value, but rather a single heap
item object. If `Example()` returned only `x` rather than a pointer `&x` it
would not be initialized with a heap item for the slot.

```go
func Example2(arg int) (res func()) {
	var x int = arg + 1
	return func() {
		println(x)
	}
}
```

The above example illustrates another use for heap items. Here we don't
reference `x`, but it is captured by the anonymous function literal (closure).
At runtime the closure `*FuncValue` captures the heap item object such that the
closure does not depend on the block at all.

Variables declared at the package (global) level may also be referred to by
pointer in anonymous functions. In the future we will allow limited upgrading
features for mutable realm packages (e.g. the ability to add new functions or
replace or "swizzle" existing ones), so all package level declared variables
are wrapped in heap item objects.

Since all global package values referenced closures can be captured as heap
objects, the execution and persistence of a closure function value does not
depend on any parent blocks. (Omitted here is how references to package
level declared functions and methods are replaced by a selector expression
on the package itself; otherwise closures would still in general depend
on their parent blocks).

## Loopvars

Go1.22 introduced loopvars to reduce programmer errors. Gno uses 
heap items to implement loopvars.

```go
for _, v := range values {
    saveClosure(func() {
        fmt.Println(v)
    })
}
```

The Gno VM does something special for loopvars. Instead of assigning the new
value `v` to the same slot, or even the same heap item object's slot, it
replaces the existing heap item object with a new one. This allows the closure
to capture a new heap item object with every iteration. This is called 
a heap definition.

The behavior is applied for implicit loops with `goto` statements. The
preprocessor first identifies all such variable definitions whether explicit in
range statements or implicit via `goto` statements that are captured by
closures or passed by pointer reference, and directs the VM to execute the
define statement by replacing the existing heap item object with a new one.
