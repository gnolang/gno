# GnoVM Op Handler Gas Audit

Every `doOpXxx` handler, its cost-varying parameters, pessimistic inputs, and rationale.
Produced through 3 successive audit passes with per-op verification against source code.

## op_binary.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpBinary1 | LAND/LOR short-circuit; right operand complexity | LAND with true LHS (must eval RHS) | Short-circuit fails, must evaluate right operand |
| doOpLor | None | Any booleans | Constant O(1) |
| doOpLand | None | Any booleans | Constant O(1) |
| doOpEql | Operand type; array element count; struct field count; BigInt/BigDec bit-length; string length | Array/struct with many elements/fields (recursive O(n) comparison) | isEql recursively compares elements/fields; BigInt.Cmp is O(bit-length) |
| doOpNeq | Same as doOpEql | Same as doOpEql | Calls isEql then negates |
| doOpLss | Operand type; string length; BigInt/BigDec bit-length | Long strings or large BigInt/BigDec | String comparison O(min(len)), BigInt.Cmp O(bit-length) |
| doOpLeq | Same as doOpLss | Same as doOpLss | Same comparison logic |
| doOpGtr | Same as doOpLss | Same as doOpLss | Same comparison logic |
| doOpGeq | Same as doOpLss | Same as doOpLss | Same comparison logic |
| doOpAdd (addAssign) | Operand type; string length; BigInt/BigDec bit-length | StringType: long strings (allocates O(len1+len2)); BigDec: high precision | String concat allocates; BigDec apd.Add O(precision) |
| doOpSub (subAssign) | Operand type; BigInt/BigDec bit-length | UntypedBigintType or BigDec with large bit-length | big.Int.Sub O(n), apd.Sub O(precision) |
| doOpMul (mulAssign) | Operand type; BigInt/BigDec bit-length | BigInt: large operands O(n^2 Karatsuba); BigDec: O(prec^2) | Multiplication is superlinear in operand size |
| doOpQuo (quoAssign) | Operand type; BigInt/BigDec bit-length; zero-check | BigInt/BigDec large operands | Division O(n^2); includes zero-check |
| doOpRem (remAssign) | Operand type; BigInt bit-length; zero-check | BigInt large operands | big.Int.Rem O(n^2) |
| doOpShl (shlAssign) | Operand type; shift amount; BigInt bit-length; Stage (overflow checks) | BigInt with shift near maxBigintShift=10000; StagePre | big.Int.Lsh O(n); overflow check allocates+Cmp |
| doOpShr (shrAssign) | Operand type; shift amount; BigInt bit-length; Stage | BigInt with large shift; StagePre | big.Int.Rsh O(n); overflow check allocates. **No maxBigintShift limit** unlike Shl (potential DoS vector) |
| doOpBand (bandAssign) | Operand type; BigInt bit-length | BigInt max bit-length | big.Int.And O(max(n1,n2) words) |
| doOpBor (borAssign) | Operand type; BigInt bit-length | BigInt max bit-length | big.Int.Or O(max(n1,n2) words) |
| doOpXor (xorAssign) | Operand type; BigInt bit-length | BigInt max bit-length | big.Int.Xor O(max(n1,n2) words) |
| doOpBandn (bandnAssign) | Operand type; BigInt bit-length | BigInt max bit-length | big.Int.AndNot O(max(n1,n2) words) |

## op_unary.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpUpos | None | Any value | No-op (identity) |
| doOpUneg | Operand type; BigInt/BigDec bit-length/precision | BigDec max precision | BigDec.Neg allocates; BigInt.Neg allocates; all scale with magnitude |
| doOpUnot | None | Any boolean | Constant O(1) bit flip |
| doOpUxor | Operand type; BigInt bit-length | BigInt max bit-length | big.Int.Not allocates O(n words) |
| doOpUrecv | N/A | N/A | Not implemented (panics) |

