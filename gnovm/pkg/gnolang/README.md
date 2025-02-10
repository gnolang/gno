# Gnolang

The GnoVM is a stack-based virtual machine that interprets Gno code. It interprets the Go programming language, with care for executing code deterministically for execution on a blockchain.

This file aims to provide an overview of the inner functioning of the virtual machine for those who wish to dive into its development, with references to the general files and functions where functionality is implemented.

Useful background reading to understand some of the terminology used in this document is the [Go specification](https://go.dev/ref/spec).

## Overview

The Virtual Machine is internally represented with the `Machine` struct (this is slightly simplified):

```go
type Machine struct {
	// State
	Ops        []Op          // main operations
	Values     []TypedValue  // buffer of values to be operated on
	Exprs      []Expr        // pending expressions
	Stmts      []Stmt        // pending statements
	Blocks     []*Block      // block (scope) stack
	Frames     []*Frame      // func call stack
	Package    *PackageValue // active package
	Realm      *Realm        // active realm
	Alloc      *Allocator    // memory allocations
	Exceptions []Exception
	NumResults int   // number of results returned
	Cycles     int64 // number of "cpu" cycles

	// Configuration
	PreprocessorMode bool // this is used as a flag when const values are evaluated during preprocessing
	ReadOnly         bool
	MaxCycles        int64
	Output           io.Writer
	Store            Store
	Context          interface{}
	GasMeter         store.GasMeter
	// ...
}
```

Here's how the GnoVM can be instructed to parse a simple expression, like `main()`:

```go
func (m *Machine) RunMain() {
	// This is a short-hand to generate an AST of a statement like "main()".
	s := S(Call(X("main")))
	// We preprocess (compile) s, so that the "main" name is
	// linked to an actual function in the file being executed.
	sn := m.LastBlock().GetSource(m.Store)
	s = Preprocess(m.Store, sn, s).(Stmt)
	// Push Halt and Exec on the Op-stack; Exec pops the latest Stmt
	// (which will be s) and executes it.
	m.PushOp(OpHalt)
	m.PushStmt(s)
	m.PushOp(OpExec)
	// Run executes the code.
	m.Run()
}
```

`Run` runs a for loop that pops an `Op` from the top of the stack, which determines what the Virtual Machine will do next. Generally:

- Statements are broken down into their individual operations, which often involve one or more expressions.
- The expressions are also pushed to the top of the `Exprs` stack. Each `Expr` is evaluated into a number of values (generally one, with some exceptions like function calls which return multiple or no values).
- The values, represented as a `TypedValue`, are then operated upon using each `Op`, which generally does one of the many basic transformation defined by the language (unary and binary operators, but also more "complex" operations like resolving composite expressions or creating struct values from struct literals).

In other words, the GnoVM is a kind of mix between a model of an interpreter and virtual machine:

- Like a virtual machine, the basic tool to perform operations on data is the "operation".
- Like an interpreter, it mostly executes code by compiling on-the-fly an AST expression into its individual operations.
- Like a virtual machine, it works on a modified and "preprocessed" (ie. compiled) version of the original AST, which enrichens the existing AST with type information and simplifies some code when deemed necessary or too complex to implement directly in the virtual machine.
- Like an interpreter, it retains much of the type information also at run-time, to ensure the correctness of the expressions it's executing.

We can say that the GnoVM is a virtual machine which works off of a "bytecode", which is the preprocessed source code. In many ways, the process of preprocessing is a kind of compilation, as it brings the AST into an "executable state".

On top of this baseline understanding for the execution, the GnoVM makes heavy use of the "Store". The Store is the "glue" layer between the virtual machine and the external world; either the blockchain node or a local filesystem being worked on.

In the following sections, we delve into greater detail in the steps involved for executing Gno code.

## Parsing

To parse Go code, the GnoVM uses the [`go/parser`](https://pkg.go.dev/go/parser) package. The parsed AST is re-organized into Gno's own AST, which is slightly different from the original.

Some example of differences:

- `panic` in Gno has its own statement, `PanicStmt`, while in Go it is parsed as a `CallExpr`.
- Go `ParenExpr` are omitted in the Gno representation.
- Array type expressions (`[]T`, `[...]T` and `[64]T`) are split between array type and slice type.
- Reference expressions (`&name`) are their own type - `RefExpr` - instead of being aggregated into `UnaryExpr` (as most other unary operators).

Each node type of the Gno AST implements the [Node](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Node) interface, and ends with `Expr`, `Stmt`, `Decl` (import, value and type declarations) or `Node` (`PackageNode` and `FileNode`).

- The node types all embed the [`Attributes`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Attributes) struct, which contains line/column information, any label placed on a statement, and any non-persisted attributes which are set during compilation.
- Aside from what is placed inside the attributes, there are other parts of the AST which are properly filled only during the preprocessing stage. One such example is [`NameExpr.Path`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#NameExpr.Path), which contains the `ValuePath` to resolve a name to an actual `TypedValue` in a block. (These will be discussed in the Compilation section.)
- A node may also embed the [`StaticBlock`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#StaticBlock) type, when the node creates a new scope and as such has its own names.

#### References

- The parsing of a source file happens in [`ParseFile`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#ParseFile), which uses [`Go2Gno`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Go2Gno) - the AST conversion function.
- Go2Gno can generally be used to inspect the differences with the Go AST.

## Compilation

To execute code, the GnoVM predominantly uses a transformed version of the AST, which is effectively the "bytecode" of the VM. Additionally, there are data structures which are used to persist "values", which are then kept in the [Store](#Store).

To better study and analyze all the steps involved in "compilation", we'll take a look at what happens in a more complex scenario than what was shown in the above `RunMain` example, namely what happens when we execute [`RunMemPackage`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Machine.RunMemPackage), and more specifically in the `runFileDecls` method. `RunMemPackage` is used when we add new packages on-chain, and when we load a package locally, and as such is a quite useful reference.

The goals of the compilation step are various. Here's a non-exhaustive list:

- Resolving all symbols (like `*NameExpr`, but also selector expressions, and others) into a ValuePath; ie. converting names into "pointers" to values in a `Block` or other data structures.
- Pre-evaluating the values of all constant expressions.
- Performing type checking, matching the rules of the Go specification, not only ensuring type checking but ensuring correctness of all operators, disallowing variable re-declaration, and rejecting all sorts of invalid code.

There are a few steps involved in the compilation stage. Many of these use the `Transcribe` function to traverse the AST, which we describe in the next section. After that, we'll take a closer look at each one of the steps of the compilation stage.

### Transcribe

To perform compilation, the [`Transcribe`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Transcribe) function is regularly used. This is a [tree traversal](https://en.wikipedia.org/wiki/Tree_traversal) function. Its functioning can be summarized as follows:

- The `Transform` function is called with the `TransStage` set to `TRANS_ENTER`.
- When entering Block statements, `Transform` is called, with `TransStage` set to `TRANS_BLOCK`.
- The "child statements" are recursively `Transcribe`'d in order, for example:
	- A `*BinaryExpr` will transcribe first the expression on the left, then the one on the right.
	- A `*ForStmt` will transcribe, in order: the initialization statement (`i := 0`), the condition (`i < 256`), the post iteration statement (`i++`), then each statement in the for block's body.
- After transcribing all the child statements, `Transform` is called once again, with `TransStage` set to `TRANS_LEAVE`.

Every time the `Transform` function is called, it knows:

- `ns`: the stack of nodes up to the one being transcribed.
- `ftype`: the "context" in which the transform function is being called (is this statement an if initialization condition? is it a top level declaration? etc.)
- `index`: the index of the expressions in the containing list, for instance in a block or in a composite literal.
- `n`: the node that should be transformed.
- `stage`: the transformer stage, as described above.

The transformer then returns the transformed node, and a value controlling how the transformer should continue its execution:

- `TRANS_CONTINUE` visits all children recursively (and is used most often).
- `TRANS_SKIP` avoids processing further children.
- `TRANS_EXIT` aborts the execution of the transcriber.

Aside from these, it is worth mentioning a common pattern often used in callers of `Transcribe`, to keep track of blocks. Before calling `Transcribe`, a stack of `BlockNode` is initialized:

```go
var stack []BlockNode = make([]BlockNode, 0, 32)
var last BlockNode = ctx
stack = append(stack, last)
```

(Note that all compiling functions require a `ctx` BlockNode to work on; often, at the top level, this is the `*PackageNode`).

When `TRANS_BLOCK` is called,`pushInitBlock(n, &last, &stack)` is then generally called to append a new block on top of the stack. Then, at `TRANS_LEAVE`, all nodes that implement `BlockNode` pop the latest block from the stack:

```go
stack = stack[:len(stack)-1]
last = stack[len(stack)-1]
```

This way, we can keep track of the blocks we're working on, which we'll see is very important, for instance in determining ValuePaths.
### Predefining names

[`PredefineFileSet`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#PredefineFileSet) is the very first proper "compilation" function called in `runFileDecls`.  Our `RunMain()` example above does not use it, because it is only called to initialize the top-level declarations in a `*FileSet`, which can have cross-file dependencies or be self-referencing types.

- initStaticBlocks
- `predefineNow` on imports, type, value decls + `tryPredefine`

### Preprocess

- Statement handling
	- If/else if/else statements (blocks that build on the last, gradually)
	- faux blocks / frames for switch and for.
	- Decomposition of complex assignments
- Conversion into ValuePaths
- Handling of ConstExpr, constTypeExpr
- Loop definition.
- evalStaticTypeOf

## Execution

The execution of GnoVM programs happens in the central [`Run`](https://gnolang.github.io/gno/github.com/gnolang/gno/gnovm/pkg/gnolang.html#Machine.Run) function. To determine what to do next, the Machine pops the latest operation from the pop of the op stack.

It is useful to take a quick look at all of the op-codes, which help us generally define the kinds of operations that can happen on the VM:

```go
const (
	/* Control operators */
	OpInvalid             Op = 0x00 // invalid
	OpHalt                Op = 0x01 // halt (e.g. last statement)
	OpNoop                Op = 0x02 // no-op
	OpExec                Op = 0x03 // exec next statement
	OpPrecall             Op = 0x04 // sets X (func) to frame
	OpCall                Op = 0x05 // call(Frame.Func, [...])
	OpCallNativeBody      Op = 0x06 // call body is native
	OpReturn              Op = 0x07 // return ...
	OpReturnFromBlock     Op = 0x08 // return results (after defers)
	OpReturnToBlock       Op = 0x09 // copy results to block (before defer)
	OpDefer               Op = 0x0A // defer call(X, [...])
	OpCallDeferNativeBody Op = 0x0B // call body is native
	OpSwitchClause        Op = 0x0E // exec next switch clause
	OpSwitchClauseCase    Op = 0x0F // exec next switch clause case
	OpTypeSwitch          Op = 0x10 // exec type switch clauses (all)
	OpIfCond              Op = 0x11 // eval cond
	OpPopValue            Op = 0x12 // pop X
	OpPopResults          Op = 0x13 // pop n call results
	OpPopBlock            Op = 0x14 // pop block NOTE breaks certain invariants.
	OpPopFrameAndReset    Op = 0x15 // pop frame and reset.
	OpPanic1              Op = 0x16 // pop exception and pop call frames.
	OpPanic2              Op = 0x17 // pop call frames.

	/* Unary & binary operators */
	OpUpos  Op = 0x20 // + (unary)
	OpUneg  Op = 0x21 // - (unary)
	OpUnot  Op = 0x22 // ! (unary)
	OpUxor  Op = 0x23 // ^ (unary)
	OpLor   Op = 0x26 // ||
	OpLand  Op = 0x27 // &&
	OpEql   Op = 0x28 // ==
	OpNeq   Op = 0x29 // !=
	OpLss   Op = 0x2A // <
	OpLeq   Op = 0x2B // <=
	OpGtr   Op = 0x2C // >
	OpGeq   Op = 0x2D // >=
	OpAdd   Op = 0x2E // +
	OpSub   Op = 0x2F // -
	OpBor   Op = 0x30 // |
	OpXor   Op = 0x31 // ^
	OpMul   Op = 0x32 // *
	OpQuo   Op = 0x33 // /
	OpRem   Op = 0x34 // %
	OpShl   Op = 0x35 // <<
	OpShr   Op = 0x36 // >>
	OpBand  Op = 0x37 // &
	OpBandn Op = 0x38 // &^

	/* Other expression operators */
	OpEval         Op = 0x40 // eval next expression
	OpBinary1      Op = 0x41 // X op ?
	OpIndex1       Op = 0x42 // X[Y]
	OpIndex2       Op = 0x43 // (_, ok :=) X[Y]
	OpSelector     Op = 0x44 // X.Y
	OpSlice        Op = 0x45 // X[Low:High:Max]
	OpStar         Op = 0x46 // *X (deref or pointer-to)
	OpRef          Op = 0x47 // &X
	OpTypeAssert1  Op = 0x48 // X.(Type)
	OpTypeAssert2  Op = 0x49 // (_, ok :=) X.(Type)
	OpStaticTypeOf Op = 0x4A // static type of X
	OpCompositeLit Op = 0x4B // X{???}
	OpArrayLit     Op = 0x4C // [Len]{...}
	OpSliceLit     Op = 0x4D // []{value,...}
	OpSliceLit2    Op = 0x4E // []{key:value,...}
	OpMapLit       Op = 0x4F // X{...}
	OpStructLit    Op = 0x50 // X{...}
	OpFuncLit      Op = 0x51 // func(T){Body}
	OpConvert      Op = 0x52 // Y(X)

	/* Type operators */
	OpFieldType       Op = 0x70 // Name: X `tag`
	OpArrayType       Op = 0x71 // [X]Y{}
	OpSliceType       Op = 0x72 // []X{}
	OpPointerType     Op = 0x73 // *X
	OpInterfaceType   Op = 0x74 // interface{...}
	OpFuncType        Op = 0x76 // func(params...)results...
	OpMapType         Op = 0x77 // map[X]Y
	OpStructType      Op = 0x78 // struct{...}

	/* Statement operators */
	OpAssign      Op = 0x80 // Lhs = Rhs
	OpAddAssign   Op = 0x81 // Lhs += Rhs
	OpSubAssign   Op = 0x82 // Lhs -= Rhs
	OpMulAssign   Op = 0x83 // Lhs *= Rhs
	OpQuoAssign   Op = 0x84 // Lhs /= Rhs
	OpRemAssign   Op = 0x85 // Lhs %= Rhs
	OpBandAssign  Op = 0x86 // Lhs &= Rhs
	OpBandnAssign Op = 0x87 // Lhs &^= Rhs
	OpBorAssign   Op = 0x88 // Lhs |= Rhs
	OpXorAssign   Op = 0x89 // Lhs ^= Rhs
	OpShlAssign   Op = 0x8A // Lhs <<= Rhs
	OpShrAssign   Op = 0x8B // Lhs >>= Rhs
	OpDefine      Op = 0x8C // X... := Y...
	OpInc         Op = 0x8D // X++
	OpDec         Op = 0x8E // X--

	/* Decl operators */
	OpValueDecl Op = 0x90 // var/const ...
	OpTypeDecl  Op = 0x91 // type ...

	/* Loop (sticky) operators (>= 0xD0) */
	OpSticky            Op = 0xD0 // not a real op.
	OpBody              Op = 0xD1 // if/block/switch/select.
	OpForLoop           Op = 0xD2
	OpRangeIter         Op = 0xD3
	OpRangeIterString   Op = 0xD4
	OpRangeIterMap      Op = 0xD5
	OpRangeIterArrayPtr Op = 0xD6
	OpReturnCallDefers  Op = 0xD7
)
```

> _Note: some op-codes are removed for simplicity if they refer to operations that are not implemented or meant to be removed._

The comments roughly guide what each op-code does, but we'll take a closer look at most of these opcodes in the upcoming sections.

### Blocks and Frames

- frames used for loops and switch statements as well

### Executing statements

In `op_exec.go`, we see the code that is involved in processing statements and pushing other, targeted op-codes to correctly execute a single statement.

The most simple case is that of `OpExec` itself, which simply pops the last statement off of the `Stmts` stack.

- [`AssignStmt`](https://go.dev/ref/spec#Assignment_statements) is converted into one of the `Op*Assign` (ie. OpAssign, or OpAddAssign for `+=`, and so on) or `OpDefine` op-codes. The right-hand side expressions are evaluated with `OpEval`, and so are the left-hand side if necessary (ie., something other than a simple `*NameExpr`).
- [`ExprStmt`](https://go.dev/ref/spec#Expression_statements) simply evaluates (`OpEval`) the expression within the statement, and discards any returned values.
- [`ForStmt`](https://go.dev/ref/spec#For_statements) creates a new frame and block

#### References

- Most statement execution code is in [`op_exec.go`](https://github.com/gnolang/gno/blob/master/gnovm/pkg/gnolang/op_exec.go).

> _Note: `op_exec.go` handles a lot of the so-called "sticky" op-codes. These are not popped with `PopOp`, but they are popped with `ForcePopOp`._

### Evaluating expressions

### Calling functions

### Constructing values from literals

### Assigning values

### Panicking

### Initializing packages

- RunFile

### Tracking gas and allocations

## Store

- MemPackage
- RefNode / RefValue
- SaveBlockNodes

## Objects and Ownership

## Go Type Checking

## Troubleshooting

## Glossary

- TypedValue
- Realm
