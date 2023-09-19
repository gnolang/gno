// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/teritori/modboards"
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
// switchrealm["gno.land/r/demo/teritori/modboards"]
// u[6d199dd1e49fa6f6555549f00e31be781730334c:106]={
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
//                     "ID": "gno.land/r/demo/teritori/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/teritori/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:107"
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:106",
//         "ModTime": "114",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:114",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:115]={
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
//                     "ID": "gno.land/r/demo/teritori/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/teritori/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:116"
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:115",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:114",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:114]={
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
//                         "Hash": "facb1f2204c075ab3915876d44b35b03a32d2714",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:106"
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
//                         "Hash": "ee98c450b69065d27b390f6cf7e0b45e75c702ba",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:115"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:114",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:101",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:117]={
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:117",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:118]={
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:118",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:119]={
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:119",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:120]={
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:120",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:121]={
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:121",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:116]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/teritori/modboards.Board"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/teritori/modboards.Board"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:83"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/teritori/modboards.PostID"
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
//                 "Hash": "a5c9b214fecb4c0f133299241c7fed3f3bc0c926",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:117"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "26353cd6f289e0161a8f77b53386094f3215c5be",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:118"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "52efc3f77f47ce85e50ec61e0248b8e89a61c3f2",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:119"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/teritori/modboards.PostID"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/teritori/modboards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/teritori/modboards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "92122c82bf8190d7d8ca69af29cb384b3ae7755f",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:120"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "46aab86193f93e2bc77a4bfa3b2fd6a59178765a",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:121"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:116",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[6d199dd1e49fa6f6555549f00e31be781730334c:113]={
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
//                     "ID": "gno.land/r/demo/teritori/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/teritori/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:107"
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:113",
//         "ModTime": "122",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:122",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:123]={
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
//                     "ID": "gno.land/r/demo/teritori/modboards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/teritori/modboards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:116"
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
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:123",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:122",
//         "RefCount": "1"
//     }
// }
// c[6d199dd1e49fa6f6555549f00e31be781730334c:122]={
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
//                         "Hash": "a109ef9fce90108117648bde2b34a6a253aedee9",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:113"
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
//                         "Hash": "71e85b5bd3de7c4f4da88a046440723ebe50c3fc",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:123"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:122",
//         "ModTime": "0",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:102",
//         "RefCount": "1"
//     }
// }
// u[6d199dd1e49fa6f6555549f00e31be781730334c:83]={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/teritori/modboards.BoardID"
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
//                 "Hash": "0f7b610282044250bba9e7803f28f7d3118b142b",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:84"
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
//                 "Hash": "24ce9ec1637acbdf323d7ecce0de006240d666ed",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:85"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "76ffc4d1ba9d18f5296dbc2bf56449536616b5ab",
//                 "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:86"
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
//                         "Hash": "bdd72236fed9ba86e5adaa251cff2900dd10a81e",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:87"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:83",
//         "IsEscaped": true,
//         "ModTime": "113",
//         "RefCount": "6"
//     }
// }
// u[6d199dd1e49fa6f6555549f00e31be781730334c:101]={
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
//                         "Hash": "68a29aa377a66ebdf03d1e28ccdbcb919aff80d0",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:114"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:101",
//         "ModTime": "113",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:100",
//         "RefCount": "1"
//     }
// }
// u[6d199dd1e49fa6f6555549f00e31be781730334c:102]={
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
//                         "Hash": "08fb6317042c1e30a8f780120d826ff6b48eb542",
//                         "ObjectID": "6d199dd1e49fa6f6555549f00e31be781730334c:122"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6d199dd1e49fa6f6555549f00e31be781730334c:102",
//         "ModTime": "113",
//         "OwnerID": "6d199dd1e49fa6f6555549f00e31be781730334c:100",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/teritori/modboards"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/teritori/modboards"]
// switchrealm["gno.land/r/boards_test"]
