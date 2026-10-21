// PKGPATH: gno.land/r/test
package test

import "math"

type Account struct {
	Balance uint256
	Debt    int256
}

// An int256 constant must fit in an int64: wider int256 values are converted
// from uint256 values at run time.
var minBits uint256 = 1 << 255

var (
	supply   uint256
	delta    int256
	zero     uint256
	acct     Account
	accounts = map[string]*Account{}
	history  [3]int256
	amounts  []uint256
	ptr      *uint256
)

func init() {
	supply = math.MaxUint256
	delta = int256(minBits) // an int256 constant must fit in an int64
	acct = Account{Balance: 1 << 200, Debt: -1}
	accounts["alice"] = &Account{Balance: 7, Debt: -7}
	history = [3]int256{-1, 0, 1}
	amounts = []uint256{1, 1 << 128}
	v := uint256(42)
	ptr = &v
}

func main(cur realm) {
	// The values written by init were persisted and are read back here.
	println(supply, delta, zero)
	println(acct.Balance, acct.Debt)
	println(accounts["alice"].Balance, accounts["alice"].Debt)
	println(history[0], history[1], history[2], amounts[0], amounts[1], *ptr)

	supply--
	delta++
	acct.Debt *= 2
	accounts["alice"].Balance <<= 250
	history[1] = int256(minBits - 1)
	amounts = append(amounts, 3)
	*ptr = 0
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935 -57896044618658097711785492504343953926634992332820282019728792003956564819968 0
// 1606938044258990275541962092341162602522202993782792835301376 -1
// 7 -7
// -1 0 1 1 340282366920938463463374607431768211456 42

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:24](416)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Uint256Value",
//                 "value": "1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Uint256Value",
//                 "value": "340282366920938463463374607431768211456"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Uint256Value",
//                 "value": "3"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:24",
//         "LastObjectSize": "416",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "328",
//     -        "ModTime": "17",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.Uint256Value",
//     -            "value": "115792089237316195423570985008687907853269984665640564039457584007913129639935"
//     +            "value": "115792089237316195423570985008687907853269984665640564039457584007913129639934"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "327",
//     -        "ModTime": "17",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.Int256Value",
//     -            "value": "-57896044618658097711785492504343953926634992332820282019728792003956564819968"
//     +            "value": "-57896044618658097711785492504343953926634992332820282019728792003956564819967"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:18](5)=
//     @@ -17,14 +17,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.Int256Value",
//     -                "value": "-1"
//     +                "value": "-2"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//              "LastObjectSize": "371",
//     -        "ModTime": "0",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:20](82)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.Uint256Value",
//     -                "value": "7"
//     +                "value": "12664759760331458874453076485325239921451404572804436691815673250865498554368"
//                  }
//              },
//              {
//     @@ -24,7 +24,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20",
//              "LastObjectSize": "312",
//     -        "ModTime": "0",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:21](107)=
//     @@ -15,6 +15,10 @@
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "2097152"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.Int256Value",
//     +                "value": "57896044618658097711785492504343953926634992332820282019728792003956564819967"
//                  }
//              },
//              {
//     @@ -31,7 +35,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:21",
//              "LastObjectSize": "351",
//     -        "ModTime": "0",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:13](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//              "LastObjectSize": "378",
//     -        "ModTime": "17",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,11 +19,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "873fbbc56ab9256035042812eccd9db86e6180df",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22"
//     +                "Hash": "4a82cee4c2713d88155e3795facd713abda26c59",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:24"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:23](-22)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23",
//              "LastObjectSize": "248",
//     -        "ModTime": "0",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "RefCount": "1"
//          },
//     @@ -10,10 +10,6 @@
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "4194304"
//     -        },
//     -        "V": {
//     -            "@type": "/gno.Uint256Value",
//     -            "value": "42"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1886",
//     -        "ModTime": "23",
//     +        "ModTime": "24",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3bc4d7e91a6d35c0c3a38790c1a8565998f262ff",
//     +                "Hash": "a6dd37f847d2de5d1c236678321787b328d0cf95",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
//     @@ -66,7 +66,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9340b624921d2f39b6e26e585f5ef9701f93f149",
//     +                "Hash": "63e19efe134ef77e5dd737c3bdc1b9cac315f323",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
//     @@ -86,7 +86,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "069707bdd4803ad8107ee933dcafe2b1cb51d838",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  }
//              },
//     @@ -105,7 +104,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b9f902a459d35af0fc089e612d643c7103f769da",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                  }
//              },
//     @@ -115,7 +113,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f617a56cbc12cda118d0c9453fe65d4e51ced7ae",
//     +                "Hash": "72a3604a9db43faf96c12b47b343c4ccef082144",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//                  }
//              },
//     @@ -125,7 +123,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b22bfd667b26ee6dff5bae12e523b696fb3ce422",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//                  }
//              },
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "345",
//     -        "ModTime": "17",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2024daf563570d4292fff4271242f987e502d7a3",
//     +            "Hash": "9eafceacd941df99972ab9c7bda9e42141b1f02c",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:19](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//              "LastObjectSize": "342",
//     -        "ModTime": "0",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "dfaa0fcb6da51a1b372437134f21f8bc0796d4b0",
//     +            "Hash": "483ad8a56d38bd600934e308a246fb61bf4b8f30",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](0)=
//     @@ -24,7 +24,7 @@
//                              "@type": "/gno.PointerValue",
//                              "Base": {
//                                  "@type": "/gno.RefValue",
//     -                            "Hash": "d269f1e734f7c9683c4398c431c009c5c1702108",
//     +                            "Hash": "8ef38efc1d19937af0073e73f82d520cfeb0b4a3",
//                                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19"
//                              },
//                              "Index": "0",
//     @@ -37,7 +37,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "459",
//     -        "ModTime": "17",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:9](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "411",
//     -        "ModTime": "23",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -23,7 +23,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "82605c406465fcf67c9f4fc0fd2d09594f56a4e6",
//     +            "Hash": "c73070b4d638a05795de0f259057457bd40ebcc6",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:11](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "LastObjectSize": "354",
//     -        "ModTime": "17",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "cc1faca9caaf65e5d351bc21fd4634ef1cc84923",
//     +            "Hash": "3b7375ee3b7f2b83ad5a857ae32427b94e28e8ce",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:21"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:14](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "LastObjectSize": "377",
//     -        "ModTime": "17",
//     +        "ModTime": "24",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2e1e2a8fdf70bde0ab875d8fcea94f1ba7a7ceeb",
//     +                "Hash": "61ac5f83298bba6c64595b05a0d30504da6d040d",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23"
//                  },
//                  "Index": "0",
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:22](-349)
