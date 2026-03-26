import type { AdenaWallet as AdenaWalletType, DoContractResponse, TransactionParams } from '@adena-wallet/sdk'

import type { SignTxResponse, WalletAPIAdapter, WalletNetwork } from '../../../types'

export class AdenaWallet implements WalletAPIAdapter {
  private get adena(): NonNullable<AdenaWalletType> {
    if (!window.adena) {
      throw new Error('Adena Wallet not installed')
    }

    return window.adena
  }

  async detect(): Promise<boolean> {
    if (document.readyState === 'complete') {
      return !!window.adena
    }

    return await new Promise((resolve) => {
      document.addEventListener('readystatechange', () => {
        resolve(!!window.adena)
      })
    })
  }

  getVersion() {
    // @ts-expect-error type missing
    return window.adena?.version as string
  }

  async connect(name = 'Gno IDE') {
    const response = await window.adena?.AddEstablish(name)

    if (response?.status !== 'success') {
      throw new Error(response?.message ?? 'Failed to connect')
    }
  }

  async getAccount() {
    const account = await window.adena?.GetAccount()
    if (!account) {
      throw new Error('Error: Failed to get account')
    }

    if (account.status !== 'success') {
      throw new Error(`Error: ${account.message}`)
    }

    return {
      name: 'Account 1',
      address: account.data.address,
      chain: account.data.chainId,
      pubKey: account.data.publicKey?.value ?? '',
      balance: account.data.coins,
    }
  }

  async getNetwork() {
    if (!window.adena?.GetNetwork) {
      throw new Error('Error: GetNetwork call not supported, please update Adena wallet')
    }

    const rsp = await window.adena?.GetNetwork()
    if (!rsp) {
      throw new Error('Error: failed to get network')
    }

    if (rsp.status !== 'success') {
      throw new Error(`Error: ${rsp.message}`)
    }

    return {
      id: rsp.data.chainId,
      name: rsp.data.networkName,
      rpcUrl: rsp.data.rpcUrl,
    }
  }

  async signAndBroadcastTransaction(transaction: unknown): Promise<DoContractResponse> {
    return await this.adena.DoContract(transaction as TransactionParams)
  }

  async signTransaction(transaction: unknown): Promise<SignTxResponse> {
    return await this.adena.SignTx(transaction as TransactionParams)
  }

  async switchNetwork(chainId: string) {
    const { status, message } = await this.adena.SwitchNetwork(chainId)
    if (status !== 'success') {
      throw new Error(message)
    }
  }

  async addNetwork(params: WalletNetwork) {
    const { status, type, message } = await this.adena.AddNetwork({
      chainId: params.id,
      chainName: params.name,
      rpcUrl: params.rpcUrl,
    })

    if (status !== 'success') {
      // Ignore if network already exists
      if ((type as string) === 'NETWORK_ALREADY_EXISTS') return

      throw new Error(message)
    }
  }

  onChangeNetwork(callback: (chainId: string) => void) {
    window.adena?.On('changedNetwork', callback)
  }

  onChangeAccount(callback: (address: string) => void) {
    window.adena?.On('changedAccount', callback)
  }
}
