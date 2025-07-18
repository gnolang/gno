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
// gno.land/r/xx: 878

// Realm:
// finalizerealm["gno.land/r/xx"]
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:9:0](257)={
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
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9:0",
//         "LastObjectSize": "257",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4:0",
//         "RefCount": "1"
//     }
// }
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:10:0](307)={
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
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10:0",
//         "LastObjectSize": "307",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6:0",
//         "RefCount": "1"
//     }
// }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:3:0](5)=
//     @@ -2,11 +2,12 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:3:0",
//              "LastObjectSize": "221",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2:0",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "fwAAAAAAAAA=",
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "64"
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:4:0](139)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4:0",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2:0",
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
//     +                "Hash": "a422f3e9664a96f30c10801185e5c606e33c1dfe",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9:0"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:5:0](30)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5:0",
//              "LastObjectSize": "220",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2:0",
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
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:6:0](140)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6:0",
//              "LastObjectSize": "240",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2:0",
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
//     +                "Hash": "780640f2109c4d1bcddb5195ba2c0b8a16b9e1fc",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10:0"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
