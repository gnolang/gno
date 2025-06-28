// PKGPATH: gno.land/r/map_test
package map_test

var (
	m = map[*int]int{}
)

func main() {
	i := 3
	tmpInt := &i
	m[tmpInt] = 1
}

// Realm:
// finalizerealm["gno.land/r/map_test"]
// c[02f0842ff638606e376be560152a77adbd377f56:7]={
//     "ObjectInfo": {
//         "ID": "02f0842ff638606e376be560152a77adbd377f56:7",
//         "ModTime": "0",
//         "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:4",
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
//     +                            "Hash": "e6404f3da482cc2c2d6d9bab5cd8a092e91c9d25",
//     +                            "ObjectID": "02f0842ff638606e376be560152a77adbd377f56:7"
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
//     +        "ModTime": "6",
//              "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:3",
//              "RefCount": "1"
//          }