## op_expressions.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpIndex1 | Container type (map vs array/slice); map key type complexity; ComputeMapKey allocation | Map with large composite key (array/struct) | ComputeMapKey allocates `make([]byte, 0, 64)` per call; composite key serialization scales with key structure |
| doOpIndex2 | Same as doOpIndex1 | Same as doOpIndex1 | Map comma-ok: same lookup + bool return |
| doOpSelector | VPField/VPSubrefField (O(1)); VPValMethod/VPPtrMethod (BoundMethodValue alloc); VPInterface (findEmbeddedFieldType recursive O(depth×fields)); VPBlock (O(Depth) parent traversal) | Large struct with deeply nested embedded fields via VPInterface path | findEmbeddedFieldType traverses recursively |
| doOpSlice | Number of indices (2-index vs 3-index); bounds checking | 3-index slice (GetSlice2: 8 bounds checks vs GetSlice: 4) | More bounds checks; allocation of new slice header |
| doOpStar | Pointer target type; DataByteType special case | Pointer to DataByteType | DataByteType path allocates new TypedValue |
| doOpRef | None significant | Any | O(1): PopAsPointer2 + AllocatePointer(32B) + NewType(&PointerType, 200B) |
| doOpTypeAssert1 | Concrete vs interface assertion; interface method count; embedded field depth | Interface with many methods | VerifyImplementedBy iterates all methods O(M); each calls findEmbeddedFieldType O(F×D) |
| doOpTypeAssert2 | Same as TypeAssert1; defaultTypedValue on miss | Interface with many methods, assertion fails | IsImplementedBy walks all methods; defaultTypedValue allocates on miss |
| doOpCompositeLit | Element count; composite type (array/slice/map/struct); keyed vs unkeyed | Map literal with many keyed elements | Dispatches to sub-ops; map pushes 2x elements (key+value) |
| doOpArrayLit | Array size; keyed vs unkeyed; element type (uint8 vs non-uint8) | Large non-uint8 array with keyed elements | uint8: NewDataArray (flat bytes). Non-uint8: NewListArray + Copy per element + defaultTypedValue fill. Keyed: set[] O(bt.Len) |
| doOpSliceLit | Element count; element type | Slice with many complex elements | NewListArray O(n); PopCopyValues copies each element |
| doOpSliceLit2 | Max index value; element count; sparsity | Large max index with few elements (sparse) | Allocates array sized to maxVal+1 even if sparse; gap-filling loop calls defaultTypedValue for every unfilled slot |
| doOpMapLit | Number of key-value pairs; key complexity | Map with many pairs and large composite keys | ComputeMapKey per element O(pairs × key_complexity). vmap initialized with capacity 0 (rehash cost) |
| doOpStructLit | Field count; keyed vs unkeyed | Large struct unkeyed (el == nf copies) | Keyed: uses pre-computed fnx.Path.Index for O(1) access. Unkeyed: Copy per field. Both: defaultStructFields O(nf) |
| doOpFuncLit | Heap capture count; capture block depth | Closure with many captures at deep block depth | Each capture: GetPointerToDirect O(path.Depth) parent block traversal. Total: O(numCaptures × avgDepth) |
| doOpConvert | Source/target type; string length; rune count | String→[]rune with long string (O(rune_count) allocs) | String→[]byte O(byte_length). String→[]rune O(rune_count) with per-rune TypedValue alloc. []rune→String O(rune_count) UTF-8 re-encode |

