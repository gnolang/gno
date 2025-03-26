//go:build zealous

package gnodebug

// Zealous is true when using the build tag "zealous". It compiles additional
// error checks in the GnoVM which are mostly redundant and slow down execution,
// but that can be enabled to zealously verify execution.
const Zealous = true
