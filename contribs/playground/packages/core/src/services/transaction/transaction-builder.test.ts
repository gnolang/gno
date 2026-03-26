import { expect, it } from 'vitest'

import { TransactionBuilder } from './transaction-builder'

it('should create a addpkg transaction with two messages', () => {
  const transactionBuilder = new TransactionBuilder()
  transactionBuilder.addPkg({
    creator: 'gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz',
    deposit: '1000uatom',
    data: {
      name: 'hello',
      path: 'gno.land/r/demo/hello',
      files: [
        {
          name: 'contract.gno',
          body: 'package hello',
        },
      ],
    },
  })

  transactionBuilder.setGas({ gasFee: 100, gasWanted: 50000 })

  transactionBuilder.addPkg({
    creator: 'gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz',
    deposit: '1000uatom',
    data: {
      name: 'new',
      path: 'gno.land/r/demo/new',
      files: [
        {
          name: 'contract.gno',
          body: 'package new',
        },
        {
          name: 'contract_test.gno',
          body: 'package new_test',
        },
      ],
    },
  })

  const result = transactionBuilder.build()
  expect(result).toMatchInlineSnapshot(`
    {
      "gasFee": 100,
      "gasWanted": 50000,
      "memo": "Deployed through play.gno.land",
      "messages": [
        {
          "type": "/vm.m_addpkg",
          "value": {
            "creator": "gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz",
            "deposit": "1000uatom",
            "package": {
              "files": [
                {
                  "body": "package hello",
                  "name": "contract.gno",
                },
              ],
              "name": "hello",
              "path": "gno.land/r/demo/hello",
            },
          },
        },
        {
          "type": "/vm.m_addpkg",
          "value": {
            "creator": "gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz",
            "deposit": "1000uatom",
            "package": {
              "files": [
                {
                  "body": "package new",
                  "name": "contract.gno",
                },
                {
                  "body": "package new_test",
                  "name": "contract_test.gno",
                },
              ],
              "name": "new",
              "path": "gno.land/r/demo/new",
            },
          },
        },
      ],
    }
  `)

  expect(result.messages.length).toBe(2)
  expect(transactionBuilder.buildCommand()).toMatchObject([
    `gnokey maketx addpkg gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz -gas-fee="100ugnot" -gas-wanted="50000" -broadcast=true -memo="Deployed through play.gno.land" -deposit="1000uatom" -name="hello" -pkgpath="gno.land/r/demo/hello" -pkgdir="./"`,
    `gnokey maketx addpkg gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz -gas-fee="100ugnot" -gas-wanted="50000" -broadcast=true -memo="Deployed through play.gno.land" -deposit="1000uatom" -name="new" -pkgpath="gno.land/r/demo/new" -pkgdir="./"`,
  ])
})

it('should create a call transaction', () => {
  const transactionBuilder = new TransactionBuilder()
  transactionBuilder.call({
    caller: 'gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz',
    send: '',
    pkg_path: 'gno.land/r/demo/hello',
    func: 'hello',
    args: ['hello', '123'],
  })

  transactionBuilder.setGas({ gasFee: 100, gasWanted: 50000 })

  expect(transactionBuilder.build()).toMatchObject({
    gasFee: 100,
    gasWanted: 50000,
    memo: 'Deployed through play.gno.land',
    messages: [
      {
        type: '/vm.m_call',
        value: {
          args: ['hello', '123'],
          caller: 'gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz',
          func: 'hello',
          pkg_path: 'gno.land/r/demo/hello',
          send: '',
        },
      },
    ],
  })

  expect(transactionBuilder.buildCommand({ rpcUrl: '127.0.0.1:26657' })).toMatchObject([
    'gnokey maketx call gno1xv9tk6v3qzjzrccg9y2zv6y2y9z7z5h6f7y0jz -gas-fee="100ugnot" -gas-wanted="50000" -broadcast=true -memo="Deployed through play.gno.land" -remote="127.0.0.1:26657" -send="" -pkgpath="gno.land/r/demo/hello" -func="hello" -args="hello" -args="123"',
  ])
})
