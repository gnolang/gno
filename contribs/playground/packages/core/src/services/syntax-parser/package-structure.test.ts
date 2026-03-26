import { expect, it } from 'vitest'

import { derivePackageStructure } from './package-structure'

it('should parse document tree', () => {
  expect(
    derivePackageStructure(`
    package std

    import (
      "fmt"
    )

    func DerivePkgAddr(pkgPath string) Address {
      return Address(derivePkgAddr(pkgPath))
    }

    func EncodeBech32(prefix string, bz [20]byte) Address {
      return Address(encodeBech32(prefix, bz))
    }

    func GetCallerAt(n int) Address {
      return Address(callerAt(n))
    }

    func GetOrigCaller() Address {
      return Address(origCaller())
    }

    func GetOrigPkgAddr() Address {
      return Address(origPkgAddr())
    }
    `),
  ).toMatchInlineSnapshot(`
    {
      "functions": [
        {
          "name": "DerivePkgAddr",
          "params": [
            {
              "name": "pkgPath",
              "type": "string",
            },
          ],
          "pos": {
            "from": 54,
            "to": 149,
          },
          "private": false,
          "return": "Address",
          "signature": "DerivePkgAddr(pkgPath string) Address",
        },
        {
          "name": "EncodeBech32",
          "params": [
            {
              "name": "prefix",
              "type": "string",
            },
            {
              "name": "bz",
              "type": "[20]byte",
            },
          ],
          "pos": {
            "from": 155,
            "to": 263,
          },
          "private": false,
          "return": "Address",
          "signature": "EncodeBech32(prefix string, bz [20]byte) Address",
        },
        {
          "name": "GetCallerAt",
          "params": [
            {
              "name": "n",
              "type": "int",
            },
          ],
          "pos": {
            "from": 269,
            "to": 342,
          },
          "private": false,
          "return": "Address",
          "signature": "GetCallerAt(n int) Address",
        },
        {
          "name": "GetOrigCaller",
          "params": [],
          "pos": {
            "from": 348,
            "to": 419,
          },
          "private": false,
          "return": "Address",
          "signature": "GetOrigCaller() Address",
        },
        {
          "name": "GetOrigPkgAddr",
          "params": [],
          "pos": {
            "from": 425,
            "to": 498,
          },
          "private": false,
          "return": "Address",
          "signature": "GetOrigPkgAddr() Address",
        },
      ],
      "imports": [
        "fmt",
      ],
      "package": "std",
    }
  `)
})
