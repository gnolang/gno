// PKGPATH: gno.land/r/test
package test

type MyStruct struct {
	Field int
}

var ms MyStruct

func init() {
	ms.Field = 1
}

func main() {
	println(ms)
	ms.Field--
	println(ms)
}

// Output:
// (struct{(1 int)} gno.land/r/test.MyStruct)
// (struct{(0 int)} gno.land/r/test.MyStruct)

// Realm:
// switchrealm["gno.land/r/test"]
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "ModTime": "4",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
