// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register("gnouser123")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	rid := boards.CreateReply(bid, pid, pid, "Reply of the second post")
	println(rid)
}

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
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
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=2&threadid=2)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)]
//
// > Reply of the second post
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// > Second reply of the second post
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/4) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=4&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=4&threadid=2)]
//

// Realm:
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/demo/boards"]
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111]=
//     @@ -1,8 +1,8 @@
//      {
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111",
//     -        "ModTime": "0",
//     -        "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106",
//     +        "ModTime": "123",
//     +        "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123",
//              "RefCount": "1"
//          },
//          "Value": {
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
//             "Hash": "3f34ac77289aa1d5f9a2f8b6d083138325816fb0",
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
//                     "Hash": "94a6665a44bac6ede7f3e3b87173e537b12f9532",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "bc8e5b4e782a0bbc4ac9689681f119beb7b34d59",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "9957eadbc91dd32f33b0d815e041a32dbdea0671",
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
//             "Hash": "c45bbd47a46681a63af973db0ec2180922e4a8ae",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:127"
//         }
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120]=
//     @@ -1,8 +1,8 @@
//      {
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120",
//     -        "ModTime": "0",
//     -        "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107",
//     +        "ModTime": "134",
//     +        "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134",
//              "RefCount": "1"
//          },
//          "Value": {
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
//             "Hash": "96b86b4585c7f1075d7794180a5581f72733a7ab",
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
//                     "Hash": "32274e1f28fb2b97d67a1262afd362d370de7faa",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "c2cfd6aec36a462f35bf02e5bf4a127aa1bb7ac2",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5cb875179e86d32c517322af7a323b2a5f3e6cc5",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134"
//         }
//     }
// }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85]=
//     @@ -49,7 +49,7 @@
//                  }
//              },
//              {
//     -            "N": "AwAAAAAAAAA=",
//     +            "N": "BAAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "65536"
//     @@ -80,7 +80,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85",
//     -        "ModTime": "110",
//     +        "ModTime": "121",
//              "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84",
//              "RefCount": "1"
//          }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106]=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "3d6aa1e96f126aba3cfd48f0203b85a287a6f1c0",
//     -                    "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:111"
//     +                    "Hash": "9809329dc1ddc5d3556f7a8fa3c2cebcbf65560b",
//     +                    "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:122"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -22,7 +22,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:106",
//     -        "ModTime": "110",
//     +        "ModTime": "121",
//              "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:105",
//              "RefCount": "1"
//          }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107]=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "96a7d0b7b5bf5d366d10c559464ba56a50f78bbc",
//     -                    "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:120"
//     +                    "Hash": "ceae9a1c4ed28bb51062e6ccdccfad0caafd1c4f",
//     +                    "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:133"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -22,7 +22,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:107",
//     -        "ModTime": "110",
//     +        "ModTime": "121",
//              "OwnerID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:105",
//              "RefCount": "1"
//          }
// u[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84]=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:84",
//              "IsEscaped": true,
//     -        "ModTime": "114",
//     -        "RefCount": "5"
//     +        "ModTime": "127",
//     +        "RefCount": "6"
//          },
//          "Value": {
//              "T": {
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "058a49e6ee26f75cc44901754d4c298fd1a1655c",
//     +            "Hash": "a88a9b837af217656ee27084309f7cd02cd94cb3",
//                  "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:85"
//              }
//          }
// u[336074805fc853987abe6f7fe3ad97a6a6f3077a:2]=
//     @@ -3,8 +3,8 @@
//          "ObjectInfo": {
//              "ID": "336074805fc853987abe6f7fe3ad97a6a6f3077a:2",
//              "IsEscaped": true,
//     -        "ModTime": "118",
//     -        "RefCount": "11"
//     +        "ModTime": "131",
//     +        "RefCount": "12"
//          },
//          "Parent": null,
//          "Source": {
// finalizerealm["gno.land/r/demo/boards"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/sys/users"]
// finalizerealm["gno.land/r/demo/boards"]
// finalizerealm["gno.land/r/demo/boards_test"]
