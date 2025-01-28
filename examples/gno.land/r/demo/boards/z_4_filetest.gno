// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	rid := boards.CreateReply(bid, pid, pid, "Reply of the second post")
	println(rid)
}

func main() {
	rid2 := boards.CreateReply(bid, pid, pid, "Second reply of the second post")
	println(rid2)
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// 3
// 4
// # Second Post (title)
//
// Body of the second post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=2)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=3)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/4) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=4)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=4)]
//

// Realm:
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/boards"]
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111",
//         "ModTime": "123",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "b3fdb7591281b0194b8069df5033d894204c3d71",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:112"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:125]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000004"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:126"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:125",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:124",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:124]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:124",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "ce0c16bc81b4ac65550e5d95535fe8a94ac248ce",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:125"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000004"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "6cb294b1c9ab711292dc78dee5804e100480e792",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "b16feac3f9201dc270be001e57759244028c2b17",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:124"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3dc296c5d72371b6954a3e4da86c75a8bfe9aedf",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:128]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:128",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:129]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:129",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:130]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:130",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:131]={
//     "Fields": [
//         {
//             "N": "AAAAgJSeXbo=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "65536"
//             }
//         },
//         {
//             "N": "AbSNdvQQIhE=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "1024"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "time.Location"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "336074805fc853987abe6f7fe3ad97a6a6f3077a:2"
//                 },
//                 "Index": "182",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:131",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:132]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "65536"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "1024"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "time.Location"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:132",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/boards.Board"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/boards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": ""
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "Second reply of the second post"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "4c8f2c5036d29a9ce62443f2621ae50c3735b075",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:128"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "72a182a8e92df147b113866866b696005eb46371",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:129"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "bfa365b9300ca460702e2c70e6719740170c6a44",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:130"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/boards.PostID"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/boards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "0fd3352422af0a56a77ef2c9e88f479054e3d51f",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:131"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "bed4afa8ffdbbf775451c947fc68b27a345ce32a",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:132"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:126",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:126]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:126",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/boards.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c7abe064e504ac086135f800abe410f299ade6f0",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127"
//         }
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120",
//         "ModTime": "134",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "aff2525ff47043fd4d353275c8289abab1e74ae4",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:121"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:136]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000004"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:126"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:136",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:135",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:135]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:135",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f0912e9804582874b334d52b3de293daa48991d7",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:136"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000004"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "e9e2cf3f9b945cb2050108e4cc3d74816f13faa4",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "97157c6cff2bc184b4c4de069a67cd61527daed7",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:135"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "95dfc9bc38a3921c344dee669d9fc34bd0dc8326",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134"
//         }
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85]={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "/r/demo/boards:test_board"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "test_board"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "6d9281b14d102f8f9449829f8cc6247580c3f0d0",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:86"
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "65536"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "36299fccbc13f2a84c4629fad4cb940f0bd4b1c6",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:87"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "34f30363d60546811583ae7bbabf61c3b22ea181",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:88"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85",
//         "ModTime": "121",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84",
//         "RefCount": "1"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "904b687823acd23a231df9a9e7ca4f6713eb4a33",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106",
//         "ModTime": "121",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:105",
//         "RefCount": "1"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "6f1eec6dbd869b35f8f320f38c00c3ed069d6df1",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107",
//         "ModTime": "121",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:105",
//         "RefCount": "1"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84]={
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84",
//         "IsEscaped": true,
//         "ModTime": "127",
//         "RefCount": "6"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/boards.Board"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "ffcff26b8d64bcc4a51d749f6db0585a565721d8",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/demo/boards_test"]
