// Mixed-origin defer-repanic chain.
// Head: user panic("A") — raised from the gno panic builtin path
// Defer: nil-pointer dereference — raised from a different VM helper
// (values.go GetPointerToFromTV via NewException) during cleanup.
//
// This is the narrow case where per-link GoStack would diverge between
// the head and the re-panic — VM-internal failure during defer cleanup.
// For everyday gno-program chains (panic in both head and defer body),
// per-link GoStack is largely redundant since both raises traverse the
// same VM helper.

package main

func main() {
	defer func() {
		var p *int
		_ = *p
	}()
	panic("A")
}

// Stacktrace:
// panic: A
// main<VPBlock(1,0)>()
//     main/panic_defer_vm_repanic.gno:19
// panic: runtime error: nil pointer dereference
// defer func(){ ... }()
//     main/panic_defer_vm_repanic.gno:17
// main<VPBlock(1,0)>()
//     main/panic_defer_vm_repanic.gno:15

// Error:
// A
// 	runtime error: nil pointer dereference