## op_assign.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpDefine | Number of LHS variables; heap escape detection | Max variables with NameExprTypeHeapDefine types | Loop O(n vars); heap escape creates HeapItemValue + PointerValue per var |
| doOpAssign | Number of LHS variables; PopAsPointer expression complexity | Max variables; nested selectors/index requiring complex pointer resolution | Loop O(n vars) reverse; PopAsPointer recurses through expression tree |
| doOpAddAssign | Operand type; string length; realm tracking | String concat or BigDec; realm object | addAssign allocates; DidUpdate with nil,nil,nil → just MarkDirty(po) |
| doOpSubAssign | Operand type; BigInt/BigDec size; realm tracking | BigDec with precision; realm object | apd.Sub; DidUpdate → MarkDirty |
| doOpMulAssign | Operand type; BigInt/BigDec size; realm tracking | BigDec (WithPrecision(1024).Mul); realm object | Higher precision setup; DidUpdate → MarkDirty |
| doOpQuoAssign | Operand type; zero-check; realm tracking | BigDec (WithPrecision(1024).Quo); realm object | Division + zero-check + precision; DidUpdate → MarkDirty |
| doOpRemAssign | Operand type; zero-check; realm tracking | BigInt; realm object | big.Int.Rem O(n^2); DidUpdate → MarkDirty |
| doOpBandAssign | Operand type; BigInt size; realm tracking | BigInt; realm object | big.Int.And allocates; DidUpdate → MarkDirty |
| doOpBorAssign | Operand type; BigInt size; realm tracking | BigInt; realm object | big.Int.Or allocates; DidUpdate → MarkDirty |
| doOpXorAssign | Operand type; BigInt size; realm tracking | BigInt; realm object | big.Int.Xor allocates; DidUpdate → MarkDirty |
| doOpBandnAssign | Operand type; BigInt size; realm tracking | BigInt; realm object | big.Int.AndNot allocates; DidUpdate → MarkDirty |
| doOpShlAssign | Operand type; shift amount; Stage; realm tracking | BigInt shift near 10000; StagePre; realm | big.Int.Lsh + overflow check allocs; DidUpdate → MarkDirty |
| doOpShrAssign | Operand type; shift amount; Stage; realm tracking | BigInt shift; StagePre; realm | big.Int.Rsh + overflow check; DidUpdate → MarkDirty. **No shift limit** |

Note: All compound assigns pass nil,nil,nil to DidUpdate — the expensive reference tracking (IncRefCount, MarkNewEscaped) never executes. Actual cost is just MarkDirty(po) on the base object.

## op_inc_dec.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpInc | Operand type (Int/Float/BigInt/BigDec); BigInt/BigDec magnitude; PopAsPointer depth; realm tracking | BigDec with max precision; deep pointer chain; realm object | BigDec arithmetic scales with precision; PopAsPointer traversal; DidUpdate walks hierarchy |
| doOpDec | Same as doOpInc | Same as doOpInc | Same cost structure as Inc |

## op_exec.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpExec (OpBody) | Body length (len(bs.Body)) | Body with many statements | Multiple statement pushes |
| doOpExec (OpForLoop) | Body length; init variable count; condition eval | Many statements; many init vars with heap items | Each init var requires HeapItemValue copy per iteration (O(NumInit) per loop) |
| doOpExec (OpRangeIter) | Array/slice length; ASSIGN vs DEFINE; key+value assignment; body length | Large slice with both key and value assigns | Upfront array copy via xv.Copy(m.Alloc) + assignments per iter |
| doOpExec (OpRangeIterString) | String byte length; rune count; multi-byte chars; body length | Long UTF-8 string with multi-byte runes + key/value assigns | UTF-8 decoding per iteration via DecodeRuneInString |
| doOpExec (OpRangeIterMap) | Map size; key/value field count; body length | Large map with key+value assigns | Map linked-list traversal; fillValueTV per element |
| doOpExec (AssignStmt) | Number of RHS/LHS expressions; operator type | Many operands with complex operator | Multiple pushes + operator switch |
| doOpExec (ReturnStmt) | Number of results; defers present; copy requirement | Many return values with defers | Multiple result evals + defer handling |
| doOpExec (RangeStmt) | Range type (array/map/string); operator; body length | String type with DEFINE + body | Type checking + operator switch + block creation |
| doOpExec (BranchStmt) | Label matching; frame depth; branch type | Labeled break/continue with deep frame stack; goto requires block state restoration; fallthrough calls ExpandWith O(clause_vars) | Frame popping loop unbounded |
| doOpExec (DeclStmt) | Number of nested declarations | Many nested decls | Nested loop pushes |
| doOpExec (ValueDecl) | Number of values; Type expression presence | Many value initializers | Multiple value evaluations |
| doOpExec (DeferStmt) | Number of defer arguments | Many function arguments | Multiple argument evaluations |
| doOpExec (SwitchStmt) | Number of clauses; type switch vs regular; init | Many clauses + type switch | TypeSwitch separate op; block creation |
| doOpExec (BlockStmt) | Block body length | Many statements | Block alloc + statement init |
| doOpIfCond | Then/Else body length; new variable count | Large body with many block-scoped vars | ExpandWith allocates heap items for new vars O(var_count) |
| doOpTypeSwitch | Number of clauses; cases per clause; type comparison complexity; InterfaceKind calls IsImplementedBy | Many clauses with multiple type cases + interface types | Nested loops: O(clauses × cases); IsImplementedBy O(methods × findEmbeddedFieldType) |
| doOpSwitchClause | Number of clauses; body length | Match at last clause; large body | Iteration to find matching clause |
| doOpSwitchClauseCase | isEql comparison complexity; remaining cases/clauses | Complex typed value (array/struct); many remaining | isEql recursive per case tried |

