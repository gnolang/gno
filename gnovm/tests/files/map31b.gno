// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
	i = 3
)

func main() {
	tmpInt := &i  // heapitem((3 int)) is dirty after this.
	m[tmpInt] = 1 // m is also dirty.
}

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](236)=
//     @@ -1,11 +1,39 @@
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
//     +                            "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5"
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
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     -        "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//     -        "RefCount": "1"
//     +        "ModTime": "7",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "N": "AwAAAAAAAAA=",
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "636",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1e746fff49b4e4d773aea460fece79334f62d235",
//                      "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3"
//                  }
//              },
//     @@ -44,7 +43,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8d8b8368884849377373fb1f9abd7add7f23f433",
//                      "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5"
//                  }
//              },
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
//     +            "Hash": "09381e8b65aa3811ed7d618c6e166cb091aa81d8",
//                  "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4"
//              }
//          }
