// PKGPATH: gno.land/r/test
package test

type (
	word uint
	nat  []word
)

var zero *Int

// structLit
type Int struct {
	abs nat
}

func main() {
	zero = &Int{
		abs: []word{0},
	}
	a := zero.abs
	println(a)
}

// Output:
// (slice[(0 gno.land/r/test.word)] gno.land/r/test.nat)

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.word"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.nat"
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "e256933ba4dfda233a7edb0902880d554118ba6e",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Int"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "2b8a024c53e94431e6203115feaf86b36413d7b2",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
