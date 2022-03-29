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
// switchrealm["gno.land/r/boards"]
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:26]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:27"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:26",
//         "ModTime": "29",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:29",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:30]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:31"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:30",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:29",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:29]={
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
//                         "Hash": "1c0bf19b029308c6195b4f362b898dbf1839c842",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:26"
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
//                         "Hash": "a7b7c09f69531d3834b7032acc4cdba63714365a",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:30"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:29",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:25",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:31]={
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
//                 "@type": "/gno.vstr",
//                 "value": "g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:31",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:28]={
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:27"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:28",
//         "ModTime": "32",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:32",
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
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:31"
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:32",
//         "RefCount": "1"
//     }
// }
// c[960d1737342909c1a4c32a4a93a88e680a6f79df:32]={
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
//                         "Hash": "21b328992dc37a61e69e08362762e8711cd19faf",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:28"
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
//                         "Hash": "53d759af01546d07e7c7eae0cc184740300c368d",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:33"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:32",
//         "ModTime": "0",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:25",
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
//                 "@type": "/gno.vstr",
//                 "value": "g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj"
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
//                         "Hash": "2587e47ad0bbfa2215fc6b4a717422ce6a74fe21",
//                         "ObjectID": "960d1737342909c1a4c32a4a93a88e680a6f79df:21"
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
//         "ModTime": "28",
//         "RefCount": "6"
//     }
// }
// u[960d1737342909c1a4c32a4a93a88e680a6f79df:25]={
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
//                 "@type": "/gno.vstr",
//                 "value": "g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj"
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
//                         "Hash": "c25ce15e32fe42888bff02631b6c123a5a07e30c",
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
//                         "Hash": "24ace6c9e354bdbf83ac29da85878f3233a86de7",
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
//         "ID": "960d1737342909c1a4c32a4a93a88e680a6f79df:25",
//         "ModTime": "28",
//         "OwnerID": "960d1737342909c1a4c32a4a93a88e680a6f79df:24",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/boards"]
