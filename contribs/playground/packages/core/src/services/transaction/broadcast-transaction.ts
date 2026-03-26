import {
  JSONRPCProvider,
  TransactionEndpoint,
  type BroadcastTxCommitResult,
  type BroadcastTxSyncResult,
  type TM2Error,
} from '@gnolang/tm2-js-client'

import { getChainById } from '../../config'
import { invariant } from '../../utils'

export async function broadcastTransaction(
  transactionHex: string,
  networkId: string,
  rpcUrl?: string,
): Promise<BroadcastTxCommitResult> {
  if (!rpcUrl) {
    const chain = getChainById(networkId)
    invariant(chain, 'Invalid network')
    rpcUrl = chain.rpcUrl
  }

  const provider = new JSONRPCProvider(rpcUrl)

  try {
    return await provider.sendTransaction(transactionHex, TransactionEndpoint.BROADCAST_TX_COMMIT)
  } catch (error) {
    const log = (error as TM2Error).log
    if (log) throw new Error(log)

    throw error
  }
}

export async function broadcastTransactionSync(
  transactionHex: string,
  networkId: string,
  rpcUrl?: string,
): Promise<[BroadcastTxSyncResult, number]> {
  if (!rpcUrl) {
    const chain = getChainById(networkId)
    invariant(chain, 'Invalid network')
    rpcUrl = chain.rpcUrl
  }

  const provider = new JSONRPCProvider(rpcUrl)
  try {
    const blockNumber = await provider.getBlockNumber()
    const response = await provider.sendTransaction(transactionHex, TransactionEndpoint.BROADCAST_TX_SYNC)

    return [response, blockNumber] as const
  } catch (error) {
    const log = (error as TM2Error).log
    if (log) throw new Error(log)

    throw error
  }
}

export type { BroadcastTxSyncResult, BroadcastTxCommitResult }
