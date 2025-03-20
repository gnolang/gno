// PKGPATH: gno.land/r/test
package test

type S struct {
	A int
}

var a, b *S

func main() {
	a = new(S)
	a.A = 4
	b = a
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Fields": [
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.S"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0adbf7812105088758da13308ac71206531f676c",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
