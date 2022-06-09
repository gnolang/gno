// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200gnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
)

var bid boards.BoardID
var pid boards.PostID

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
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/3) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// > \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/4) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=4&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=4)]

// Realm:
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/boards"]
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:95]={
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
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:96"
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:95",
//         "ModTime": "101",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:101",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:102]={
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
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103"
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:102",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:101",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:101]={
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "a2e47d2ee8b19cb4023504afb17d78de5d1caa91",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:95"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "369c0bdcc880fdcd738a3a993ee59318fe57ffbd",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:102"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:101",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:92",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:104]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:104",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:105]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:105",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:106]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:106",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:103]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/boards.Board"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/boards.Board"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:81"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/boards.PostID"
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
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "fb30b843cca60a8889ae62a386b456a5b5879b1b",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:104"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "8a14052db32b85af68804e8dd01d9dc18ead3154",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:105"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "685c580b00cc8caad7739a5e7679f323f95a8be8",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:106"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/boards.PostID"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/boards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Time"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Time"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:100]={
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
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:96"
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:100",
//         "ModTime": "107",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:107",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:108]={
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
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103"
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:108",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:107",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:107]={
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
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "dde6f19f29d65501177723f4588ed1c2a1d38667",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:100"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "c180ebdf0669d435dd52d36c042a958a82daa54a",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:108"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:107",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:93",
//         "RefCount": "1"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:81]={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "/r/boards:test_board"
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
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "a88e19f7fd69c993c13dd1ca4c0a19017f5e21f5",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:82"
//                     }
//                 }
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
//                 "ID": "std.Time"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:81",
//         "IsEscaped": true,
//         "ModTime": "100",
//         "RefCount": "6"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:92]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "43f582bc92dc9515ddb3580f1b90644a60e26c3a",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:101"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:92",
//         "ModTime": "100",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:91",
//         "RefCount": "1"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:93]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "fb2c9abd7253d78df43772a2b8828cf95fcef1c9",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:107"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:93",
//         "ModTime": "100",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:91",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/boards"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/boards"]
// switchrealm["gno.land/r/boards_test"]
