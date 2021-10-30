// PKGPATH: gno.land/r/test
package test

var root interface{}

func main() {
	println(root)
	root = 1
	println(root)
}

// Output:
// nil
// 1

// The below tests that the realm's block (of 1 variable) changed.  The first
// element image in the package (block) is for the "main" function, which
// appears first because function declarations are defined in a file before
// vars.

// Realm:
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.rfn",
//         "BlockNode": null,
//         "Location": {
//             "File": "",
//             "Line": "0",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//             },
//             "V": {
//                 "@type": "/gno.fun",
//                 "Closure": {
//                     "@type": "/gno.rfv",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "files/zrealm0.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm0.go",
//                         "Line": "6",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.rft",
//                     "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//                 }
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "6da88c34ba124c41f977db66a4fc5c1a951708d2"
//             }
//         }
//     ]
// }
//
