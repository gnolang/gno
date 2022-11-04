// PKGPATH: gno.land/r/boards_test
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
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// > \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/4) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=2&postid=4&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=2&postid=4)]

// Realm:
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/boards"]
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:100]={
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
//                     "ID": "gno.land/r/demo/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:101"
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:100",
//         "ModTime": "108",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:108",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:109]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110"
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:109",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:108",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:108]={
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "b0e1bb401a1e4392dc4aa8353e35364905be64af",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:100"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "a14457b34570e396c59208d49f40e1cc23b80a11",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:109"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:108",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:95",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:112]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:112",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:113]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:113",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:114]={
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
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:114",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:115]={
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
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:115",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/boards.Board"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:81"
//                     }
//                 }
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
//                     "ID": "gno.land/p/demo/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "9adae740a66b6f3d0066d3ddf478290286fd8a6e",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "630649048eb057bc03298c019417120984ad62b7",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:112"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "691da6b946b1873fba7f8dcf100e3e9d280e0759",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:113"
//                     }
//                 }
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
//                 "Hash": "91a60131fc119b8a406c4be71e1e991e37ca522b",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:114"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "39aedff3aa8d8246df681ff0f576f8ac31cb5d55",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:115"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107]={
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
//                     "ID": "gno.land/r/demo/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:101"
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107",
//         "ModTime": "116",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:116",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:117]={
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
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Escaped": true,
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:110"
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:117",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:116",
//         "RefCount": "1"
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:116]={
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
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "625e87966cb6643c1ad5ccd0a9ee1b0e3fe767fb",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "280f673fa84e1981a3138849f782b5dd87fbfd09",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:117"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:116",
//         "ModTime": "0",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:96",
//         "RefCount": "1"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:81]={
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
//                 "value": "g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "ecc6212ef0cb96a4785bb34427a7d0ee7f4c8400",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:82"
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
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "7e9fd9bb5e90a06c7751585cd80f23aedddde25b",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:83"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.MutTree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:81",
//         "IsEscaped": true,
//         "ModTime": "107",
//         "RefCount": "6"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:95]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "1a1636b4c1593d1694512b57780449f80c14d92f",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:108"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:95",
//         "ModTime": "107",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:94",
//         "RefCount": "1"
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:96]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "4df019c20a64842ebe5ae2cf1aed14cc7799588d",
//                         "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:116"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:96",
//         "ModTime": "107",
//         "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:94",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/users"]
// switchrealm["gno.land/r/demo/boards"]
// switchrealm["gno.land/r/boards_test"]
