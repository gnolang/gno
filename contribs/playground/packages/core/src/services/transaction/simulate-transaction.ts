import { MsgEndpoint } from '@gnolang/gno-js-client'
import { ABCIEndpoint, newRequest, RestService, type ABCIResponse, type Tx } from '@gnolang/tm2-js-client'
import { extractSimulateFromResponse } from '@gnolang/tm2-js-client/bin/provider/utility/provider.utility'

import {
  encodeTransaction,
  makeAddPackageMessage,
  makeMsgCallMessage,
  makeMsgRunMessage,
  makeMsgSendMessage,
  TransactionBuilder,
  type TransactionMessage,
} from '@adena-wallet/sdk'

import { FeeToken } from '../../config'
import { ABCIError } from '../../utils'
import type { TransactionDocument } from './transaction-builder'

interface SimulateTransactionParams {
  rpcUrl: string
  transaction: TransactionDocument
}

export async function simulateTransaction({ rpcUrl, transaction }: SimulateTransactionParams) {
  const tx = documentToDefaultTx(transaction)
  const encodedTx = encodeTransaction(tx)

  try {
    const abciResponse = await RestService.post<ABCIResponse>(rpcUrl, {
      request: newRequest(ABCIEndpoint.ABCI_QUERY, ['.app/simulate', `${encodedTx}`, '0', false]),
    })

    const simulateResult = extractSimulateFromResponse(abciResponse)
    if (simulateResult.response_base?.error) {
      throw ABCIError.fromResponseDeliverTx(simulateResult)
    }

    return simulateResult
  } catch (err) {
    if (err instanceof ABCIError) throw err
    throw ABCIError.fromError(err)
  }
}

function documentToDefaultTx(document: TransactionDocument): Tx {
  const messages = document.messages.map(encodeMessageValue)

  const tx = TransactionBuilder.create()
  tx.messages(...messages)
  tx.memo(document.memo ?? '')
  tx.fee(document.gasFee, FeeToken.denom)
  tx.gasWanted(document.gasWanted)

  return {
    ...tx.build(),
    signatures: [
      {
        pub_key: {
          type_url: '',
          value: new Uint8Array(),
        },
        signature: new Uint8Array(),
      },
    ],
  }
}

function encodeMessageValue(message: { type: string; value: any }): TransactionMessage {
  switch (message.type as MsgEndpoint) {
    case MsgEndpoint.MSG_ADD_PKG: {
      return makeAddPackageMessage(message.value)
    }
    case MsgEndpoint.MSG_CALL: {
      const args: string[] = message.value.args ? (message.value.args.length === 0 ? null : message.value.args) : null
      return makeMsgCallMessage({
        ...message.value,
        args,
        send: message.value.send ?? '',
      })
    }
    case MsgEndpoint.MSG_SEND: {
      return makeMsgSendMessage(message.value)
    }
    case MsgEndpoint.MSG_RUN: {
      return makeMsgRunMessage({
        ...message.value,
        send: message.value.send ?? '',
      })
    }
    default: {
      throw new Error(`message type not supported`)
    }
  }
}
