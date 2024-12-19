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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "826b2c6a2dc588d5612559e3c2db148e53677273",
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
//                 "Origin": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:7:[7]",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "617aa28ef20731f64f987cc74328d93068404195",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "eca577404bf9b24c1f42716d1ca53a7241934d2e",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111"
//                 },
//                 "Index": "0",
//                 "Origin": "26a90809de387e567a70972c32a4cc9e8d44e55f:8:[4]",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "92be030245d9441c93a2a7b0994bceca35a6762c",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:124"
//                 },
//                 "Index": "0",
//                 "Origin": "",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "e7ce72d977effa764ebce5ce69c81180769800f5",
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
//                     "ID": "gno.land/p/demo/avl.Node"
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
//                     "ID": "gno.land/p/demo/avl.Node"
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
//                     "ID": "gno.land/p/demo/avl.Node"
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
//                 "Origin": "336074805fc853987abe6f7fe3ad97a6a6f3077a:1:[182]",
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
//                 "Origin": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:7:[5]",
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
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "f91e355bd19240f0f3350a7fa0e6a82b72225916",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:128"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "9ee9c4117be283fc51ffcc5ecd65b75ecef5a9dd",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:129"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "eb768b0140a5fe95f9c58747f0960d647dacfd42",
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
//                 "Hash": "430705bbd54c3cef064353a53d10621ca719f759",
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
//             "Hash": "293d21c8d5610f948f32d4871d403a6aa9aba719",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "266d7868f2c3fedac6f099c42f12c9af87d4cbdf",
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
//                 "Origin": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:7:[7]",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "517ed502712f1f0c1261888c3b649b134bd5efa9",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "e8f24b3ccb13ce38a5921adfa524f7e0c7ab12c3",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120"
//                 },
//                 "Index": "0",
//                 "Origin": "26a90809de387e567a70972c32a4cc9e8d44e55f:8:[4]",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "80a6eb23228a7ec0c774621f32265f2452fdde31",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:135"
//                 },
//                 "Index": "0",
//                 "Origin": "",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "57209d9c283b62a9f40acdc4440d2f31b90ca3c0",
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
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c66cd814f4660f30d8af0a3e845fc43f1c676c0f",
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
//                 "Hash": "46b114b6fb76bd3d7999f75f6fd22c12c03ad53b",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:87"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "af6ed0268f99b7f369329094eb6dfaea7812708b",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "027e69e43b91a8d0d386340ad942ae18e68f1af1",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122"
//                 },
//                 "Index": "0",
//                 "Origin": "26a90809de387e567a70972c32a4cc9e8d44e55f:8:[4]",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3fca080af447330ba61226bfe5d31c83f08ffcf0",
//                     "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133"
//                 },
//                 "Index": "0",
//                 "Origin": "26a90809de387e567a70972c32a4cc9e8d44e55f:8:[4]",
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
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/demo/boards_test"]
