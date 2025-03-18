// PKGPATH: gno.land/r/test
package test

type S struct {
	A *int
}

var a, b *S

func init() {
	a = new(S) // heapItemValue new real
	a.A = new(int)
	*a.A = 4
}

func main() {
	b = a // dirty, refCount change.
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.S"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f24d8c3b3b924a754166274405548b15452b047b",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
