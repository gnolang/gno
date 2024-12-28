// PKGPATH: gno.land/r/test
package test

var (
	a = &A{A: "here"}
	b [2]*B
)


type A struct {
	A string
}
type B struct {
	A *A
	B string
}

func init() {
 	c := B{
		A: a,
		B: "c",
	}
	b[0] = &c

	d := B{
		A: a,
		B: "d",
	}
	b[1] = &d
}

func main() {
	b[0] = nil
	b[1] = nil
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.B"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.B"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.A"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "c"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "Hash": "c22ccb7832b422c83fec9943b751cb134fcbed0b",
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "0"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.A"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "d"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "Hash": "86c916fd78da57d354cb38019923bf64c1a3471c",
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "0"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "IsEscaped": true,
//         "ModTime": "9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.A"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "95dca2f5b12899b6367402ecdac04c7ca59a03d9",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//         }
//     }
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]
