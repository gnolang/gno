// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var local_m map[string]int

var f = func(m map[string]int) {
	local_m = m
}

func main() {
	crossrealm.GetMap2(f)
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
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
//                     "N": "AQAAAAAAAAA=",
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "32"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
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
//                 "Hash": "e539b0c14adfd8f7a89207e13caa3817869fb597",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "m",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.MapType",
//                             "Key": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             },
//                             "Value": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
//                             }
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "",
//                 "IsMethod": false,
//                 "Name": "",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "9",
//                         "File": "files/zrealm_crossrealm37a.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "m",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.MapType",
//                                 "Key": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 },
//                                 "Value": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
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
//                 "FileName": "files/zrealm_crossrealm37a.gno",
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
//                         "File": "files/zrealm_crossrealm37a.gno",
//                         "Line": "12",
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
// switchrealm["gno.land/r/demo/tests/crossrealm/map"]
// switchrealm["gno.land/r/crossrealm_test"]
