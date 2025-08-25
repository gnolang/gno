// PKGPATH: gno.land/r/borrow
package borrow

type NoopMutator struct{ num int }

func (nm *NoopMutator) Mutate() {
	println("==========Mutate, maybe malicous...")
}

var root interface{}

func main() {
	var nm = &NoopMutator{num: 1}
	root = nm
	// ext.SetMutator(cross, nm) // attached to ext firstly.
}

// Realm:
// finalizerealm["gno.land/r/borrow"]
// c[c459c7103cd5303d426ed1e6e642fffcffd55f7c:7](214)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:7",
//         "LastObjectSize": "214",
//         "ModTime": "0",
//         "OwnerID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:6",
//         "RefCount": "1"
//     }
// }
// c[c459c7103cd5303d426ed1e6e642fffcffd55f7c:6](345)={
//     "ObjectInfo": {
//         "ID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:6",
//         "LastObjectSize": "345",
//         "ModTime": "0",
//         "OwnerID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/borrow.NoopMutator"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "90b7c42c3340e65db3a06771f64c2e3607bf1fef",
//             "ObjectID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:7"
//         }
//     }
// }
// u[c459c7103cd5303d426ed1e6e642fffcffd55f7c:3](205)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/borrow.NoopMutator"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "d4b5b3a875c50b7e84973e3f1fb5b08ba78acfc4",
//     +                "ObjectID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:6"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
