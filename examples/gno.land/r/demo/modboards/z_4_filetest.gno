// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/modboards"
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
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/modboards:test_board/2) \[[reply](/r/demo/modboards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)] \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=2&postID=2)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/modboards:test_board/2/3) \[[reply](/r/demo/modboards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=2&postID=3)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// > \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/modboards:test_board/2/4) \[[reply](/r/demo/modboards?help&__func=CreateReply&bid=1&threadid=2&postid=4&body.type=textarea)] \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=2&postID=4)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=2&postid=4)]

// Realm:
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/modboards"]
// u[a973285dd27436279df76e639f8819ee0b31a1cd:106]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000003"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:107"
//                     }
//                 }
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:106",
//         "ModTime": "114",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:114",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:115]={
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
//                     "ID": "gno.land/r/demo/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:116"
//                     }
//                 }
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:115",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:114",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:114]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "6ac03215aa32d3183fa9b9653f60bd8aae0a1819",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:106"
//                     }
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
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "2c8e04fc2cf067f5f9f49b6a0d8d93090d372c8e",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:115"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:114",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:101",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:117]={
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:117",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:118]={
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:118",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:119]={
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:119",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:120]={
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
//                 "Index": "188",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:120",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:121]={
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:121",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:116]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/modboards.Board"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/modboards.Board"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:83"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/modboards.PostID"
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
//                 "Hash": "17799e17b6fed4e28bc10ae6fd624823c3dd3585",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:117"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "2dc8eaf2e1e79d135c69b7d4ae125cda10cb5890",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:118"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "3ca3af32f921eb0e80848b87addc5bcb33ec722a",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:119"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/modboards.PostID"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/modboards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/modboards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "cf0d8b3ab0d01918f851ad9062f513b57bc845c8",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:120"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "25d4f2e4a9e13ca66a8a27545922223701f27349",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:121"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:116",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[a973285dd27436279df76e639f8819ee0b31a1cd:113]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0000000003"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:107"
//                     }
//                 }
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:113",
//         "ModTime": "122",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:122",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:123]={
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
//                     "ID": "gno.land/r/demo/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:116"
//                     }
//                 }
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
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:123",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:122",
//         "RefCount": "1"
//     }
// }
// c[a973285dd27436279df76e639f8819ee0b31a1cd:122]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "312eaafbf4da5ea10e0f873ebae1e377b8349bd6",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:113"
//                     }
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
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "3647a9eb6f3f01c8c27d6bf23c18a3156a6f7e6f",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:123"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:122",
//         "ModTime": "0",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:102",
//         "RefCount": "1"
//     }
// }
// u[a973285dd27436279df76e639f8819ee0b31a1cd:83]={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/modboards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "/r/demo/modboards:test_board"
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
//                 "value": "g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "60aaa77b338f464d9c94dc7c1388d2e65f7b7ef6",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:84"
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
//                 "Hash": "b8ce64eda662aea25b2d8af05067dbc13340832e",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:85"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "f211b456f1e43a05b838d296d9fd37afc2c72448",
//                 "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:86"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/teritori/flags_index.FlagsIndex"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/teritori/flags_index.FlagsIndex"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "f8c2e4fe362ead246e9022c33aef87c4b08b5049",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:87"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:83",
//         "IsEscaped": true,
//         "ModTime": "113",
//         "RefCount": "6"
//     }
// }
// u[a973285dd27436279df76e639f8819ee0b31a1cd:101]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "d8a1c80764eff592ccac21d3a3895bf16c666b83",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:114"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:101",
//         "ModTime": "113",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:100",
//         "RefCount": "1"
//     }
// }
// u[a973285dd27436279df76e639f8819ee0b31a1cd:102]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "a3ea780ad62e5348d0865004ae781a711867bc8f",
//                         "ObjectID": "a973285dd27436279df76e639f8819ee0b31a1cd:122"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a973285dd27436279df76e639f8819ee0b31a1cd:102",
//         "ModTime": "113",
//         "OwnerID": "a973285dd27436279df76e639f8819ee0b31a1cd:100",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/modboards"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/modboards"]
// switchrealm["gno.land/r/boards_test"]
