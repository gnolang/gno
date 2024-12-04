// PKGPATH: gno.land/r/test
package test

var (
	a  = &A{A: "here"}
	a2 = &A{A: "here"}
	b  [2]*B
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "11",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
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
//         "Hash": "eb2a077c834c2611affad67bdeddfb4d481d0947",
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "0"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]={
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
//         "Hash": "3e711a4eade7634fe616a9b66646034f1d28dafc",
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "0"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "IsEscaped": true,
//         "ModTime": "11",
//         "RefCount": "0"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "files/zrealm17.gno",
//             "Line": "18",
//             "PkgPath": "gno.land/r/test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "3316460f1bfec49fce2aaf28a28277b12b1f5f15",
//                 "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "8cc8f6bfb70f530e2024a48391f7a1a697d1e65f",
//                 "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//             }
//         }
//     ]
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "IsEscaped": true,
//         "ModTime": "11",
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "IsEscaped": true,
//         "ModTime": "11",
//         "RefCount": "3"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "files/zrealm17.gno",
//             "Line": "2",
//             "PkgPath": "gno.land/r/test"
//         }
//     },
//     "Values": []
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]