## op_call.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpPrecall | Function type (FuncValue/BoundMethod/TypeValue); IsWithCross; IsCrossing; NumArgs | FuncValue with IsWithCross+IsCrossing requiring realm creation | NewConcreteRealm allocation + Assign |
| doOpEnterCrossing | Frame stack depth | Deep call stack with realm boundary check | **O(n²)**: PeekCallFrame(i) internally iterates backwards through ALL frames each time. Code has TODO: "O(n²), optimize." |
| doOpCall | Number of captures; block size (NumNames); return count; param count; native vs Gno; variadic expansion | Gno function with many captures + params + returns + heap-defined results | Captures copy O(n); NewBlock O(NumNames); defaultTypedValue per result O(results); popCopyArgs variadic O(nvar) slice alloc; Store lookups for GetSource/GetType/GetParent |
| doOpCallNativeBody | None (delegates to native) | Any native call | Cost depends on native implementation |
| doOpCallDeferNativeBody | None | Any deferred native | Pop + call |
| doOpReturn | Frame depth; realm boundary checks | Deep call stack with realm finalization | PopUntilLastCallFrame O(n frames); FinalizeRealmTransaction (expensive: processNewCreated/Deleted/Escaped + markDirtyAncestors + saveUnsaved) |
| doOpReturnAfterCopy | Number of return values; heap-defined results; realm | Many named results heap-defined + realm crossing | Per-result Copy + realm finalization |
| doOpReturnFromBlock | Number of return values; fillValueTV complexity | Many named returns with RefValues/HeapItemValues | O(results) fillValueTV may deref recursively; store lookups |
| doOpReturnToBlock | Number of return values | Many named returns | O(results) AssignToBlock + Copy |
| doOpReturnCallDefers | Number of defers; capture count; variadic args; native vs Gno | Many defers with many captures + variadic params | Per-defer: PushFrameCall + NewBlock O(NumNames) + captures copy O(n) + popCopyArgs O(nvar) |
| doOpDefer | Argument count; variadic expansion; BoundMethod receiver; captures | Variadic with many args + receiver + captures | MustPeekCallFrame O(frames); popCopyArgs O(nvar) slice alloc + element copy |
| doOpPanic2 | Frame depth; exception chain depth | Deep call stack with nested exceptions | PopUntilLastCallFrame O(n frames); if unhandled: makeUnhandledPanicError O(numExceptions × Sprint cost) |

## op_eval.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpEval | Expression type; NameExpr block depth; literal string length; composite type | Deep NameExpr (Depth>0) requiring block traversal; large string/bigint literal; hex float literal | NameExpr: LastBlock + GetPointerTo traverses O(depth); INT literal: big.SetString O(len); FLOAT hex: regex + apd parsing; STRING: alloc.NewString O(len) |

## op_decl.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpValueDecl | Number of NameExprs; Type presence; untyped conversion; interface kind; const flag | Many uninitialized non-interface vars without explicit type | Loop O(n vars); defaultTypedValue recursively allocates for struct/array types (O(outer × inner × ...)) |
| doOpTypeDecl | Type definition complexity | Complex nested type | PopValue.GetType traversal + Assign2 |

## op_types.go

