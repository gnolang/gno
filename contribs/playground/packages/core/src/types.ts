import { type DeliverTx } from '@gnolang/tm2-js-client'

import type { AdenaSignTx, DoContractResponse } from '@adena-wallet/sdk'

import type { MemPackage } from './services/transaction/transaction-types'

export type SignTxResponse = Awaited<ReturnType<AdenaSignTx>>

export interface PlainFile {
  path: string
  content: string
}

export interface WorkbenchFile {
  path: string
  treeId: string
}

export interface WalletAccount {
  name: string
  address: string
  chain: string
  pubKey: string
  balance: string
}

export interface WalletNetwork {
  id: string
  name: string
  rpcUrl: string
}

export interface CallArgument {
  index: number
  value: string
}

/**
 * CallDetailsResponse contains original body of /vm.m_call transaction.
 */
export interface CallDetailsResponse {
  args: Array<{ index: number; value: string }>
  caller: string
  func: string
  height: number
  pkgPath: string
  send: string
  deliverTx?: DeliverTx
}

/**
 * RunDetailsResponse contains original body of /vm.m_run transaction.
 */
export interface RunDetailsResponse {
  send: string
  caller: string
  package: MemPackage
  deliverTx?: DeliverTx
  height: number
}

export interface WalletAPIAdapter {
  detect: () => Promise<boolean>
  getVersion: () => string | undefined
  connect: () => Promise<void>
  getAccount: () => Promise<WalletAccount>
  getNetwork: () => Promise<WalletNetwork>
  switchNetwork: (chain: string) => Promise<void>
  addNetwork: (params: WalletNetwork) => Promise<void>

  onChangeNetwork: (callback: (chain: string) => void) => void
  onChangeAccount: (callback: (address: string) => void) => void

  signAndBroadcastTransaction: (transaction: unknown) => Promise<DoContractResponse>
  signTransaction: (transaction: unknown) => Promise<SignTxResponse>
}
