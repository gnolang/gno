import { type WalletAPIAdapter } from '../../types'
import { AdenaWallet } from './adapters/adena'
import { supportedWallets } from './supported-wallets'

export class WalletService {
  static getAdapter(providerKey: string): WalletAPIAdapter {
    if (providerKey === supportedWallets.adena.key) {
      return new AdenaWallet()
    }

    throw new Error('Provider not supported')
  }
}