| Op handler | Cost-varying parameters | Pessimistic input | Why pessimistic |
|---|---|---|---|
| doOpFieldType | Tag string length | Long tag string | String extraction |
| doOpArrayType | Length value magnitude | Large specified length | Value pop + convert |
| doOpSliceType | None | Any | Constant cost |
| doOpFuncType | Parameter count; result count | Many params + results | Loop pops O(params + results) |
| doOpMapType | None | Any | Pop key + value types |
| doOpStructType | Field count; embedded field detection | Many fields with embedded types | PopValues O(fields); fillEmbeddedName per field |
| doOpInterfaceType | Method count; embedded interfaces | Many methods with embedded interfaces | PopValues O(methods); fillEmbeddedName per method |
| doOpChanType | None | Any | Pop element type |
| doOpStaticTypeOf | Expression type; selector depth; block depth | SelectorExpr with deep nesting; IndexExpr requiring full eval | Recursive OpStaticTypeOf; Store lookups for packages |

---

## Corrections from multi-pass audit

| Item | Original claim | Correction |
|---|---|---|
| doOpRef | "Heap capture count; HeapCaptures iteration O(n)" | HeapCaptures iteration is in doOpFuncLit, NOT doOpRef. doOpRef is O(1) |
| doOpEnterCrossing | "O(n) frame loop" | Actually **O(n²)** because PeekCallFrame(i) scans backwards each time |
| doOpShr BigInt | "Shift behavior similar to Shl" | BigInt right shift has **no maxBigintShift limit** unlike left shift (capped at 10000) |
| doOpStructLit keyed | "Searches by field index O(el)" | Uses pre-computed fnx.Path.Index for O(1) direct array access |
| Comparison ops | "lessAssign helper" | No shared helper — 4 separate functions: isLss, isLeq, isGtr, isGeq |
| Compound assigns DidUpdate | "Realm ref-counting" | Passes nil,nil,nil → expensive paths never execute; just MarkDirty(po) |

## Benchmark gap analysis

### Missing benchmarks (not yet covered)

1. **isEql ArrayKind** — recursive O(N) comparison
2. **isEql StructKind** — recursive O(fields) comparison
3. **doOpConvert String→[]rune** — O(rune_count) allocs
4. **doOpConvert []rune→String** — O(rune_count) re-encode
5. **doOpSliceLit2 sparse** — maxVal amplification
6. **doOpSelector VPValMethod** — BoundMethodValue allocation
7. **doOpSelector VPInterface** — findEmbeddedFieldType recursion
8. **doOpSelector VPBlock** — block depth traversal
9. **doOpTypeAssert1 interface** — VerifyImplementedBy O(M×F×D)
10. **doOpFuncLit with captures** — O(n×depth)
11. **doOpCall** — captures, block alloc, variadic
12. **doOpReturn** — frame depth, realm finalization
13. **doOpDefer** — argument count, variadic
14. **doOpExec OpForLoop** — heap item copy overhead
15. **doOpExec OpRangeIter** — upfront array copy
16. **doOpExec OpRangeIterString** — UTF-8 decode per rune
17. **doOpExec OpRangeIterMap** — linked list traversal
18. **doOpIfCond** — ExpandWith heap var allocation
19. **doOpTypeSwitch** — clause × case iteration
20. **doOpSwitchClauseCase** — isEql per case
21. **doOpEval NameExpr** — block depth traversal
22. **doOpEval BasicLitExpr** — literal parsing (hex float most expensive)
23. **doOpValueDecl** — defaultTypedValue recursion
24. **doOpPrecall** — realm creation for crossing
25. **doOpPanic2** — exception chain Sprint
26. **doOpEnterCrossing** — O(n²) frame scan

### Existing benchmarks needing additional parameterizations

1. **BenchmarkOpShl_BigInt** — add shift values near maxBigintShift=10000
2. **BenchmarkOpShr_BigInt** — add very large shifts (no limit!)
3. **BenchmarkOpConvert_StringToBytes** — add String→[]rune variant
4. **BenchmarkOpSliceLit** — add sparse keyed variant (doOpSliceLit2)
5. **BenchmarkOpArrayLit** — add uint8 element type variant
6. **BenchmarkOpSelector** — add method selector and interface selector variants
7. **BenchmarkOpTypeAssert1** — add interface with many methods variant
8. **BenchmarkOpEql** — add array and struct equality variants
