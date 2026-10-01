// PKGPATH: gno.land/r/test
// A recovered VM runtime error (.runtimeError) persists in realm state: init
// stores it, the store is committed and the package reloaded (mimicking
// on-chain behavior), and main still sees an error value whose Error() works.
// Re-storing it in main locks the serialized form in the Realm golden.
package test

var (
	saved  error
	saved2 error
)

func init() {
	defer func() {
		saved = recover().(error)
	}()
	var p *int
	_ = *p
}

func main(cur realm) {
	println(saved.Error())
	saved2 = saved
}

// Output:
// runtime error: nil pointer dereference

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](108)=
//     @@ -2,9 +2,18 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": ".uverse..runtimeError"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.StringValue",
//     +            "value": "runtime error: nil pointer dereference"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c179e3d792fe6624f9de9aed3795cb7b597a4eac",
//     +                "Hash": "5ee2184d5765dce39681e4bc16225230f780ee6b",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
