import { decodeTxMessages, MsgEndpoint } from '@gnolang/gno-js-client'
import { base64ToUint8Array, newRequest, RestService, Tx, type DeliverTx } from '@gnolang/tm2-js-client'

import type { CallDetailsResponse, RunDetailsResponse } from '../../types'
import type { CallMessageInput, RunMessageInput } from './transaction-types'

export { MsgEndpoint } from '@gnolang/gno-js-client'

interface TxResult {
  height: number
  tx_result: DeliverTx
  tx: string
}

interface GetTransactionParams {
  /**
   * Transaction hash.
   */
  hash: string

  /**
   * Node RPC URL.
   */
  rpcUrl: string
}

export type WaitForTransactionParams = GetTransactionParams & { timeout?: number }

async function getTransactionResult({ hash, rpcUrl }: GetTransactionParams) {
  return await RestService.post<TxResult>(rpcUrl, { request: newRequest('tx', [hash]) })
}

async function waitForTransaction({ hash, rpcUrl, timeout = 15000 }: WaitForTransactionParams): Promise<TxResult> {
  return await new Promise((resolve, reject) => {
    const interval = setInterval(async () => {
      try {
        const response = await getTransactionResult({ hash, rpcUrl })
        resolve(response)
      } catch (_error) {
        console.log(`Transaction "${hash}" not found, waiting...`)
      }
    }, 1500)

    setTimeout(() => {
      clearInterval(interval)
      reject(new Error(`Timeout while waiting for transaction "${hash}" to be included in a block`))
    }, timeout)
  })
}

export type TransactionMessage<T extends object> = {
  '@type': string
} & T

/**
 * Retrieves transaction information and its messages using a TX hash.
 */
export const getTransactionMessages = async <T extends object>(params: WaitForTransactionParams) => {
  const response = await waitForTransaction(params)

  const decodedTx = Tx.decode(base64ToUint8Array(response.tx))
  const messages: Array<TransactionMessage<T>> = decodeTxMessages(decodedTx.messages)

  return { response, messages }
}

/**
 * Gets transaction messages and returns a message matching a specified type.
 *
 * @param msgType Message type to find in transaction messages.
 * @param params Transaction search params.
 * @returns CallDetailsResponse
 */
const getTransactionMessageByType = async <T extends object>(msgType: string, params: WaitForTransactionParams) => {
  const { response, messages } = await getTransactionMessages<T>(params)

  // Call message can be expression call or run.
  const message = messages.find((msg) => msg['@type'] === msgType)
  if (!message) {
    throw new Error(`Original "${msgType}" message not found in a transaction`)
  }

  return { response, message }
}

/**
 * Retrieves call message details from a transaction.
 */
export const getCallResultFromNode = async (params: WaitForTransactionParams): Promise<CallDetailsResponse> => {
  const { response, message } = await getTransactionMessageByType<CallMessageInput>(MsgEndpoint.MSG_CALL, params)
  const { send, caller, func, pkg_path: pkgPath, args } = message

  return {
    send,
    caller,
    func,
    pkgPath,
    height: Number(response.height),
    args: args?.map((arg: string, index: number) => ({ index, value: arg })) ?? [],
    deliverTx: response.tx_result,
  }
}

/**
 * Retrieves run message details from a transaction
 */
export const getRunResultFromNode = async (params: WaitForTransactionParams): Promise<RunDetailsResponse> => {
  const { response, message } = await getTransactionMessageByType<RunMessageInput>(MsgEndpoint.MSG_RUN, params)
  return {
    ...message,
    height: Number(response.height),
    deliverTx: response.tx_result,
  }
}
