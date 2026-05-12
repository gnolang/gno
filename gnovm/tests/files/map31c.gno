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
// c[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7](216)={
//     "ObjectInfo": {
//         "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
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
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](236)=
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
//     +                            "Hash": "155b33746175a490e1a154d886189491a2e224fe",
//     +                            "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7"
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
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "LastObjectSize": "388",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "RefCount": "1"
//          },
//     @@ -23,7 +23,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "dfbcb2822726b9dfd148ba20845a654ebfb174ce",
//     +            "Hash": "99dd92d11b6bf04b25032b17d3b55c2775efee35",
//                  "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4"
//              }
//          }
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
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
//     -                "Hash": "1e746fff49b4e4d773aea460fece79334f62d235",
//     +                "Hash": "77251523bc49b9e32bc691608d5ff0be0a84c1cb",
//                      "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3"
//                  }
//              },
