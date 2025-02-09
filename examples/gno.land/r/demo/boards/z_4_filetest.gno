// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	rid := boards.CreateReply(bid, pid, pid, "Reply of the second post")
	println(rid)
}

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
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
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=2&threadid=2)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// > Second reply of the second post
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/4) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=4&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=4&threadid=2)]
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "dedb49c348bedfeb8e560cc1d6d230093b4b4d70",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:112"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:125]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "4"
//             },
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a0c0fc7d5ac90cd80f865fe5461937c3a977dd89",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:125"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:123]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "2a6696d77d94cd3c9c6948c13a7981ce6466e5fa",
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
//                     "Hash": "cc4b292ae242cb7a9106a2cf4cd550e0a7068edf",
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5c2698b640b75007bc4286e6ce9ec5f91eed91e8",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "11"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "617e5b805ad827e105c13c6dcaeb65cc53c6c842",
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/boards.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6c1213aad1d31a560b2ef29727b77fa81f4b559e",
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "551943b450e02af21e71eb9af143211ca11ef6ac",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:121"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:136]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "4"
//             },
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3be68992404dd0459755435b1a93b4b06880349b",
//             "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:136"
//         }
//     }
// }
// c[f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:134]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "d2de77b22a5800c4074b114a4edead28806979ae",
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
//                     "Hash": "aa005854a2259624c21d4a11433fbea479a73f42",
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "8063bc5838316297f6991918a9a6b376f76f0329",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ee55adef07d09bbb491ec92f277b1baeb287b036",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "time.Time"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "08de9a4baa4c4355239148b6ef9f743961f0e452",
//                 "ObjectID": "f6dbf411da22e67d74cd7ddba6a76cd7e14a4822:87"
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "256b7b5523037d9ed40f602c906cf3c3c80b91d2",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "c9f3fb4672c03a1ee7fa14701800803cc2dd98ab",
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
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/boards.Board"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "ebb8fbaa23097343dfb73f0998c708aeb9b96a42",
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
