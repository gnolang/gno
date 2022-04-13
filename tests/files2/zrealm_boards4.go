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
// - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2) [reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)
//
// > Reply of the second post
// > - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#3) [reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)
//
// > Second reply of the second post
// > - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#4) [reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=4&body.type=textarea)

// Realm:
// switchrealm["gno.land/r/boards"]
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:80]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:81"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:80",
//         "ModTime": "83",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:83",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:84]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:85"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:84",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:83",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:83]={
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
//                         "Hash": "1027b5b2211609578a46dd8241aa07d04e434b00",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:80"
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
//                         "Hash": "b2a6eda59e1ba6bf69a410e1c4b906c6d15595bb",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:84"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:83",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:79",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:85]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:73"
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
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:85",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:82]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:81"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:82",
//         "ModTime": "86",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:86",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:87]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:85"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:87",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:86",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:86]={
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
//                         "Hash": "6a6ed28f4d1db8b7bb9e44ea5c094b29ed3f4b02",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:82"
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
//                         "Hash": "c9dad8cb9d0ce381fbfef1b3738385d36ea979ff",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:87"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:86",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:79",
//         "RefCount": "1"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:73]={
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
//                         "Hash": "5d6955444233de20d2860a94a4a98cb9826bbf4f",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:77"
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
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:73",
//         "IsEscaped": true,
//         "ModTime": "82",
//         "RefCount": "6"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:79]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:73"
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
//                 "value": "Second Post (title)"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "Body of the second post. (body)"
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
//                         "Hash": "49040fae4f9f8f4a27a362cf07f6755c80b05da0",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:83"
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
//                         "Hash": "14ff1264ad7901f3365b7eaf90f8247b627bab77",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:86"
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
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:79",
//         "ModTime": "82",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:78",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/users"]
// switchrealm["gno.land/r/boards"]
// switchrealm["gno.land/r/boards_test"]
