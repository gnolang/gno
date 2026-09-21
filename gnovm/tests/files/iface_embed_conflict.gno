// Embedding two interfaces with the same method name but conflicting
// signatures is rejected, matching Go's "duplicate method" compile error.
// go/types reports it (TypeCheckError); the VM's flattenInterfaceMethods
// reports it as a positioned preprocess Error when go/types is not run.
package main

type A interface{ M() int }
type B interface{ M() string }

func main() {
	var x interface {
		A
		B
	}
	_ = x
	println("unreached")
}

// Error:
// main/iface_embed_conflict.gno:11:8-14:3: duplicate method M with conflicting types in interface

// TypeCheckError:
// main/iface_embed_conflict.gno:13:3: duplicate method M; main/iface_embed_conflict.gno:12:3: 	other declaration of method M
