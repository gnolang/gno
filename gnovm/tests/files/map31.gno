// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
)

func main() {
	i := 3
	tmpInt := &i
	m[tmpInt] = 1
}

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0](220)={
//     "ObjectInfo": {
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0",
//         "LastObjectSize": "220",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0",
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
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0](239)=
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
//     +                            "Hash": "c3c2be620ddd8f708a6b3ceac9dc4a368262d013",
//     +                            "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0"
//     +                        },
//     +                        "Index": "0",
//     +                        "TV": null
//     +                    }
//     +                },
//     +                "Value": {
//     +                    "N": "AQAAAAAAAAA=",
//     +                    "T": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "32"
//     +                    }
//     +                }
//     +            }
//     +        ]
//          },
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0",
//              "LastObjectSize": "179",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3:0",
//              "RefCount": "1"
//          }
