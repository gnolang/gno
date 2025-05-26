// PKGPATH: gno.land/r/map_test
package map_test

var (
	m = map[*int]int{}
	i = 3 // real
)

func main() {
	tmpInt := &i
	m[tmpInt] = 1
}

// Realm:
// finalizerealm["gno.land/r/map_test"]
// u[02f0842ff638606e376be560152a77adbd377f56:4]=
//     @@ -1,10 +1,39 @@
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
//     +                            "ObjectID": "02f0842ff638606e376be560152a77adbd377f56:5"
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
//              "ID": "02f0842ff638606e376be560152a77adbd377f56:4",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:3",
//              "RefCount": "1"
//          }
// u[02f0842ff638606e376be560152a77adbd377f56:5]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "02f0842ff638606e376be560152a77adbd377f56:5",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "7",
//              "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "N": "AwAAAAAAAAA=",
