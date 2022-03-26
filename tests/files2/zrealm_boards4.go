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
	rid := boards.CreateReply(bid, pid, "Reply of the second post")
	println(rid)
}

func main() {
	rid2 := boards.CreateReply(bid, pid, "Second reply of the second post")
	println(rid2)
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// 3
// 4
// # Second Post (title)
//
// Body of the second post. (body)
// - by g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2)
//
// > Reply of the second post
// > - by g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#3)
//
// > Second reply of the second post
// > - by g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#4)
//

// Realm:
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:29]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "3"
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:30"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:29",
//         "ModTime": "33",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:34]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "4"
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:35"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:34",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:33]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "4"
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
//                         "Hash": "4b762ed5352e7cd51e2d4a5129d154dd682ee1bf",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:29"
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
//                         "Hash": "5cafaf74856139765f1e1a0e84f01976eba9c151",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:34"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:27",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:36]={
//     "Data": "dGVzdGFkZHJfX19fX19fX19fX18=",
//     "List": null,
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:36",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:35",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:35]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:19"
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
//                 "@type": "/gno.vref",
//                 "Hash": "71fc8e70c779795f90c6cafa7071490607247bc4",
//                 "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:36"
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
//                 "@type": "/gno.tpri",
//                 "value": "1024"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:35",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:32]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "3"
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:30"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:32",
//         "ModTime": "37",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:37",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:38]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "4"
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:35"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:38",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:37",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:37]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "4"
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
//                         "Hash": "d2384f62b8659782be261034b6a00f31c58799d4",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:32"
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
//                         "Hash": "eb7453f8c95cbfed339f797c681a2bfe3d4c9605",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:38"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:37",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:27",
//         "RefCount": "1"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:19]={
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
//                 "@type": "/gno.vref",
//                 "Hash": "de9a667b3ba0569f8a51108eea2f1416059cf27b",
//                 "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:20"
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
//                         "Hash": "d71c2ee39d23c55743b3190d0483adace37dc006",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:22"
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
//                 "@type": "/gno.tpri",
//                 "value": "1024"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:19",
//         "IsEscaped": true,
//         "ModTime": "32",
//         "RefCount": "6"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:27]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:19"
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
//                 "@type": "/gno.vref",
//                 "Hash": "78a635aa056e7f53002ef29a75b4f6c665906099",
//                 "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:28"
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
//                         "Hash": "f46ca00e4c2b6affdc6a6a030bc28e82739d85ea",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33"
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
//                         "Hash": "688e4f38547a80d3ebbe976f9caf1175c3fe8066",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:37"
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
//                 "@type": "/gno.tpri",
//                 "value": "1024"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:27",
//         "ModTime": "32",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:26",
//         "RefCount": "1"
//     }
// }
