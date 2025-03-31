// PKGPATH: gno.land/r/xx
package xx

var i int8
var is []int8
var s string
var ss []string

func main() {
	i = 127
	is = []int8{0, 127}
	s = "a"
	ss = []string{"a", "aa"}

}

// Storage:
// gno.land/r/xx: 844

// Realm:
// switchrealm["gno.land/r/xx"]
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:4](253)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "fwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//         "RefCount": "1"
//     }
// }
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:5](302)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "a"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "aa"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//         "RefCount": "1"
//     }
// }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:2](289)=
//     @@ -19,6 +19,7 @@
//          },
//          "Values": [
//              {
//     +            "N": "fwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "64"
//     @@ -32,12 +33,27 @@
//                          "value": "64"
//                      },
//                      "Vrd": false
//     +            },
//     +            "V": {
//     +                "@type": "/gno.SliceValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "c56056975e482a4315531d03c1540f79c9eaf42f",
//     +                    "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4"
//     +                },
//     +                "Length": "2",
//     +                "Maxcap": "2",
//     +                "Offset": "0"
//                  }
//              },
//              {
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "16"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.StringValue",
//     +                "value": "a"
//                  }
//              },
//              {
//     @@ -48,6 +64,17 @@
//                          "value": "16"
//                      },
//                      "Vrd": false
//     +            },
//     +            "V": {
//     +                "@type": "/gno.SliceValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "666b921823ae903b9b40435f947775fbd7dd5df1",
//     +                    "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5"
//     +                },
//     +                "Length": "2",
//     +                "Maxcap": "2",
//     +                "Offset": "0"
//                  }
//              },
//              {
