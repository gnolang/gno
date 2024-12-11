// PKGPATH: gno.land/r/ptr_map
package ptr_map

type MyStruct struct {
	Name string
	Age  int
}

var m = make(map[MyStruct]string)
var i1 = MyStruct{Name: "alice", Age: 2} // this is owned by package block

func init() {
	m[i1] = "first key"
	println(m[i1])
}

func main() {
	delete(m, i1)

	println(m[i1] == "")
}

// Output:
// first key
// true

// Realm:
// switchrealm["gno.land/r/ptr_map"]
// u[15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:3]={
//     "List": {
//         "List": []
//     },
//     "ObjectInfo": {
//         "ID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:3",
//         "ModTime": "5",
//         "OwnerID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:2",
//         "RefCount": "1"
//     }
// }
// u[15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:2]={
//     "Blank": {
//         "Path": ""
//     },
//     "ObjectInfo": {
//         "ID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:2",
//         "IsEscaped": true,
//         "ModTime": "5",
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
//             "PkgPath": "gno.land/r/ptr_map"
//         }
//     },
//     "Values": [
//         {
//             "Path": "",
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Name",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Age",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/ptr_map"
//                     },
//                     "Methods": [],
//                     "Name": "MyStruct",
//                     "PkgPath": "gno.land/r/ptr_map"
//                 }
//             }
//         },
//         {
//             "Path": "",
//             "T": {
//                 "@type": "/gno.MapType",
//                 "Key": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/ptr_map.MyStruct"
//                 },
//                 "Value": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "896c84e53ec971bcc4dfb30c374852a796fc0395",
//                 "ObjectID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:3"
//             }
//         },
//         {
//             "Path": "",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/ptr_map.MyStruct"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a0b24ad8a3e16a1c44fb342faf4e71657fcec26b",
//                 "ObjectID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:4"
//             }
//         },
//         {
//             "Path": "",
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:5"
//                 },
//                 "FileName": "main.gno",
//                 "IsMethod": false,
//                 "Name": "init.3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/ptr_map",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "main.gno",
//                         "Line": "12",
//                         "PkgPath": "gno.land/r/ptr_map"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "Path": "",
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "15bc39c5756bbda67fd0d56bc9e86b29d7b444fc:5"
//                 },
//                 "FileName": "main.gno",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/ptr_map",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "main.gno",
//                         "Line": "17",
//                         "PkgPath": "gno.land/r/ptr_map"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
