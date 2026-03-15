NOTE: These issues describe the *benchops contract-level* system (cmd/benchops).
Most are now superseded by 350 Go-level microbenchmarks in bench_ops_test.go
with full parameterization (BigInt bit-lengths, string lengths, interface method
counts, array sizes, etc.). The benchops system remains useful for end-to-end
measurement but is no longer the primary calibration source.

---

⏺ 1. Single test case per op — Each op is exercised with one specific input shape, repeated 1000 times. Ops
   with parameter-dependent cost get a single data point that may not represent typical usage.
  2. Ops where cost varies significantly with parameters but only one case is tested:                      
                                                            
  | Op               | What's tested                 | What varies                          |              
  |------------------|-------------------------------|--------------------------------------|
  | OpSliceLit       | 10 elements                   | # of elements                        |
  | OpSliceLit2      | {9: 90} (sparse)              | allocated size                       |
  | OpArrayLit       | 2 elements                    | # of elements                        |
  | OpStructLit      | 1-field struct                | # of fields                          |
  | OpMapLit         | empty make(map)               | # of entries                         |
  | OpCall           | 0-1 args                      | # of args                            |
  | OpPrecall        | simple cases                  | # of args, variadic                  |
  | OpDefine         | 1 target                      | # of LHS targets                     |
  | OpAssign         | 1 target                      | # of LHS targets                     |
  | OpReturn         | 0-1 values                    | # of return values                   |
  | OpEql/OpNeq      | int only                      | type (string comparison much slower) |
  | OpIndex1         | array, map, string all lumped | container type                       |
  | OpSwitchClause   | 3 cases                       | # of cases                           |
  | OpCallNativeBody | println, len only             | which native fn                      |
  | OpEval           | all expression types averaged | const vs var lookup vs fn ref        |

  3. Constant-folded dead code in OpBinary — c = true || false and c = true && false are folded to
  constants at parse time; they never generate OpLor/OpLand. Not missing coverage (separate functions
  exist), but misleading.
  4. Stats use weighted mean — sum(totalTime) / sum(count) is vulnerable to GC pause outliers. The median
  fix in stats.go would address this (we wrote it but reverted).
  5. Gas charges are flat constants for ops that should scale — Currently m.incrCPU(OpCPUStructLit) charges
   the same gas whether the struct has 1 field or 10 fields. For accurate gas, parameterized ops need
  formulas like OpCPUStructLitBase + numFields * OpCPUStructLitPerField.
