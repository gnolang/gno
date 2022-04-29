// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"strconv"

	"gno.land/r/boards"
)

var bid boards.BoardID
var pid boards.PostID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
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
// \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/3) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// > \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/4) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=4&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=4)]

// Realm:
// switchrealm["gno.land/r/boards"]
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:95]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000003"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:96"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000004"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000004"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "1523e12528ed6cb1f2cfbe6dacc8008d1df39424",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:95"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "28b9f2bc9d46acc743856e72dcbe958b10c3c79e",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/boards.Board"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/boards.Board"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:81"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/boards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": ""
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "Second reply of the second post"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "7e785fcbee72359096c1812574c0f54beb5ffc98",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:104"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "b716aa7cb58920eaac772d8c4f469d6e83a4fe3c",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:105"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "acab77847d1ea45ccadc1f4bf0a14d48974fc918",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:106"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/boards.PostID"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/boards.PostID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "std.Time"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000003"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:96"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000004"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/boards.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/boards.Post"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Escaped": true,
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:103"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "0000000004"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "af19b5fdbedf091b6895d5d8f3703680f3563ed1",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:100"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "53f452c8f3005506a28d0ddba65769488b308ea1",
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
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/boards.BoardID"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "/r/boards:test_board"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "test_board"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.MutTree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.MutTree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "e62f68b3047a509accd9054b14a7598617fd10a1",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:82"
//                     }
//                 }
//             }
//         },
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "65536"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "std.Time"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "23104b7967860eaac4ac8a3909b205a147caf8df",
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
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "f031357c04026165910a82d4fa5faf115b6ab00c",
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
// switchrealm["gno.land/r/boards"]
// switchrealm["gno.land/r/boards_test"]
