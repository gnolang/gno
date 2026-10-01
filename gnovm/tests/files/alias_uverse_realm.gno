// PKGPATH: gno.land/r/aliasuverse
package aliasuverse

// type alias to a uverse type — persisted via SetType/block commit.
type MyErr = error

func main(cur realm) {
	println("ok")
}

// Output:
// ok

// Types:
// MyErr[.uverse.error]=
//     {
//         "ID": ".uverse.error"
//     }
