# Op Gas Formula Shapes

Each opcode's CPU gas cost follows one of these patterns:
- **flat**: single constant, no parameters
- **base + slope * N**: linear in one dimension
- **base + slope * N1 * N2**: product of two operand sizes
- **base + slopeP * P + slopeC * C**: sum of two independent dimensions

Alloc gas is charged separately by the allocator and is not included here.

| Op | Formula | N = |
|---|---|---|
| **Binary arithmetic** | | |
| OpAdd/Sub/Mul/Quo/Rem (int) | flat | — |
| OpAdd/Sub/Mul/Quo (float64) | flat | — |
| OpAdd (string) | flat | — (alloc gas covers the O(N) memory cost) |
| OpAdd/Sub (BigInt) | base + slope * max(N1, N2) | N1, N2 = bit widths |
| OpMul (BigInt) | base + slope * N1 * N2 | N1, N2 = bit widths |
| OpQuo/Rem (BigInt) | base + slope * N1 * N2 | N1 = dividend bits, N2 = divisor bits |
| OpBand/Bandn (BigInt) | base + slope * min(N1, N2) | N1, N2 = bit widths |
| OpBor/Xor (BigInt) | base + slope * max(N1, N2) | N1, N2 = bit widths |
| OpShl/Shr (int) | flat | — |
| OpLand/Lor | flat | — |
| OpAdd/Sub (BigDec) | base + slope * max(N1, N2) | N1, N2 = digit counts |
| OpMul (BigDec) | base + slope * N1 * N2 | N1, N2 = digit counts |
| OpQuo (BigDec) | base + slope * N1 * N2 | N1 = dividend digits, N2 = divisor digits |
| **Comparison** | | |
| OpEql/Neq/Lss/Leq/Gtr/Geq (int) | flat | — |
| OpEql (float64) | flat | — |
| OpEql/Neq (BigInt) | base + slope * min(N1, N2) | N1, N2 = bit widths |
| OpLss (BigInt) | base + slope * min(N1, N2) | N1, N2 = bit widths |
| OpEql (string) | base + slope * min(N1, N2) | N1, N2 = string lengths |
| OpLss (string) | base + slope * min(N1, N2) | N1, N2 = string lengths |
| OpEql (array of int) | base + slope * N | N = element count |
| OpEql (struct of int) | base + slope * N | N = field count |
| OpEql (byte array) | base + slope * min(N1, N2) | N1, N2 = byte lengths |
| **Unary** | | |
| OpUpos/OpUnot/OpUxor (int) | flat | — |
| OpUneg (int) | flat | — |
| OpUneg/Uxor (BigInt) | base + slope * N | N = bit width |
| OpUneg (BigDec) | base + slope * N | N = digit count |
| OpUrecv | flat | — |
| **Inc/Dec** | | |
| OpInc/Dec (int, float64) | flat | — |
| OpInc/Dec (BigInt) | base + slope * N | N = bit width |
| OpInc/Dec (BigDec) | base + slope * N | N = digit count |
| **Assign ops** | | |
| OpXxxAssign (int) x11 | flat | — |
| OpDefine | base + slope * N | N = LHS count |
| OpAssign | base + slope * N | N = LHS count |
| **Composite literals** | | |
| OpMapLit | base + slope * N | N = entry count |
| OpArrayLit | base + slope * N | N = element count |
| OpSliceLit | base + slope * N | N = element count |
| OpSliceLit2 (sparse) | base + slope * N | N = allocated size |
| OpStructLit (unnamed) | base + slope * N | N = field count |
| OpStructLit (named) | base + slope * N | N = field count |
| **Expressions** | | |
| OpEval (const/type/BasicLitInt) | flat | — |
| OpEval (NameExpr) | base + slope * N | N = block depth |
| OpBinary1 | flat | — |
| OpIndex1 | flat | — |
| OpSlice (array/slice/byte/3idx) | flat | — |
| OpSlice (string) | flat | — (alloc gas covers the O(N) memory cost) |
| OpStar/OpRef | flat | — |
| OpCompositeLit | flat | — |
| OpSelector (own/VPBlock/method) | flat | — |
| OpSelector (VPInterface) | base + slope * N | N = method count |
| OpFuncLit | base + slope * N | N = capture count |
| **Conversions** | | |
| OpConvert (int to string, int to int64) | flat | — |
| OpConvert (str to runes) | base + slope * N | N = string length |
| OpConvert (runes to str) | base + slope * N | N = rune count |
| OpConvert (str to bytes, bytes to str) | flat | — (alloc gas covers the O(N) memory cost) |
| **Type assertions** | | |
| OpTypeAssert1/2 (concrete) | flat | — |
| OpTypeAssert1/2 (interface) | base + slope * N | N = method count |
| **Call** | | |
| OpPrecall | flat | — |
| OpCall | base + slopeP * P + slopeC * C | P = params, C = captures |
| OpCallNativeBody | flat | — |
| OpReturn/ReturnAfterCopy/FromBlock/ToBlock | flat | — |
| OpReturnCallDefers | base + slope * N | N = defer count |
| OpDefer/OpPanic2 | flat | — |
| **Control flow** | | |
| OpBody/OpIfCond | flat | — |
| OpForLoop (simple) | flat | — |
| OpForLoop (heap copy) | base + slope * N | N = heap var count |
| OpRangeIter (array) | base + slope * N | N = element count |
| OpRangeIterString | flat | — (called once per rune, not per string) |
| OpRangeIterMap | flat | — (called once per entry, not per map) |
| OpSwitchClause/Case | flat | — |
| OpTypeSwitch | base + slopeC * C + slopeM * M | C = clause count, M = total method count across interface clauses |
| **Type construction** | | |
| OpFieldType/ArrayType/SliceType/ChanType/MapType | flat | — |
| OpFuncType | base + slope * N | N = param + result count |
| OpStructType | base + slope * N | N = field count |
| OpInterfaceType | base + slope * N | N = method count |
| **Declarations** | | |
| OpValueDecl (int) | flat | — |
| OpValueDecl (struct/array) | base + slope * N | N = field/element count |
| OpTypeDecl | flat | — |
