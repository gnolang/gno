// PKGPATH: gno.land/r/xx
package xx

var i int8
var is []int8
var s string
var ss []string

func main() {
	crossing()
	i = 127
	is = []int8{0, 127}
	s = "a"
	ss = []string{"a", "aa"}

}

// Storage:
// gno.land/r/xx: 854

// Realm:
// switchrealm["gno.land/r/xx"]
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:9](253)={
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
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//         "RefCount": "1"
//     }
// }
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:10](303)={
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
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//         "RefCount": "1"
//     }
// }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:3](2)=
//     @@ -1,11 +1,12 @@
//      {
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:3",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "fwAAAAAAAAA=",
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "64"
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:4](134)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,17 @@
//                      "value": "64"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "e84e8bceb6edffed359aabd64d88ad4cdaae23c8",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:5](27)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -9,6 +9,10 @@
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "16"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.StringValue",
//     +            "value": "a"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:6](135)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,17 @@
//                      "value": "16"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "af828a01b9823a07083d08cb1c69bbe1f87c0edb",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
