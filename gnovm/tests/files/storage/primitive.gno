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
// gno.land/r/xx: 866

// Realm:
// finalizerealm["gno.land/r/xx"]
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
//         "LastObjectSize": "253",
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
//         "LastObjectSize": "303",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//         "RefCount": "1"
//     }
// }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:3](5)=
//     @@ -2,11 +2,12 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:3",
//              "LastObjectSize": "217",
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
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:4](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//              "LastObjectSize": "237",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
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
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:5](30)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5",
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,10 @@
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
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:6](138)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//              "LastObjectSize": "236",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
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
