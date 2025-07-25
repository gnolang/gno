// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
	i = 3 // real
)

func main() {
	tmpInt := &i
	m[tmpInt] = 1
}

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](239)=
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
//     +                            "Escaped": true,
//     +                            "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5"
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
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "N": "AwAAAAAAAAA=",
