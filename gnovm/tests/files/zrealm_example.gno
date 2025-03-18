// PKGPATH: gno.land/r/example
package example

import (
	"gno.land/p/demo/dom"
)

var gPlot *dom.Plot

func init() {
	gPlot = &dom.Plot{Name: "First Plot"}
}

func main() {
	gPlot.AddPost("TEST_TITLE", "TEST_BODY")
	println(gPlot.String())
}

// Output:
// # [plot] First Plot
//
// ## TEST_TITLE
// TEST_BODY

// Realm:
// switchrealm["gno.land/r/example"]
// c[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:11]={
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
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "ModTime": "0",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:10]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "TEST_TITLE"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "TEST_BODY"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "0df7427101e263c586a307ff8d14d6b6d34117c4",
//                 "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "ModTime": "0",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     }
// }
// c[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:9]={
//     "ObjectInfo": {
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "ModTime": "0",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "122407cf2a9297da3d572f131cb647633df98900",
//             "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//         }
//     }
// }
// c[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/dom.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "398175e75cdd5013733e87c72156acad90f78adc",
//                     "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
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
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "ModTime": "0",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//         "RefCount": "1"
//     }
// }
// c[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:7]={
//     "ObjectInfo": {
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//         "ModTime": "0",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:6",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a864d30b0497eec996739b6740a387ece26d13c9",
//             "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:8"
//         }
//     }
// }
// u[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:6]={
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
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "f695e807587d8f5eafc96ec986d285d92ed31873",
//                     "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:7"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:6",
//         "ModTime": "6",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:5",
//         "RefCount": "1"
//     }
// }
// u[r:1ffd45e074aa1b8df562907c95ad97526b7ca187:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "First Plot"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "abc8b9cffe6efd08be3158fb611babf76389d52f",
//                 "ObjectID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:6"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:5",
//         "ModTime": "6",
//         "OwnerID": "r:1ffd45e074aa1b8df562907c95ad97526b7ca187:4",
//         "RefCount": "1"
//     }
// }
