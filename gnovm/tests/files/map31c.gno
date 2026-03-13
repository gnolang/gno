// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
)

func main() {
	i := 3
	tmpInt := &i
	m[tmpInt] += 2
	println(m[tmpInt])
}

// Output:
// 2

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7](216)={
//     "ObjectInfo": {
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "N": "AwAAAAAAAAA=",
//         "T": {
//             "@type": "/gno.PrimitiveType",
//             "value": "32"
//         }
//     }
// }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](236)=
//     @@ -1,11 +1,40 @@
//      {
//          "List": {
//     -        "List": []
//     +        "List": [
//     +            {
//     +                "Key": {
//     +                    "T": {
//     +                        "@type": "/gno.PointerType",
//     +                        "Elt": {
//     +                            "@type": "/gno.PrimitiveType",
//     +                            "value": "32"
//     +                        }
//     +                    },
//     +                    "V": {
//     +                        "@type": "/gno.PointerValue",
//     +                        "Base": {
//     +                            "@type": "/gno.RefValue",
//     +                            "Hash": "707048547d382f3872ac35d41c3bfc2ac692d6ad",
//     +                            "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7"
//     +                        },
//     +                        "Index": "0",
//     +                        "TV": null
//     +                    }
//     +                },
//     +                "Value": {
//     +                    "N": "AgAAAAAAAAA=",
//     +                    "T": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "32"
//     +                    }
//     +                }
//     +            }
//     +        ]
//          },
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "LastObjectSize": "388",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "RefCount": "1"
//          },
//     @@ -23,7 +23,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "57424a2a4c05dd2e2f21e0c76ab122ccd15830ce",
//     +            "Hash": "89f613589f783f65f67af7f03f4a63bc4a0a4dff",
//                  "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4"
//              }
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "497",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d9bfc05ece20a632345b370988b3fb6a5afd7651",
//     +                "Hash": "59d32a9a4e2e876f6a7390e8d4cd830c27bbf72b",
//                      "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3"
//                  }
//              },
