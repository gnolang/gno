// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var m1 map[string]int

func main() {
	m1 = crossrealm.GetMap()
	println(m1)
	crossrealm.ModifyMap()
	println(m1)
}

// Output:
// map{("a" string):(1 int)}
// map{("a" string):(2 int)}

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/map"]
// switchrealm["gno.land/r/demo/tests/crossrealm/map"]
// switchrealm["gno.land/r/crossrealm_test"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "0",
//             "File": "",
//             "Line": "0",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.MapType",
//                 "Key": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 },
//                 "Value": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "5a06306c27263d4b9ae8f2f30e82f8ef014042cd:6"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "files/zrealm_crossrealm37_map.gno",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_crossrealm37_map.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
// u[5a06306c27263d4b9ae8f2f30e82f8ef014042cd:6]={
//     "List": {
//         "List": [
//             {
//                 "Key": {
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "V": {
//                         "@type": "/gno.StringValue",
//                         "value": "a"
//                     }
//                 },
//                 "Value": {
//                     "N": "AgAAAAAAAAA=",
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "32"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "5a06306c27263d4b9ae8f2f30e82f8ef014042cd:6",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "OwnerID": "5a06306c27263d4b9ae8f2f30e82f8ef014042cd:2",
//         "RefCount": "2"
//     }
// }
